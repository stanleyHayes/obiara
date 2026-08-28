// Command api is the composition root for the Obiara modular monolith
// (agent_plan.md §7.1). It loads configuration, connects infrastructure,
// builds hexagonal modules and wires inbound adapters. Health semantics
// (agent_plan.md §12): GET /live is process liveness only; GET /ready is
// dependency-aware and returns 503 while MongoDB is unreachable.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/stanleyHayes/obiara/internal/notifications"
	deliverystats "github.com/stanleyHayes/obiara/internal/notifications/deliverystats/adapters/outbound/mongodb"
	deliverystatsapp "github.com/stanleyHayes/obiara/internal/notifications/deliverystats/application"
	"github.com/stanleyHayes/obiara/internal/notifications/email"
	"github.com/stanleyHayes/obiara/internal/notifications/push"
	whatsappmongodb "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/adapters/outbound/mongodb"
	whatsappapp "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/application"
	whatsappdomain "github.com/stanleyHayes/obiara/internal/notifications/whatsapp/domain"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/privacy"
	"github.com/stanleyHayes/obiara/internal/safety"
	"github.com/stanleyHayes/obiara/services/api/internal/admin"
	adminemail "github.com/stanleyHayes/obiara/services/api/internal/admin/adapters/outbound/email"
	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics"
	"github.com/stanleyHayes/obiara/services/api/internal/calls"
	callsapp "github.com/stanleyHayes/obiara/services/api/internal/calls/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle"
	circledomain "github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
	circleroom "github.com/stanleyHayes/obiara/services/api/internal/circle/room"
	circleroomapp "github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog"
	catalogauthority "github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/adapters/outbound/adminauthority"
	commerceescrow "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger"
	ledgerauthority "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/adminauthority"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit"
	communityauditauthority "github.com/stanleyHayes/obiara/services/api/internal/communityaudit/adapters/outbound/adminauthority"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa"
	onboardingconsent "github.com/stanleyHayes/obiara/services/api/internal/consent"
	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap"
	consentdomain "github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship"
	courtshipproposal "github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal"
	"github.com/stanleyHayes/obiara/services/api/internal/fire"
	firemongodb "github.com/stanleyHayes/obiara/services/api/internal/fire/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/adapters/outbound/fireauthority"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ampe"
	"github.com/stanleyHayes/obiara/services/api/internal/games/anansesem"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition"
	"github.com/stanleyHayes/obiara/services/api/internal/games/ebe"
	owaresession "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session"
	"github.com/stanleyHayes/obiara/services/api/internal/identity"
	identitymongodb "github.com/stanleyHayes/obiara/services/api/internal/identity/adapters/outbound/mongodb"
	identityapplication "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack"
	"github.com/stanleyHayes/obiara/services/api/internal/member"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/delivery"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol"
	flagcontroldomain "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/health"
	apihttp "github.com/stanleyHayes/obiara/services/api/internal/platform/http"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/telemetry"
	"github.com/stanleyHayes/obiara/services/api/internal/profile"
	"github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit"
	livekitapp "github.com/stanleyHayes/obiara/services/api/internal/realtime/livekit/application"
	seedstage "github.com/stanleyHayes/obiara/services/api/internal/seed"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance"
	gardenmongodb "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/adapters/outbound/mongodb"
	gardenprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/adapters/outbound/privacy"
	gardenapp "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc"
	"github.com/stanleyHayes/obiara/services/api/internal/suban"
	"github.com/stanleyHayes/obiara/services/api/internal/trust"
	"github.com/stanleyHayes/obiara/services/api/internal/verification"
	adminverificationmongodb "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/adapters/outbound/mongodb"
	adminverificationprivacy "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/adapters/outbound/privacy"
	adminverificationapp "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness"
	"github.com/stanleyHayes/obiara/services/api/internal/waitlist"

	adminmongodb "github.com/stanleyHayes/obiara/services/api/internal/admin/adapters/outbound/mongodb"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "api startup failed:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	telemetryRuntime, err := telemetry.NewRuntime(ctx, os.Stdout, telemetry.RuntimeConfig{
		Service: "obiara-api", Version: cfg.ServiceVersion, Environment: cfg.Environment,
		Endpoint: cfg.TelemetryEndpoint, Insecure: cfg.TelemetryInsecure,
	})
	if err != nil {
		return fmt.Errorf("configure telemetry: %w", err)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		_ = telemetryRuntime.Shutdown(shutdownCtx)
	}()

	// Give the transport layer a sink for 5xx causes; error envelopes stay
	// opaque to callers, so without this a fault leaves no trace at all.
	apihttp.SetServerLogger(telemetryRuntime.Logger)

	connectCtx, cancel := context.WithTimeout(ctx, cfg.MongoConnectTimeout)
	defer cancel()
	client, err := apimongo.Connect(connectCtx, cfg.MongoURI)
	if err != nil {
		return err
	}
	waitlistStore := waitlist.NewStore(client.Database(cfg.MongoDatabase), time.Now)
	if err = waitlistStore.EnsureIndexes(connectCtx); err != nil {
		return fmt.Errorf("ensure waitlist indexes: %w", err)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	// Outbound delivery adapters are built before the modules that depend on
	// them, so a channel that cannot reach its provider stops the deploy
	// here rather than accepting messages and dropping them (agent_plan.md
	// §11). Configuration has already rejected simulators outside
	// development.
	whatsappSender, err := delivery.WhatsAppSender(cfg.Notifications)
	if err != nil {
		return err
	}
	whatsappLog := whatsappmongodb.NewDeliveryLog(client.Database(cfg.MongoDatabase))
	if err := whatsappLog.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure whatsapp delivery indexes: %w", err)
	}
	whatsappChannel := whatsappapp.NewChannelService(whatsappSender, whatsappLog, nil, time.Now)
	emailSender, err := delivery.EmailSender(cfg.Notifications)
	if err != nil {
		return err
	}
	// Transactional email (E13-S04): Resend channel with signed delivery
	// webhooks.
	emailModule, err := email.NewModule(ctx, client.Database(cfg.MongoDatabase), emailSender, os.Getenv("RESEND_WEBHOOK_SECRET"))
	if err != nil {
		return fmt.Errorf("build email module: %w", err)
	}
	// The OTP router needs the email service, not just the provider: a
	// member who verified an address receives their sign-in code through
	// the same logged, webhook-correlated path as every other message.
	otpSender, err := delivery.OtpSender(cfg.Notifications, whatsappChannel, emailModule.Email, telemetryRuntime.Logger)
	if err != nil {
		return err
	}
	// Ask the email provider whether it will accept us, before any operator
	// needs a sign-in code. Non-fatal: it logs the answer and carries on.
	delivery.PreflightEmail(ctx, emailSender, telemetryRuntime.Logger)

	pushSender, err := delivery.PushSender(cfg.Notifications)
	if err != nil {
		return err
	}
	pushModule, err := push.NewModule(ctx, client.Database(cfg.MongoDatabase), pushSender)
	if err != nil {
		return fmt.Errorf("build push module: %w", err)
	}
	// Courtship proposals (private two-party negotiation). Member references
	// and proposal details are keyed with the circle secret, which already
	// protects the adjacent private-room surfaces.
	proposalModule, err := courtshipproposal.NewModule(ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret)
	if err != nil {
		return fmt.Errorf("build courtship proposal module: %w", err)
	}
	// The courtship room mechanics: pace, pause, honesty, closure and the
	// in-room safety actions, all keyed with the same secret.
	courtshipRoomModule, err := courtship.NewRoomModule(ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret)
	if err != nil {
		return fmt.Errorf("build courtship room module: %w", err)
	}
	// The seed stage a member meets before a courtship room exists.
	seedStageModule, err := seedstage.NewStageModule(ctx, client.Database(cfg.MongoDatabase), cfg.SeedHMACSecret)
	if err != nil {
		return fmt.Errorf("build seed stage module: %w", err)
	}
	// The weekly seed allowance: server-authoritative and non-purchasable,
	// renewed on the Monday of the member's own week.
	allowanceModule, err := allowance.NewModule(ctx, client.Database(cfg.MongoDatabase),
		cfg.SeedHMACSecret, cfg.SeedWeekTimezone, cfg.SeedWeeklyAllowance)
	if err != nil {
		return fmt.Errorf("build seed allowance module: %w", err)
	}
	// The run sheet a host works through while running a fire. Authority
	// comes from the fire aggregate itself, which owns who hosts what.
	runSheetModule, err := runsheet.NewModule(ctx, client.Database(cfg.MongoDatabase),
		fireauthority.New(firemongodb.NewRepository(client.Database(cfg.MongoDatabase))),
		cfg.CircleHMACSecret)
	if err != nil {
		return fmt.Errorf("build run sheet module: %w", err)
	}

	// Modules are composed here at startup (agent_plan.md §7.2).
	memberModule, err := member.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build member module: %w", err)
	}
	// The identity module provides session issuance and phone OTP
	// registration (E03-S01).
	identityModule, err := identity.NewModule(ctx, client.Database(cfg.MongoDatabase), otpSender)
	if err != nil {
		return fmt.Errorf("build identity module: %w", err)
	}
	onboardingConsentModule, err := onboardingconsent.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build onboarding consent module: %w", err)
	}
	// Verification (E03-S03) promotes accounts through the identity tier
	// state machine via the composition-time bridge.
	identityProvider, err := delivery.IdentityProvider(cfg.Verification)
	if err != nil {
		return err
	}
	verificationModule, err := verification.NewModule(ctx, client.Database(cfg.MongoDatabase), identityProvider, tierBridge{tiers: identityModule.Tiers}, []byte(cfg.VerificationHMACSecret))
	if err != nil {
		return fmt.Errorf("build verification module: %w", err)
	}
	livenessProvider, err := delivery.LivenessProvider(cfg.Verification)
	if err != nil {
		return err
	}
	livenessModule, err := liveness.NewModule(ctx, client.Database(cfg.MongoDatabase), livenessProvider, cfg.LivenessHMACSecret)
	if err != nil {
		return fmt.Errorf("build liveness module: %w", err)
	}
	// Privacy (E03-S10) serves export/deletion requests and legal holds.
	privacyModule, err := privacy.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build privacy module: %w", err)
	}
	// Trust visibility composes the bounded S3-013 projection and S3-015
	// disclosure policy. Consent and non-owner endpoint disclosure remain
	// explicitly fail closed until persistence-backed adapters are available.
	trustModule, err := trust.NewModule(
		ctx,
		client.Database(cfg.MongoDatabase),
		ownerProjectionAuthorizer{},
		denyTrustConsent{},
		ownerEndpointAuthorizer{},
	)
	if err != nil {
		return fmt.Errorf("build trust module: %w", err)
	}
	// Profile doorway question and photo vault (E03-S09).
	profileModule, err := profile.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build profile module: %w", err)
	}
	// Listening eligibility (E06-S03) feeds the sow boundary.
	listeningModule, err := listening.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build listening module: %w", err)
	}
	gardenRepository := gardenmongodb.NewRepository(client.Database(cfg.MongoDatabase))
	if err := gardenRepository.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure seed garden indexes: %w", err)
	}
	gardenKeyer, err := gardenprivacy.NewKeyer([]byte(cfg.SeedHMACSecret))
	if err != nil {
		return fmt.Errorf("configure seed garden privacy: %w", err)
	}
	gardenService := gardenapp.NewService(gardenRepository, gardenKeyer, time.Now)
	circleModule, err := circle.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build circle module: %w", err)
	}
	gamePairs := circleGamePairResolver{circles: circleModule.Circles}
	owareModule, err := owaresession.NewModule(
		ctx,
		client.Database(cfg.MongoDatabase),
		cfg.CircleHMACSecret,
		gamePairs,
		gamePairs,
	)
	if err != nil {
		return fmt.Errorf("build oware module: %w", err)
	}
	anansesemModule, err := anansesem.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret, gamePairs,
	)
	if err != nil {
		return fmt.Errorf("build anansesem module: %w", err)
	}
	ampeModule, err := ampe.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret, gamePairs,
	)
	if err != nil {
		return fmt.Errorf("build ampe module: %w", err)
	}
	ebeModule, err := ebe.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret,
		gamePairs, composedReviewerAuthority{},
	)
	if err != nil {
		return fmt.Errorf("build ebe module: %w", err)
	}
	competitionModule, err := competition.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret,
		composedReviewerAuthority{},
	)
	if err != nil {
		return fmt.Errorf("build competition module: %w", err)
	}
	circleRoomModule, err := circleroom.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CircleHMACSecret,
		circleRoomAuthorizer{circles: circleModule.Circles},
	)
	if err != nil {
		return fmt.Errorf("build circle room module: %w", err)
	}
	// Fire scheduling and attendance (E09-S01).
	fireModule, err := fire.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build fire module: %w", err)
	}
	membershipModule, err := membership.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CommerceHMACSecret,
	)
	if err != nil {
		return fmt.Errorf("build membership module: %w", err)
	}
	matchmakerModule, err := matchmaker.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build matchmaker module: %w", err)
	}
	escrowModule, err := commerceescrow.NewModule(ctx, client.Database(cfg.MongoDatabase), []byte(cfg.CommerceHMACSecret))
	if err != nil {
		return fmt.Errorf("build escrow module: %w", err)
	}
	reconciliationModule, err := reconciliation.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build reconciliation module: %w", err)
	}
	// Embers (E06-S10). The mutual-ember doorway opener stays nil until the
	// sprout module composes.
	emberModule, err := ember.NewModule(ctx, client.Database(cfg.MongoDatabase), nil)
	if err != nil {
		return fmt.Errorf("build ember module: %w", err)
	}
	// Notification preferences and caps (E13-S01).
	notificationModule, err := notifications.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build notifications module: %w", err)
	}
	// Admin principals and MFA (E16-S01); codes ride the email channel.
	adminModule, err := admin.NewModule(ctx, client.Database(cfg.MongoDatabase), adminemail.NewSender(emailModule.Email, cfg.AdminConsoleURL))
	if err != nil {
		return fmt.Errorf("build admin module: %w", err)
	}
	// The commerce catalog: operators curate it, members read what is
	// published. Curation authority comes from the admin roles that already
	// carry commercial responsibility.
	catalogModule, err := catalog.NewModule(ctx, client.Database(cfg.MongoDatabase),
		catalogauthority.New(adminModule.Admin), cfg.CommerceHMACSecret)
	if err != nil {
		return fmt.Errorf("build catalog module: %w", err)
	}
	// The double-entry ledger behind the catalog: what the platform owes and
	// is owed. Finance desk only, in both directions.
	ledgerModule, err := ledger.NewModule(ctx, client.Database(cfg.MongoDatabase),
		ledgerauthority.New(adminModule.Admin), cfg.CommerceHMACSecret)
	if err != nil {
		return fmt.Errorf("build ledger module: %w", err)
	}
	// The community audit desk: conduct cases an operator reviews. Gated on
	// the trust-and-safety roles, with evidence access and decisions behind
	// the same MFA step-up the rest of the console uses.
	communityAuditModule, err := communityaudit.NewModule(ctx, client.Database(cfg.MongoDatabase),
		communityauditauthority.NewAuthority(adminModule.Admin),
		communityauditauthority.NewMFAGate(adminModule.Admin), cfg.AdminHMACSecret)
	if err != nil {
		return fmt.Errorf("build community audit module: %w", err)
	}

	adminSubjectKeyer, err := adminverificationprivacy.NewHMACKeyer([]byte(cfg.AdminHMACSecret))
	if err != nil {
		return fmt.Errorf("configure admin subject references: %w", err)
	}
	adminVerificationRepository := adminverificationmongodb.NewRepository(client.Database(cfg.MongoDatabase), adminSubjectKeyer)
	if err := adminVerificationRepository.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure admin verification indexes: %w", err)
	}
	adminVerificationService := adminverificationapp.NewService(adminVerificationRepository, time.Now)
	flagEnvironment := flagcontroldomain.EnvironmentStaging
	if strings.EqualFold(cfg.Environment, "production") {
		flagEnvironment = flagcontroldomain.EnvironmentProduction
	}
	flagControlModule, err := flagcontrol.NewModule(
		ctx, client.Database(cfg.MongoDatabase), adminModule.Admin,
		adminSubjectKeyer, flagEnvironment, os.Getenv,
	)
	if err != nil {
		return fmt.Errorf("build runtime flag controls: %w", err)
	}
	adminPrincipalResolver := func(r *http.Request) (adminverificationapp.Principal, error) {
		authorization := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") {
			return adminverificationapp.Principal{}, adminverificationapp.ErrForbidden
		}
		session, principal, authErr := adminModule.Admin.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")))
		if authErr != nil {
			return adminverificationapp.Principal{}, adminverificationapp.ErrForbidden
		}
		scopes := make([]adminverificationapp.Scope, 0, 3)
		if principal.HasRole(admindomain.RoleVerifier) || principal.HasRole(admindomain.RoleAdmin) {
			scopes = append(scopes,
				adminverificationapp.ScopeQueueRead,
				adminverificationapp.ScopeEvidenceRead,
				adminverificationapp.ScopeReview,
			)
		}
		if principal.HasRole(admindomain.RoleAdmin) {
			scopes = append(scopes, adminverificationapp.ScopeOperations)
		}
		if principal.HasRole(admindomain.RoleFinance) || principal.HasRole(admindomain.RoleAdmin) {
			scopes = append(scopes, adminverificationapp.ScopeFinance)
		}
		if principal.HasRole(admindomain.RoleTSAgent) || principal.HasRole(admindomain.RoleAdmin) {
			scopes = append(scopes, adminverificationapp.ScopeSafety)
		}
		return adminverificationapp.Principal{
			ActorID: principal.ID(), Scopes: scopes, MFAVerified: session.SteppedUp(),
		}, nil
	}
	// Suban character ledger (E15-S04): append-only events, recomputed marks.
	subanModule, err := suban.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build suban module: %w", err)
	}
	// Market-pack governance (E16-S06): four-eyes publishing with
	// configuration audit.
	marketPackModule, err := marketpack.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build market pack module: %w", err)
	}
	// Consent map (Doc 08 §8): purpose toggles with receipts.
	consentModule, err := consentmap.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build consent module: %w", err)
	}
	profileModule = profileModule.WithConsent(profileConsent{consents: consentModule.ConsentMap})
	// Scam-arc detection (E11-S11): rules-first signals with the action
	// ladder; case creation bridges to the safety context when wired.
	scamModule, err := scamarc.NewModule(ctx, client.Database(cfg.MongoDatabase), monitoringConsent{consents: consentModule.ConsentMap}, nil)
	if err != nil {
		return fmt.Errorf("build scamarc module: %w", err)
	}
	// Analytics pipeline and P0 funnel metrics (E15-S01/S02/S07), gated by
	// the consent map's product-analytics row.
	analyticsModule, err := analytics.NewModule(ctx, client.Database(cfg.MongoDatabase), consentGate{consents: consentModule.ConsentMap})
	if err != nil {
		return fmt.Errorf("build analytics module: %w", err)
	}
	// In-app calls (E09-S09): LiveKit tokens, no phone exposure. The
	// realtime adapter activates when LIVEKIT_API_KEY/LIVEKIT_API_SECRET are
	// configured; otherwise the call routes report not-configured cleanly.
	tokenIssuer := callsapp.TokenIssuer(unconfiguredLivekit{})
	if apiKey, apiSecret := os.Getenv("LIVEKIT_API_KEY"), os.Getenv("LIVEKIT_API_SECRET"); apiKey != "" && apiSecret != "" {
		adapter, adapterErr := livekit.New(livekit.Config{
			APIKey: apiKey, APISecret: apiSecret, MaxTTL: 30 * time.Minute, ClockSkew: 30 * time.Second,
		}, time.Now)
		if adapterErr != nil {
			return fmt.Errorf("configure livekit: %w", adapterErr)
		}
		tokenIssuer = adapter
	}
	callsModule, err := calls.NewModule(ctx, client.Database(cfg.MongoDatabase), tokenIssuer)
	if err != nil {
		return fmt.Errorf("build calls module: %w", err)
	}
	// Safety intake (E12-S01): reports ride the durable outbox to queue
	// processors.
	safetyOutbox := outbox.NewStore(client.Database(cfg.MongoDatabase), time.Now)
	if err := safetyOutbox.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure outbox indexes: %w", err)
	}
	enforcement := identityapplication.NewEnforcementService(identitymongodb.NewAccountRepository(client.Database(cfg.MongoDatabase)), time.Now)
	safetyModule, err := safety.NewModule(ctx, client.Database(cfg.MongoDatabase), safetyOutbox, enforcement, identityModule.Sessions)
	if err != nil {
		return fmt.Errorf("build safety module: %w", err)
	}
	// Nnoboa kin nominations (E13-S06): consent invites ride the WhatsApp
	// channel composed above.
	nnoboaModule, err := nnoboa.NewModule(ctx, client.Database(cfg.MongoDatabase), nnoboa.SenderFunc(
		func(ctx context.Context, msg whatsappdomain.Message) error {
			_, err := whatsappChannel.SendNnoboaConsent(
				ctx, msg.To(), msg.Params()["kin_name"],
				msg.Params()["nomination_id"], msg.Params()["consent_token"],
			)
			return err
		}), cfg.NnoboaInviteSecret)
	if err != nil {
		return fmt.Errorf("build nnoboa module: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /live", health.Live())
	mux.Handle("GET /ready", health.Ready(func(ctx context.Context) error {
		return client.Ping(ctx, readpref.Primary())
	}))
	apihttp.RegisterMemberRoutes(mux, memberModule.Register.Handle)
	apihttp.RegisterWaitlistRoutes(mux, waitlistStore, adminPrincipalResolver)
	apihttp.RegisterAuthRoutes(mux, identityModule.Registration, identityModule.Sessions)
	apihttp.RegisterPushRoutes(mux, pushModule.Push, identityModule.Sessions)
	apihttp.RegisterCourtshipProposalRoutes(mux, proposalModule.Proposals, identityModule.Sessions)
	apihttp.RegisterCourtshipRoomRoutes(mux, courtship.NewRoom(courtshipRoomModule), identityModule.Sessions)
	apihttp.RegisterSeedStageRoutes(mux, seedstage.NewStage(seedStageModule), identityModule.Sessions)
	apihttp.RegisterFireRunSheetRoutes(mux, runSheetModule.RunSheets, identityModule.Sessions)
	apihttp.RegisterCatalogRoutes(mux, catalogModule.Catalog, identityModule.Sessions)
	apihttp.RegisterSeedAllowanceRoutes(mux, allowanceModule.Allowances, identityModule.Sessions)
	apihttp.RegisterLedgerRoutes(mux, ledgerModule.Ledger)
	apihttp.RegisterCommunityAuditRoutes(mux, communityAuditModule.Audit)
	apihttp.RegisterOnboardingConsentRoutes(mux, onboardingConsentModule.Onboarding, identityModule.Sessions)
	apihttp.RegisterVerificationRoutes(mux, verificationModule.Verification, identityModule.Sessions)
	apihttp.RegisterLivenessRoutes(mux, livenessModule.Liveness, livenessModule.Artifacts, identityModule.Sessions)
	apihttp.RegisterPrivacyRoutes(mux, privacyModule.Privacy, identityModule.Sessions)
	apihttp.RegisterTrustVisibilityRoutes(mux, trustModule.Visibility, identityModule.Sessions)
	apihttp.RegisterDoorwayRoutes(mux, profileModule.Doorway, profileModule.Vault, identityModule.Sessions)
	apihttp.RegisterProfileRoutes(mux, profileModule.Profile, consentModule.ConsentMap, identityModule.Sessions)
	apihttp.RegisterListeningRoutes(mux, listeningModule.Listening, identityModule.Sessions)
	apihttp.RegisterGardenRoutes(mux, gardenService, identityModule.Sessions)
	apihttp.RegisterCircleRoutes(mux, circleModule.Circles, identityModule.Sessions)
	apihttp.RegisterCircleRoomRoutes(mux, circleRoomModule.Rooms, identityModule.Sessions)
	apihttp.RegisterOwareRoutes(mux, owareModule.Sessions, gamePairs, identityModule.Sessions)
	apihttp.RegisterAnansesemRoutes(mux, anansesemModule.Stories, gamePairs, identityModule.Sessions)
	apihttp.RegisterAmpeRoutes(mux, ampeModule.Rounds, ampeModule.Presence, gamePairs, identityModule.Sessions)
	apihttp.RegisterEbeRoutes(mux, ebeModule.Catalog, ebeModule.Duels, gamePairs, identityModule.Sessions, adminPrincipalResolver)
	apihttp.RegisterCompetitionRoutes(
		mux, competitionModule.Cohorts, competitionModule.Manager,
		competitionModule.Competitions, competitionModule.Competitions,
		competitionModule.Oware,
		identityModule.Sessions, adminPrincipalResolver,
	)
	apihttp.RegisterFireRoutes(mux, fireModule.Fires, identityModule.Sessions, identityModule.Tiers)
	apihttp.RegisterMembershipRoutes(mux, membershipModule.Membership, membershipModule.Keyer, identityModule.Sessions)
	apihttp.RegisterMatchmakerRoutes(mux, matchmakerModule.Engagements, membershipModule.Keyer, identityModule.Sessions)
	apihttp.RegisterEscrowRoutes(mux, escrowModule.Escrows, membershipModule.Keyer, identityModule.Sessions)
	apihttp.RegisterEmberRoutes(mux, emberModule.Embers, identityModule.Sessions)
	apihttp.RegisterNotificationRoutes(mux, notificationModule.Notifications, identityModule.Sessions)
	apihttp.RegisterSafetyRoutes(mux, safetyModule.Safety, identityModule.Sessions)
	apihttp.RegisterSubanRoutes(mux, subanModule.Suban, subanModule.Explanation, identityModule.Sessions)
	apihttp.RegisterAdminRoutes(mux, adminModule.Admin)
	// The operator inbox is a projection of queues that already exist, so it
	// takes their ports rather than a store of its own; only the
	// acknowledgement watermark is persisted.
	apihttp.RegisterAdminNotificationRoutes(
		mux,
		adminModule.Admin,
		adminVerificationService,
		adminmongodb.NewNotificationMarks(client.Database(cfg.MongoDatabase)),
		adminPrincipalResolver,
		time.Now,
	)
	apihttp.RegisterAdminMatchmakerRoutes(mux, matchmakerModule.Catalog, adminPrincipalResolver)
	apihttp.RegisterAdminEscrowRoutes(mux, escrowModule.Escrows, matchmakerModule.Engagements, adminPrincipalResolver)
	apihttp.RegisterAdminFinanceRoutes(mux, reconciliationModule.Queries, adminPrincipalResolver)
	apihttp.RegisterAdminVerificationRoutes(mux, adminVerificationService, adminPrincipalResolver)
	apihttp.RegisterAdminSafetyRoutes(mux, safetyModule.Cases, safetyModule.Evidence, membershipModule.Keyer, adminPrincipalResolver)
	apihttp.RegisterAdminCareRoutes(mux, safetyModule.Care, membershipModule.Keyer, adminPrincipalResolver)
	apihttp.RegisterAdminControlRoutes(mux, flagControlModule.Controls, flagControlModule.Repo, adminSubjectKeyer, adminPrincipalResolver)
	apihttp.RegisterAdminMemberRoutes(mux, identitymongodb.NewAccountRepository(client.Database(cfg.MongoDatabase)), adminSubjectKeyer, adminPrincipalResolver)
	apihttp.RegisterCallRoutes(mux, callsModule.Calls, identityModule.Sessions)
	apihttp.RegisterMetricsRoutes(mux, analyticsModule.Metrics, adminPrincipalResolver)
	apihttp.RegisterScamArcRoutes(mux, scamModule.ScamArc, adminPrincipalResolver)
	apihttp.RegisterDeliveryStatsRoutes(mux, deliverystatsapp.NewStatsService(deliverystats.NewStore(client.Database(cfg.MongoDatabase)), time.Now), adminPrincipalResolver)
	apihttp.RegisterConsentRoutes(mux, consentModule.ConsentMap, identityModule.Sessions)
	apihttp.RegisterMarketPackRoutes(mux, marketPackModule.Packs, adminPrincipalResolver)
	apihttp.RegisterNominationRoutes(mux, nnoboaModule.Nominations, identityModule.Sessions)
	apihttp.RegisterResendWebhookRoute(mux, emailModule.Webhook, inbox.NewStore(client.Database(cfg.MongoDatabase), time.Now))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apihttp.Correlation(telemetryRuntime.HTTP(apihttp.FeatureFlags(mux, flagControlModule.Flags), apihttp.CorrelationID)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.ListenAndServe() }()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

type ownerProjectionAuthorizer struct{}

type circleRoomAuthorizer struct {
	circles interface {
		Allows(context.Context, string, string, circledomain.Capability) (bool, error)
	}
}

type circleGamePairResolver struct {
	circles interface {
		Get(context.Context, string, string) (circledomain.Circle, error)
	}
}

func (resolver circleGamePairResolver) Pair(ctx context.Context, circleID, actorID string) (string, error) {
	current, err := resolver.circles.Get(ctx, strings.TrimSpace(circleID), strings.TrimSpace(actorID))
	if err != nil {
		return "", err
	}
	active := make([]string, 0, 2)
	actorActive := false
	for _, membership := range current.Memberships() {
		switch membership.State() {
		case circledomain.StateMember, circledomain.StateHost, circledomain.StateOwner:
			active = append(active, membership.MemberID())
			if membership.MemberID() == strings.TrimSpace(actorID) {
				actorActive = true
			}
		}
	}
	if len(active) != 2 || !actorActive {
		return "", errors.New("private game requires exactly two active circle members")
	}
	if active[0] == strings.TrimSpace(actorID) {
		return active[1], nil
	}
	return active[0], nil
}

func (resolver circleGamePairResolver) RequireParticipant(ctx context.Context, roomID, actorID string) error {
	_, err := resolver.Pair(ctx, roomID, actorID)
	return err
}

func (resolver circleGamePairResolver) Revalidate(ctx context.Context, roomID, firstID, secondID string) error {
	other, err := resolver.Pair(ctx, roomID, firstID)
	if err != nil || other != strings.TrimSpace(secondID) {
		return errors.New("private game participant pair is not current")
	}
	return nil
}

func (resolver circleGamePairResolver) RevalidateAuthors(ctx context.Context, roomID, firstID, secondID string) error {
	return resolver.Revalidate(ctx, roomID, firstID, secondID)
}

func (authorizer circleRoomAuthorizer) Authorize(ctx context.Context, decision circleroomapp.Decision) error {
	capability := circledomain.CapabilityView
	switch decision.Capability {
	case circleroomapp.CapabilityRead:
		capability = circledomain.CapabilityView
	case circleroomapp.CapabilityPost:
		capability = circledomain.CapabilityPost
	case circleroomapp.CapabilityHost:
		capability = circledomain.CapabilityManage
	default:
		return circleroomapp.ErrDenied
	}
	allowed, err := authorizer.circles.Allows(ctx, decision.CircleID, decision.ActorID, capability)
	if err != nil || !allowed {
		return circleroomapp.ErrDenied
	}
	return nil
}

func (ownerProjectionAuthorizer) CanProject(_ context.Context, requesterID, rootID string) (bool, error) {
	return requesterID != "" && requesterID == rootID, nil
}

type denyTrustConsent struct{}

func (denyTrustConsent) Allows(context.Context, string, string) (bool, error) {
	return false, nil
}

type ownerEndpointAuthorizer struct{}

func (ownerEndpointAuthorizer) CanReveal(_ context.Context, requesterID, endpointID string) (bool, error) {
	return requesterID != "" && requesterID == endpointID, nil
}

// consentGate bridges the analytics ConsentGate port to the consent map.
type consentGate struct {
	consents consentMapService
}

type profileConsentState interface {
	StateFor(context.Context, string, consentdomain.Purpose) (bool, error)
}

type profileConsent struct {
	consents profileConsentState
}

func (bridge profileConsent) Allows(ctx context.Context, memberID, consentRef string) (bool, error) {
	if consentRef != "cons_profile_visibility" {
		return false, nil
	}
	return bridge.consents.StateFor(ctx, memberID, consentdomain.PurposeProfileVisibility)
}

func (gate consentGate) AllowsAnalytics(ctx context.Context, memberID string) (bool, error) {
	return gate.consents.StateFor(ctx, memberID, consentdomain.PurposeProductAnalytics)
}

// monitoringConsent bridges the scam-arc MonitoringConsent port.
type monitoringConsent struct {
	consents consentMapService
}

func (bridge monitoringConsent) MonitoringAllowed(ctx context.Context, roomID string) (bool, error) {
	// Rooms are member-scoped in the consent map (per-room override arrives
	// with room-scoped consent records; member state applies meanwhile).
	return bridge.consents.StateFor(ctx, roomID, consentdomain.PurposeScamArc)
}

type consentMapService interface {
	StateFor(ctx context.Context, memberID string, purpose consentdomain.Purpose) (bool, error)
}

// unconfiguredLivekit reports cleanly when no LiveKit credentials exist
// (local/dev without the managed boundary).
type unconfiguredLivekit struct{}

func (unconfiguredLivekit) Issue(context.Context, livekitapp.JoinRequest) (livekitapp.JoinToken, error) {
	return livekitapp.JoinToken{}, errors.New("livekit is not configured")
}

// composedReviewerAuthority is intentionally reachable only behind the
// operations-scoped admin HTTP boundary. It rejects missing actor identities;
// the resolver performs the current session, role, and scope revalidation.
type composedReviewerAuthority struct{}

func (composedReviewerAuthority) RequireReviewer(_ context.Context, actorID string) error {
	if strings.TrimSpace(actorID) == "" {
		return errors.New("reviewer identity is required")
	}
	return nil
}

func (authority composedReviewerAuthority) RequireTournamentManager(ctx context.Context, actorID string) error {
	return authority.RequireReviewer(ctx, actorID)
}

// tierBridge adapts the verification context's provider-neutral tier port
// to the identity context's tier state machine. Cross-context calls happen
// only at the composition root (agent_plan.md §7.2).
type tierBridge struct {
	tiers identityapplication.TierService
}

func (bridge tierBridge) Transition(ctx context.Context, accountID string, target int, reason, actorID string) error {
	_, err := bridge.tiers.Transition(ctx, accountID, identitydomain.Tier(target), reason, actorID)
	return err
}
