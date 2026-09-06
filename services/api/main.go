// Command api is the composition root for the Obiara modular monolith
// (agent_plan.md §7.1). It loads configuration, connects infrastructure,
// builds hexagonal modules and wires inbound adapters. Health semantics
// (agent_plan.md §12): GET /live is process liveness only; GET /ready is
// dependency-aware and returns 503 while MongoDB is unreachable.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
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
	safetymongodb "github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	safetyapplication "github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/services/api/internal/admin"
	adminemail "github.com/stanleyHayes/obiara/services/api/internal/admin/adapters/outbound/email"
	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics"
	authzapplication "github.com/stanleyHayes/obiara/services/api/internal/authz/application"
	authzdomain "github.com/stanleyHayes/obiara/services/api/internal/authz/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/calls"
	callsapp "github.com/stanleyHayes/obiara/services/api/internal/calls/application"
	"github.com/stanleyHayes/obiara/services/api/internal/circle"
	circledomain "github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
	circleroom "github.com/stanleyHayes/obiara/services/api/internal/circle/room"
	circleroomapp "github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog"
	catalogauthority "github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/adapters/outbound/adminauthority"
	commerceescrow "github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger"
	ledgerauthority "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/adminauthority"
	ledgersystemauthority "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/adapters/outbound/systemauthority"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership"
	membershipprivacy "github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/adapters/outbound/paystack"
	momoapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
	purchasemongo "github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase/adapters/outbound/mongodb"
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
	"github.com/stanleyHayes/obiara/services/api/internal/introduction"
	introductionmongo "github.com/stanleyHayes/obiara/services/api/internal/introduction/adapters/outbound/mongodb"
	introductionretention "github.com/stanleyHayes/obiara/services/api/internal/introduction/retention"
	"github.com/stanleyHayes/obiara/services/api/internal/marketpack"
	"github.com/stanleyHayes/obiara/services/api/internal/media"
	mediamongo "github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/objectstore"
	"github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/sharingpolicy"
	mediaapplication "github.com/stanleyHayes/obiara/services/api/internal/media/application"
	"github.com/stanleyHayes/obiara/services/api/internal/member"
	membermongo "github.com/stanleyHayes/obiara/services/api/internal/member/adapters/outbound/mongodb"
	memberdomain "github.com/stanleyHayes/obiara/services/api/internal/member/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/organization"
	organizationapplication "github.com/stanleyHayes/obiara/services/api/internal/organization/application"
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
	safeguarding "github.com/stanleyHayes/obiara/services/api/internal/safeguarding"
	safeguardingapplication "github.com/stanleyHayes/obiara/services/api/internal/safeguarding/application"
	safeguardingdomain "github.com/stanleyHayes/obiara/services/api/internal/safeguarding/domain"
	safeguardingretention "github.com/stanleyHayes/obiara/services/api/internal/safeguarding/retention"
	seedstage "github.com/stanleyHayes/obiara/services/api/internal/seed"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance"
	allowanceapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/application"
	allowancedomain "github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
	declineapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/decline/application"
	gardenmongodb "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/adapters/outbound/mongodb"
	gardenprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/adapters/outbound/privacy"
	gardenapp "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening"
	listeningapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/listening/application"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod"
	podmongo "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/adapters/outbound/mongodb"
	podprivacy "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/adapters/outbound/privacy"
	podapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/application"
	poddomain "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/reviewdesk"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow"
	sowmedia "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/adapters/outbound/media"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sproutapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/application"
	"github.com/stanleyHayes/obiara/services/api/internal/sentinel/scamarc"
	"github.com/stanleyHayes/obiara/services/api/internal/suban"
	"github.com/stanleyHayes/obiara/services/api/internal/trust"
	"github.com/stanleyHayes/obiara/services/api/internal/verification"
	adminverificationmongodb "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/adapters/outbound/mongodb"
	adminverificationprivacy "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/adapters/outbound/privacy"
	adminverificationapp "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
	verificationapplication "github.com/stanleyHayes/obiara/services/api/internal/verification/application"
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
	circleModule, err := circle.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build circle module: %w", err)
	}

	seedStageModule, err := seedstage.NewStageModule(ctx, client.Database(cfg.MongoDatabase), cfg.SeedHMACSecret, circleModule.Repository)
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
	// The age gate is composed before verification because verification
	// cannot be built without it: a date of birth reaches this service there
	// and nowhere else, so that is the only place it can be checked before
	// being written down (M1-02).
	safeguardingModule, err := safeguarding.NewModule(ctx, client.Database(cfg.MongoDatabase), []byte(cfg.SafeguardingHMACSecret))
	if err != nil {
		return fmt.Errorf("build safeguarding module: %w", err)
	}
	// Assess purges inline once and gives up; this keeps retrying until the
	// data is actually gone. The interval is far inside the 24-hour SLA so a
	// failed attempt has many more before the deadline rather than one.
	go safeguardingretention.NewSweeper(
		safeguardingModule.Safeguarding, time.Now, slog.Default(),
	).Run(ctx, 15*time.Minute)

	verificationModule, err := verification.NewModule(ctx, client.Database(cfg.MongoDatabase), identityProvider, tierBridge{tiers: identityModule.Tiers}, safeguardingBridge{safeguarding: safeguardingModule.Safeguarding}, []byte(cfg.VerificationHMACSecret))
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

	// Listening is attached to the block check inside the media block, where
	// the asset reader exists. Without object storage there are no recordings
	// to hear, and the service refuses rather than listening blind to blocks.
	listeningBlockedListening := listeningModule.Listening

	gardenRepository := gardenmongodb.NewRepository(client.Database(cfg.MongoDatabase))
	if err := gardenRepository.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure seed garden indexes: %w", err)
	}
	gardenKeyer, err := gardenprivacy.NewKeyer([]byte(cfg.SeedHMACSecret))
	if err != nil {
		return fmt.Errorf("configure seed garden privacy: %w", err)
	}
	gardenService := gardenapp.NewService(gardenRepository, gardenKeyer, time.Now)
	// Safety is composed before the circle games, because pairing two people
	// into a private game is direct contact and has to honour a block.
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
	gamePairs := circleGamePairResolver{
		circles: circleModule.Circles,
		blocks:  sproutBlockBridge{safety: safetyModule.Safety},
	}
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
	// The verification ladder (FR-101). One gate, built once, asked by every
	// route that names a tier-gated action; the rules live in the authorization
	// kernel's grant table, not here.
	memberGate := apihttp.NewMemberGate(identityModule.Tiers)
	apihttp.RegisterMemberRoutes(mux, memberModule.Register.Handle)
	apihttp.RegisterWaitlistRoutes(mux, waitlistStore, adminPrincipalResolver)
	apihttp.RegisterAuthRoutes(mux, identityModule.Registration, identityModule.Sessions)
	apihttp.RegisterPushRoutes(mux, pushModule.Push, identityModule.Sessions)
	apihttp.RegisterCourtshipProposalRoutes(mux, proposalModule.Proposals, identityModule.Sessions, memberGate)
	apihttp.RegisterCourtshipRoomRoutes(mux,
		courtship.NewRoom(courtshipRoomModule).
			WithBlocks(sproutBlockBridge{safety: safetyModule.Safety}),
		identityModule.Sessions, memberGate)
	// Present only when the seed stage was given a circle reader; without one
	// there is nothing to resolve candidates from and the routes stay absent.
	if seedStageModule.Sources != nil {
		apihttp.RegisterSeedSourceRoutes(mux, seedStageModule.Sources, identityModule.Sessions, memberGate)
	}
	apihttp.RegisterFireRunSheetRoutes(mux, runSheetModule.RunSheets, identityModule.Sessions)
	apihttp.RegisterCatalogRoutes(mux, catalogModule.Catalog, identityModule.Sessions)
	apihttp.RegisterSeedAllowanceRoutes(mux, allowanceModule.Allowances, identityModule.Sessions)
	apihttp.RegisterLedgerRoutes(mux, ledgerModule.Ledger)
	apihttp.RegisterCommunityAuditRoutes(mux, communityAuditModule.Audit)
	// Voice of Introduction. Composed only when a bucket is configured: the
	// recording has nowhere to go without one, and a half-built surface that
	// accepts a member's two-minute take and drops it is worse than a surface
	// that is plainly absent.
	if cfg.ObjectStorage.Configured() {
		// The pod store and its keyer are built before media, because the
		// media policy has to be able to ask them whether a listener is a
		// recipient. The pod module below reaches the same collection; this
		// is the same repository, not a second one.
		podRepository := podmongo.NewRepository(client.Database(cfg.MongoDatabase))
		podKeyer, podKeyerErr := podprivacy.NewKeyer([]byte(cfg.SeedHMACSecret))
		if podKeyerErr != nil {
			return fmt.Errorf("build pod keyer: %w", podKeyerErr)
		}
		mediaModule, mediaErr := media.NewModule(
			ctx,
			client.Database(cfg.MongoDatabase),
			objectstore.Config{
				Endpoint:  cfg.ObjectStorage.Endpoint,
				Region:    cfg.ObjectStorage.Region,
				Bucket:    cfg.ObjectStorage.Bucket,
				AccessKey: cfg.ObjectStorage.AccessKey,
				SecretKey: cfg.ObjectStorage.SecretKey,
				PathStyle: cfg.ObjectStorage.PathStyle,
			},
			[]string{
				introduction.ConsentPurposeID,
				poddomain.PlaybackPurposeID,
				sowmedia.RecordingPurposeID,
			},
			// Who, other than the owner, may hear a recording. Without these
			// the only policy is owner-only, and in a product where people
			// meet through their voices nobody can hear anybody: a pod rests
			// at a house front and will not open, and the listen gate that
			// arms a sow can never be satisfied. See agent_plan.md §63.
			map[string]sharingpolicy.Entitlement{
				poddomain.PlaybackPurposeID: podRecipientEntitlement{
					pods: podRepository, keyer: podKeyer,
					blocks: sproutBlockBridge{safety: safetyModule.Safety},
					now:    time.Now,
				},
				introduction.ConsentPurposeID: voiceOfIntroductionEntitlement{
					blocks: sproutBlockBridge{safety: safetyModule.Safety},
				},
			},
		)
		if mediaErr != nil {
			return fmt.Errorf("build media module: %w", mediaErr)
		}
		// The Voice of Introduction, built before anything that reaches
		// toward a person: the listen gate resolves a target's recordings
		// through this store, so both the sprout path and the sow path need
		// it to exist first.
		introductionModule, introErr := introduction.NewModule(
			ctx,
			client.Database(cfg.MongoDatabase),
			onboardingConsentModule.Consents,
			mediaModule.Access,
			mediaModule.Assets,
			// The eraser, not the asset row repository. The row repository's
			// Delete marks the row and leaves the audio in the bucket, so
			// withdrawing a recording used to keep the recording.
			media.NewEraser(mediaModule.Assets, mediaModule.Objects),
			introductionLadderBridge{tiers: identityModule.Tiers, log: slog.Default()},
			cfg.LivenessHMACSecret,
		)
		if introErr != nil {
			return fmt.Errorf("build introduction module: %w", introErr)
		}
		// The pod is the last step: a released sow is delivered by being
		// placed at the recipient's house front. Its eligibility check is
		// the block rule applied everywhere else two members meet.
		podModule, podErr := pod.NewModule(
			ctx,
			client.Database(cfg.MongoDatabase),
			podAuthorizerBridge{tiers: identityModule.Tiers},
			podEligibilityBridge{
				pods: podRepository,
				blocks: listeningBlockBridge{
					assets: mediaModule.Assets, safety: safetyModule.Safety,
				},
			},
			podIssuerBridge{access: mediaModule.Access},
			cfg.SeedHMACSecret,
		)
		if podErr != nil {
			return fmt.Errorf("build pod module: %w", podErr)
		}
		// Screening and the sow. Every sow is read by a person before it is
		// delivered, so screening's only outcome here is the review queue —
		// see agent_plan.md §49. The locale is left unset: nothing has been
		// language-reviewed yet, and an unreviewed language routes to a
		// person rather than refusing the sow, so "undetermined" is both
		// truthful and safe.
		screeningModule, screeningErr := screening.NewModule(
			ctx, client.Database(cfg.MongoDatabase), mediaModule.Assets, "", nil,
		)
		if screeningErr != nil {
			return fmt.Errorf("build screening module: %w", screeningErr)
		}
		sowModule, sowErr := sow.NewModule(
			ctx,
			client.Database(cfg.MongoDatabase),
			screeningModule.Screening,
			// Owned by the sower, and actually in the bucket. A sow carrying
			// a recording that never finished uploading would be delivered
			// as a pod that plays nothing.
			sowmedia.NewOwnership(mediaModule.Assets).
				WithArrival(media.NewEraser(mediaModule.Assets, mediaModule.Objects)),
			// The same three rules the sprout path applies. A sow is that
			// gesture carrying words, so it answers to them too.
			sproutListenBridge{
				introductions: introductionModule.Store,
				listening:     listeningModule.Listening,
			},
			sproutBlockBridge{safety: safetyModule.Safety},
			sproutDeclineBridge{declines: seedStageModule.Decline, now: time.Now},
			// What makes a sow arrive: a delivered sow is placed as a pod at
			// the recipient's house front.
			sowDeliveryBridge{pods: podModule.Pods},
			cfg.SeedHMACSecret,
			cfg.SeedWeeklyAllowance,
		)
		if sowErr != nil {
			return fmt.Errorf("build sow module: %w", sowErr)
		}
		apihttp.RegisterPodRoutes(mux, podModule.Pods, identityModule.Sessions, memberGate)

		apihttp.RegisterSowRoutes(mux, sowModule.Sows, identityModule.Sessions, memberGate)
		// Where a sow's recording is made. The Voice of Introduction path
		// cannot serve: it answers one of three fixed questions and is
		// offered to anyone who may hear the member.
		apihttp.RegisterSowRecordingRoutes(
			mux,
			sowmedia.NewRecorder(mediaModule.Access, mediaModule.Assets, sowAssetIDs{}, time.Now),
			identityModule.Sessions,
			memberGate,
		)
		// The desk settles the sow first and records the judgement second,
		// so a failure between them leaves a review a reviewer sees again
		// rather than a sow held forever with a seed inside it.
		apihttp.RegisterAdminScreeningRoutes(
			mux,
			screeningModule.Reviews,
			reviewdesk.New(sowModule.Sows, screeningModule.Reviews),
			adminPrincipalResolver,
		)

		// Hearing somebody else. Separate from the routes above, which are a
		// member's own recording and its withdrawal.
		apihttp.RegisterMemberVoiceRoutes(
			mux,
			introductionModule.Store,
			introductionModule.Playback,
			identityModule.Sessions,
			memberGate,
		)
		apihttp.RegisterIntroductionRoutes(
			mux,
			introductionModule.Introductions,
			introductionModule.Store,
			introductionModule.Playback,
			identityModule.Sessions,
			memberGate,
		)

		// Erasure runs here rather than in the worker because the aggregate
		// records it: MarkPurged appends the audit event that makes a
		// withdrawal provable, and a sweep writing to Mongo directly would
		// remove the bytes and lose the proof.
		go introductionretention.NewSweeper(
			introductionModule.Store,
			media.NewEraser(mediaModule.Assets, mediaModule.Objects),
			introductionModule.Introductions,
			time.Now,
			slog.Default(),
		).Run(ctx, time.Hour)

		// The safety context is composed after the seed stage, so the block check
		// is attached here. Until it is, the introduction visibility refuses to
		// offer anybody rather than offering people blind to blocks.
		seedStageModule = seedStageModule.WithBlocks(sproutBlockBridge{safety: safetyModule.Safety})
		listeningBlockedListening = listeningModule.Listening.WithBlocks(listeningBlockBridge{
			assets: mediaModule.Assets, safety: safetyModule.Safety,
		})
		seedStageModule.Sprout = seedStageModule.Sprout.WithListenGate(sproutListenBridge{
			introductions: introductionModule.Store,
			listening:     listeningModule.Listening,
		})
	}
	// The allowance does not depend on object storage, so it is attached
	// whether or not recordings are configured.
	seedStageModule.Sprout = seedStageModule.Sprout.
		WithAllowance(sproutAllowanceBridge{allowances: allowanceModule.Allowances}).
		WithDeclineLock(sproutDeclineBridge{declines: seedStageModule.Decline, now: time.Now}).
		WithBlockList(sproutBlockBridge{safety: safetyModule.Safety})
	// Registered after the gate is attached. Without object storage there are
	// no recordings, so nobody can have heard anyone and the sprout service
	// reports itself unavailable rather than accepting an unarmed sow.
	apihttp.RegisterSeedStageRoutes(mux, seedstage.NewStage(seedStageModule), identityModule.Sessions, memberGate)
	apihttp.RegisterOnboardingConsentRoutes(mux, onboardingConsentModule.Onboarding, identityModule.Sessions)
	apihttp.RegisterOnboardingStatusRoutes(
		mux,
		onboardingConsentModule.Consents,
		verificationModule.Verification,
		livenessModule.Liveness,
		identityModule.Sessions,
		identityModule.Tiers,
	)
	apihttp.RegisterVerificationRoutes(mux, verificationModule.Verification, identityModule.Sessions)
	apihttp.RegisterLivenessRoutes(mux, livenessModule.Liveness, livenessModule.Artifacts, identityModule.Sessions)
	apihttp.RegisterPrivacyRoutes(mux, privacyModule.Privacy, identityModule.Sessions)
	apihttp.RegisterTrustVisibilityRoutes(mux, trustModule.Visibility, identityModule.Sessions)
	apihttp.RegisterDoorwayRoutes(mux, profileModule.Doorway, profileModule.Vault, identityModule.Sessions)
	apihttp.RegisterProfileRoutes(mux, profileModule.Profile, consentModule.ConsentMap, identityModule.Sessions)
	apihttp.RegisterListeningRoutes(mux, listeningBlockedListening, identityModule.Sessions, memberGate)
	apihttp.RegisterGardenRoutes(mux, gardenService, identityModule.Sessions)
	apihttp.RegisterCircleRoutes(mux, circleModule.Circles, identityModule.Sessions, memberGate)
	apihttp.RegisterCircleRoomRoutes(mux, circleRoomModule.Rooms, identityModule.Sessions, memberGate)
	apihttp.RegisterOwareRoutes(mux, owareModule.Sessions, gamePairs, identityModule.Sessions, memberGate)
	apihttp.RegisterAnansesemRoutes(mux, anansesemModule.Stories, gamePairs, identityModule.Sessions, memberGate)
	apihttp.RegisterAmpeRoutes(mux, ampeModule.Rounds, ampeModule.Presence, gamePairs, identityModule.Sessions, memberGate)
	apihttp.RegisterEbeRoutes(mux, ebeModule.Catalog, ebeModule.Duels, gamePairs, identityModule.Sessions, adminPrincipalResolver, memberGate)
	apihttp.RegisterCompetitionRoutes(
		mux, competitionModule.Cohorts, competitionModule.Manager,
		competitionModule.Competitions, competitionModule.Competitions,
		competitionModule.Oware,
		identityModule.Sessions, adminPrincipalResolver, memberGate,
	)
	apihttp.RegisterFireRoutes(mux, fireModule.Fires, identityModule.Sessions, identityModule.Tiers, memberGate)
	// Organizations: the bodies a discount code is issued for. An operator
	// surface, because codes are issued by staff on their behalf — see
	// agent_plan.md §41.
	organizationModule, err := organization.NewModule(
		ctx, client.Database(cfg.MongoDatabase), cfg.CommerceHMACSecret,
	)
	if err != nil {
		return fmt.Errorf("build organization module: %w", err)
	}
	apihttp.RegisterAdminOrganizationRoutes(
		mux, organizationModule.Organizations, adminPrincipalResolver)

	// Discount codes, issued in an organization's name. The issuer check is
	// the organization context answering one question and nothing else, so
	// the promotion context never learns anything more about a body than
	// whether it is still a live relationship.
	promotionModule, err := promotion.NewModule(
		ctx, client.Database(cfg.MongoDatabase),
		organizationIssuerBridge{organizations: organizationModule.Organizations},
		membershipModule.Keyer,
	)
	if err != nil {
		return fmt.Errorf("build promotion module: %w", err)
	}
	apihttp.RegisterAdminPromotionRoutes(mux, promotionModule.Promotions, adminPrincipalResolver)

	// The referral scheme. Composed only when a commission and a withholding
	// rate are both set: a scheme that accrues but can never legally pay out
	// is a liability that only grows, so an unset rate leaves it absent.
	//
	// Affiliates are outside parties. Members are never affiliates
	// (agent_plan.md §41), which is what the member check below enforces.
	var affiliates apihttp.AdminAffiliates
	var affiliatePayouts apihttp.AdminPayouts
	if cfg.Paystack.Configured() && cfg.Affiliates.Configured() {
		transfers, transferErr := paystack.New(paystack.Config{
			BaseURL:   cfg.Paystack.BaseURL,
			SecretKey: cfg.Paystack.SecretKey,
		})
		if transferErr != nil {
			return fmt.Errorf("build affiliate transfer rail: %w", transferErr)
		}
		affiliateModule, affiliateErr := affiliate.NewModule(
			ctx, client.Database(cfg.MongoDatabase),
			conversionBridge{
				tiers: identityModule.Tiers,
				// The same case collection the safety desk reads. Asked one
				// bool and nothing else: whether a report against this member
				// was upheld.
				safety: safetymongodb.NewCaseRepository(client.Database(cfg.MongoDatabase)),
			},
			transfers,
			membershipModule.Keyer,
			affiliate.Settings{
				CommissionPesewas:      cfg.Affiliates.CommissionPesewas,
				MinimumPayoutPesewas:   cfg.Affiliates.MinimumPayoutPesewas,
				WithholdingBasisPoints: cfg.Affiliates.WithholdingBasisPoints,
			},
		)
		if affiliateErr != nil {
			return fmt.Errorf("build affiliate module: %w", affiliateErr)
		}
		affiliates, affiliatePayouts = affiliateModule.Affiliates, affiliateModule.Payouts
		apihttp.RegisterAdminAffiliateRoutes(
			mux, affiliates, affiliatePayouts,
			memberLookupBridge{members: memberModule.Members}.IsMember,
			organizationModule.Keyer,
			adminPrincipalResolver,
		)
	}

	// Buying a membership. Composed only when Paystack is configured: without
	// a secret key there is no way to take money and no way to verify a
	// webhook, and a purchase route that always failed would be worse than a
	// surface that is plainly absent — the same rule the Voice of
	// Introduction follows about object storage.
	//
	// Until this existed, membership.Service.Grant had no callers anywhere.
	// No pass could be created, so nothing in the product could be bought
	// (agent_plan.md §72).
	var purchases apihttp.Purchases
	if cfg.Paystack.Configured() {
		provider, providerErr := paystack.New(paystack.Config{
			BaseURL:     cfg.Paystack.BaseURL,
			SecretKey:   cfg.Paystack.SecretKey,
			CallbackURL: cfg.Paystack.CallbackURL,
		})
		if providerErr != nil {
			return fmt.Errorf("build paystack provider: %w", providerErr)
		}
		momoModule, momoErr := momo.NewModule(
			ctx, client.Database(cfg.MongoDatabase), provider, cfg.CommerceHMACSecret)
		if momoErr != nil {
			return fmt.Errorf("build collection module: %w", momoErr)
		}
		// Settlement posts through a system authority rather than the admin
		// one: money that arrives on a webhook has no operator standing
		// behind it, and it must still be booked.
		settlementLedger, ledgerErr := ledger.NewModule(ctx, client.Database(cfg.MongoDatabase),
			ledgersystemauthority.New(), cfg.CommerceHMACSecret)
		if ledgerErr != nil {
			return fmt.Errorf("build settlement ledger: %w", ledgerErr)
		}
		orders := purchasemongo.NewOrders(client.Database(cfg.MongoDatabase))
		if orderErr := orders.EnsureIndexes(ctx); orderErr != nil {
			return fmt.Errorf("ensure purchase order indexes: %w", orderErr)
		}
		purchases = purchase.New(
			catalogModule.Catalog,
			momoModule.Intents,
			membershipModule.Membership,
			purchaseKeyer{
				members: membershipModule.Keyer,
				secret:  []byte(cfg.CommerceHMACSecret),
			},
			purchase.NewSaleBook(settlementLedger.Ledger, ledgersystemauthority.Actor),
			time.Now,
		).WithDiscounts(promotionModule.Promotions).
			WithOrders(orders).
			WithMembers(memberReceiptBridge{members: memberModule.Members})
		// The webhook secret is the Paystack secret key: it is what Paystack
		// signs with, so it is what verification needs.
		apihttp.RegisterPurchaseRoutes(
			mux, purchases, identityModule.Sessions, cfg.Paystack.SecretKey)
		if !cfg.Paystack.Live() {
			slog.Default().Warn(
				"paystack is configured with test keys; no real money will move",
			)
		}
	}

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
	apihttp.RegisterAdminSafetyRoutes(mux, safetyModule.Cases, safetyModule.Evidence, safetyModule.Actions, membershipModule.Keyer, adminPrincipalResolver)
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

// circleGamePairResolver names the other member of a two-person circle.
//
// Every private circle game goes through it — Ampe, Oware, Anansesem, the
// competition — and so does every revalidation of a game already in progress.
// It is the one place that says "these two are playing together", which makes
// it the one place a block has to be honoured.
//
// Pairing is direct contact: it is the product putting two people together,
// not two people happening to share a room. That distinction is why this
// refuses across a block while circle membership itself does not — a block
// should not eject somebody from a community they belong to, and it must not
// let the product introduce them to the person they blocked.
type circleGamePairResolver struct {
	circles interface {
		Get(context.Context, string, string) (circledomain.Circle, error)
	}
	blocks sproutBlockBridge
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
		return "", errNoPairing
	}
	other := active[0]
	if other == strings.TrimSpace(actorID) {
		other = active[1]
	}
	// Refused in the same words as a circle that is not a pair, because the
	// difference is the rejection signal a block exists to withhold. A
	// missing check refuses too: not knowing whether these two have blocked
	// each other is not permission to pair them.
	if resolver.blocks.safety == nil {
		return "", errNoPairing
	}
	blocked, err := resolver.blocks.Blocked(ctx, strings.TrimSpace(actorID), other)
	if err != nil || blocked {
		return "", errNoPairing
	}
	return other, nil
}

// errNoPairing reads exactly like the refusal for a circle that is not a pair,
// on purpose.
var errNoPairing = errors.New("private game requires exactly two active circle members")

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

// introductionLadderBridge promotes a member to the sowing rung once their
// Voice of Introduction is complete. Cross-context calls happen only at the
// composition root (agent_plan.md §7.2).
type introductionLadderBridge struct {
	tiers identityapplication.TierService
	log   *slog.Logger
}

// SowingEarned reads the rung before writing it rather than treating every
// rejected transition as "already there". Both an account that is already on
// the sowing rung and one trying to skip a rung are refused by the same
// error, and only the first is success — swallowing both would hide a
// Tier-0 account reaching a surface it should never have reached.
func (bridge introductionLadderBridge) SowingEarned(ctx context.Context, memberID string) error {
	current, err := bridge.tiers.Tier(ctx, memberID)
	if err != nil {
		bridge.log.WarnContext(ctx, "sowing promotion could not read the rung",
			slog.String("reason", err.Error()))
		return err
	}
	if current >= identitydomain.TierSowing {
		return nil
	}
	if _, err := bridge.tiers.Transition(ctx, memberID, identitydomain.TierSowing,
		"voice of introduction complete", "introduction"); err != nil {
		// Not fatal to the recording, which is already stored. It is retried
		// the next time any recording is confirmed, but it must be visible
		// while it is outstanding.
		bridge.log.WarnContext(ctx, "sowing promotion deferred",
			slog.String("reason", err.Error()))
		return err
	}
	return nil
}

// podAuthorizerBridge carries the pod's authorization to the authz kernel.
//
// The pod asks in its own vocabulary — an action and a resource id — and the
// kernel answers on tier. This is the same gate every other member surface
// goes through; the pod simply asks it through a port of its own shape.
type podAuthorizerBridge struct {
	tiers identityapplication.TierService
}

func (bridge podAuthorizerBridge) Require(ctx context.Context, actorID, action, _ string) error {
	tier, err := bridge.tiers.Tier(ctx, actorID)
	if err != nil {
		return err
	}
	return authzapplication.NewAuthorizer().Require(
		authzdomain.Subject{MemberID: actorID, Tier: authzdomain.Tier(tier)},
		action, authzdomain.Resource{Type: "pod"},
	)
}

// podIssuerBridge mints the short-lived grant that lets a recipient hear what
// is inside a pod.
//
// The grant names the listener. The media context authorizes reads per
// subject, and a token issued without one would have to be issued as somebody
// else — which is how a link that leaks becomes a link that works for anyone.
type podIssuerBridge struct {
	access mediaapplication.AccessService
}

func (bridge podIssuerBridge) Issue(ctx context.Context, listenerID, mediaRef, _ string, ttl time.Duration) (string, error) {
	access, err := bridge.access.RequestRead(ctx, mediaapplication.ReadRequest{
		SubjectID: listenerID,
		AssetID:   mediaRef,
		Purpose:   poddomain.PlaybackPurposeID,
		TTL:       ttl,
	})
	if err != nil {
		return "", err
	}
	return access.URL, nil
}

// podEligibilityBridge answers what can change after a pod was created and
// still ought to stop it being opened.
//
// The aggregate already refuses a non-recipient, an inactive pod, an expired
// one, a stale revision and a replayed command. What it cannot know is
// whether the two people have since blocked each other — which is exactly the
// rule now applied everywhere else two members come into contact, so it is
// the rule here too rather than a new consent purpose invented for the
// occasion. See agent_plan.md §59.
//
// The pod's owner is keyed, as a person should be, so this does not ask the
// pod who sent it. It asks the recording: the asset knows its own owner, and
// that owner is the sender. It is the same resolution the listening gate
// already does, for the same reason.
type podEligibilityBridge struct {
	pods   *podmongo.Repository
	blocks listeningBlockBridge
}

func (bridge podEligibilityBridge) Revalidate(ctx context.Context, actorID, podID string) error {
	pod, err := bridge.pods.Find(ctx, podID)
	if err != nil {
		return err
	}
	blocked, err := bridge.blocks.Blocked(ctx, actorID, pod.MediaRef())
	if err != nil {
		return err
	}
	if blocked {
		return errPodNotAvailable
	}
	return nil
}

// errPodNotAvailable says the outcome and not the reason, like every other
// refusal that could otherwise reveal a block.
var errPodNotAvailable = errors.New("this pod is not available")

// listeningBlockBridge answers whether a listener may hear a recording at
// all, by resolving whose recording it is and asking the same block question
// everything else asks.
//
// Listening is what arms a sow, so a listening surface blind to blocks would
// let somebody accumulate the right to reach a person who had already said
// they wanted nothing to do with them.
type listeningBlockBridge struct {
	assets *mediamongo.AssetRepository
	safety safetyapplication.SafetyService
}

func (bridge listeningBlockBridge) Blocked(ctx context.Context, listenerID, assetID string) (bool, error) {
	asset, err := bridge.assets.FindByID(ctx, assetID)
	if err != nil {
		// A recording nothing can account for is not one anybody listens to.
		return true, nil
	}
	if asset.OwnerID() == listenerID {
		// Hearing your own recording back is not contact with anybody.
		return false, nil
	}
	return sproutBlockBridge{safety: bridge.safety}.Blocked(ctx, listenerID, asset.OwnerID())
}

// sproutBlockBridge answers the most basic question either member can have
// settled about the other: has one of them blocked the other?
//
// Both directions, because a block is a decision to be apart rather than a
// one-way filter the blocker can step around. Until this existed,
// SafetyService.IsBlocked had no callers anywhere: members could block each
// other and nothing in the product honoured it.
type sproutBlockBridge struct {
	safety blockReader
}

// blockReader is the one question this file asks the safety context.
type blockReader interface {
	IsBlocked(ctx context.Context, memberID, otherID string) (bool, error)
}

func (bridge sproutBlockBridge) Blocked(ctx context.Context, memberID, otherID string) (bool, error) {
	blocked, err := bridge.safety.IsBlocked(ctx, memberID, otherID)
	if err != nil || blocked {
		return blocked, err
	}
	return bridge.safety.IsBlocked(ctx, otherID, memberID)
}

// sproutDeclineBridge answers M4-AC-01 for the seed stage: is the target
// still shielded from this member by a decline?
//
// The seed stage keys participants under its own namespace, so it cannot
// compare keys with the decline context directly. Passing raw ids and letting
// the decline service key them its own way is what makes every decline
// already on record enforceable.
type sproutDeclineBridge struct {
	declines declineapplication.Service
	now      func() time.Time
}

func (bridge sproutDeclineBridge) Locked(ctx context.Context, sowerID, targetID string) (bool, error) {
	return bridge.declines.Locked(ctx, sowerID, targetID, bridge.now())
}

// sproutAllowanceBridge spends one seed for a sow (FR-201a).
type sproutAllowanceBridge struct {
	allowances *allowanceapplication.Service
}

// Spend opens a ledger for a first-time sower before charging it.
//
// The two calls carry different command ids on purpose. A ledger records each
// command id alongside a fingerprint of what that command did, so reusing the
// sow's id to open the ledger would make the spend that follows look like the
// same command with different input — and every first sow would be refused as
// a conflict rather than charged.
func (bridge sproutAllowanceBridge) Spend(ctx context.Context, memberID, commandID string) error {
	if _, err := bridge.allowances.CurrentOrIssue(ctx, memberID, "sow-open:"+commandID); err != nil {
		return err
	}
	if _, err := bridge.allowances.Spend(ctx, memberID, "sow:"+commandID, 1); err != nil {
		if errors.Is(err, allowancedomain.ErrInsufficient) {
			return sproutapplication.ErrNoSeeds
		}
		return err
	}
	return nil
}

// sproutListenBridge answers FR-202 for the seed stage: has this member heard
// enough of that member's Voice of Introduction to reach toward them?
//
// It resolves the target's recordings itself rather than taking an asset id
// from the caller, so the gate cannot be satisfied with a recording that
// belongs to somebody else. A member who has recorded nothing has no assets
// here, so nobody can sow toward them — which is the same rule read from the
// other side: people meet you through your voice.
type sproutListenBridge struct {
	introductions *introductionmongo.Store
	listening     listeningapplication.ListeningService
}

func (bridge sproutListenBridge) Heard(ctx context.Context, listenerID, targetID string) (bool, error) {
	assets, err := bridge.introductions.AssetIDsByOwner(ctx, targetID)
	if err != nil {
		return false, err
	}
	for _, asset := range assets {
		eligible, _, err := bridge.listening.Eligibility(ctx, listenerID, asset)
		if err != nil {
			return false, err
		}
		if eligible {
			return true, nil
		}
	}
	return false, nil
}

// safeguardingBridge carries verification's narrow age-gate port to the
// safeguarding context's assessment. Cross-context calls happen only at the
// composition root (agent_plan.md §7.2).
type safeguardingBridge struct {
	safeguarding safeguardingapplication.Service
}

// Assess refuses on every error, not only on ErrUnder18. An assessment that
// could not be completed is not an assessment that passed, and the two are
// distinguished here only so the member gets an honest message: one says they
// may not join, the other says to try again.
// MinimumAge reports the safeguarding domain's threshold, so the case records
// the rule it was actually decided under rather than a constant duplicated in
// the verification context.
func (bridge safeguardingBridge) MinimumAge() int { return safeguardingdomain.MinimumAge }

func (bridge safeguardingBridge) Assess(ctx context.Context, commandID, subjectID, sourceRef string, dateOfBirth time.Time) error {
	_, err := bridge.safeguarding.Assess(ctx, safeguardingapplication.Assessment{
		CommandID:   commandID,
		SubjectID:   subjectID,
		SourceRef:   sourceRef,
		DateOfBirth: dateOfBirth,
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, safeguardingapplication.ErrUnder18) {
		return verificationapplication.ErrBelowMinimumAge
	}
	return verificationapplication.ErrAgeGateUnavailable
}

// podRecipientEntitlement lets somebody hear what was left for them.
//
// A pod's recipients are keyed, as people should be, so this keys the
// listener the same way rather than asking the pod to name anybody. The
// question it answers is narrow on purpose: not "was this ever sent to you"
// but "is it resting for you now" — a pod that was taken back or has closed
// is not at anybody's house front, and a grant minted for one would let
// somebody hear a recording after the moment for hearing it had passed.
//
// The block check is here for the same reason it is on the open path: two
// people who have decided to be apart do not hear each other, and a media
// grant that ignored that would be a way around every other place the rule is
// applied.
type podRecipientEntitlement struct {
	pods   restingPods
	keyer  memberKeyer
	blocks sproutBlockBridge
	now    func() time.Time
}

// restingPods and memberKeyer are the two things this entitlement needs.
type restingPods interface {
	HoldsFor(ctx context.Context, recipientKey, mediaRef string, at time.Time) (bool, error)
}

type memberKeyer interface {
	Key(namespace, value string) (string, error)
}

func (e podRecipientEntitlement) MayHear(
	ctx context.Context, listenerID, ownerID, assetID string,
) (bool, error) {
	blocked, err := e.blocks.Blocked(ctx, listenerID, ownerID)
	if err != nil || blocked {
		return false, err
	}
	recipientKey, err := e.keyer.Key(podapplication.MemberKeyNamespace, listenerID)
	if err != nil {
		return false, err
	}
	return e.pods.HoldsFor(ctx, recipientKey, assetID, e.now())
}

// voiceOfIntroductionEntitlement lets one member hear another's Voice of
// Introduction.
//
// This is the recording people meet each other through, so the rule is not
// who was invited to hear it — it is who has not been shut out. A block in
// either direction is the whole of it, which is the same rule the listening
// surface already applies (listeningBlockBridge) and the same one the sow
// path applies before it will let anybody reach.
//
// The media context has already established that the asset exists, is not
// deleted, is available now, and is being asked for under the introduction
// purpose. What is left for this to decide is the part about the two people.
type voiceOfIntroductionEntitlement struct {
	blocks sproutBlockBridge
}

func (e voiceOfIntroductionEntitlement) MayHear(
	ctx context.Context, listenerID, ownerID, _ string,
) (bool, error) {
	blocked, err := e.blocks.Blocked(ctx, listenerID, ownerID)
	if err != nil {
		return false, err
	}
	return !blocked, nil
}

// sowDeliveryBridge is what makes a sow arrive.
//
// A sow that screening cleared, or that a reviewer released, is delivered by
// being placed as a pod at the recipient's house front. Until this existed
// nothing read StatusDelivered at all: the sow was marked delivered, the seed
// stayed spent, and the recipient never learned anything had been sent.
//
// The pod is created as the sower, because it is their recording resting at
// somebody's door — the same act as POST /v1/seed/pods, reached from the
// other side. The command id is derived from the sow's own id so a retried
// delivery leaves one pod rather than two.
type sowDeliveryBridge struct {
	pods podPlacer
}

// podPlacer is the one thing delivery does.
type podPlacer interface {
	Create(context.Context, podapplication.Command, podapplication.Proposal) (podapplication.Result, error)
}

func (bridge sowDeliveryBridge) Place(ctx context.Context, sow sowapplication.Deliverable) error {
	// One pod per recording, each resting for the one person the sow reached.
	// A sow carries at most four, and the pod's own limit of twenty-five
	// recipients is irrelevant here: a sow is toward somebody, not to a room.
	for index, ref := range sow.MediaRefs {
		if _, err := bridge.pods.Create(ctx,
			podapplication.Command{
				ID:      fmt.Sprintf("sow-delivery:%s:%d", sow.SowID, index),
				ActorID: sow.SowerID,
			},
			podapplication.Proposal{
				OwnerID:      sow.SowerID,
				MediaRef:     ref,
				RecipientIDs: []string{sow.TargetID},
				TTL:          poddomain.RestingPeriod,
			},
		); err != nil {
			return err
		}
	}
	return nil
}

// sowAssetIDs names a sow's recording.
type sowAssetIDs struct{}

func (sowAssetIDs) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		// A weak identifier on a recording that will cost a member a seed is
		// worse than a crash at startup, which is where this would surface.
		panic(err)
	}
	return "sow_asset_" + hex.EncodeToString(value)
}

// purchaseKeyer joins the two digests a purchase needs.
//
// They come from different contexts with different keying rules — the member
// key is the membership context's, the phone reference is the payment
// context's — and this is the composition root's job precisely because
// neither context should know about the other.
type purchaseKeyer struct {
	members membershipprivacy.Keyer
	secret  []byte
}

func (k purchaseKeyer) MemberKey(memberID string) (string, error) {
	return k.members.MemberKey(memberID)
}

func (k purchaseKeyer) PhoneRef(phone string) (string, error) {
	return momoapplication.PhoneRef(k.secret, phone)
}

// organizationIssuerBridge answers the one question the promotion context asks
// about an organization: is it still a live relationship?
//
// A bool and nothing else. The promotion context has no business knowing an
// organization's name or its billing address, and a bridge that handed over
// the whole record would give it both.
type organizationIssuerBridge struct {
	organizations organizationapplication.Service
}

func (bridge organizationIssuerBridge) Issuing(
	ctx context.Context, organizationID string,
) (bool, error) {
	organization, err := bridge.organizations.Find(ctx, organizationID)
	if err != nil {
		return false, err
	}
	return organization.Issuing(), nil
}

// memberReceiptBridge finds where a payment receipt goes.
//
// A payment processor needs an email: it is where a receipt is sent and what a
// dispute attaches to. Handing one over is a deliberate disclosure to the
// processor the member is paying through and to nobody else — it is not
// written into any row this product keeps about the payment, and the intent
// still stores only digests.
type memberReceiptBridge struct {
	members interface {
		FindByID(context.Context, string) (memberdomain.Member, error)
	}
}

func (bridge memberReceiptBridge) Email(ctx context.Context, memberID string) (string, error) {
	member, err := bridge.members.FindByID(ctx, memberID)
	if err != nil {
		return "", err
	}
	return member.Email(), nil
}

// conversionBridge answers whether a referral has converted.
//
// Two questions, and both have to say yes: the member reached Tier 1 and has
// no upheld safety finding. Commission never accrues on a signup — a scheme
// paying per signup rewards exactly the bulk recruitment the tier ladder, age
// assurance and Sentinel exist to slow down.
type conversionBridge struct {
	tiers interface {
		Tier(ctx context.Context, memberID string) (identitydomain.Tier, error)
	}
	safety interface {
		HasUpheldAgainst(ctx context.Context, subjectID string) (bool, error)
	}
}

func (bridge conversionBridge) Verified(ctx context.Context, memberID string) (bool, error) {
	tier, err := bridge.tiers.Tier(ctx, memberID)
	if err != nil {
		return false, err
	}
	// Tier 1 or above, checked now rather than remembered from signup: a
	// member who verified and then lost it has not stayed.
	return tier != identitydomain.TierUnverified, nil
}

// Clean reports no upheld safety finding.
//
// Answered conservatively: if the safety context cannot be asked, the referral
// is not clean, it is unknown — and the sweep leaves an unknown pending rather
// than paying on it. Returning an error is what produces that.
func (bridge conversionBridge) Clean(ctx context.Context, memberID string) (bool, error) {
	upheld, err := bridge.safety.HasUpheldAgainst(ctx, memberID)
	if err != nil {
		return false, err
	}
	return !upheld, nil
}

// memberLookupBridge answers whether an identifier belongs to a member, which
// is the one rule keeping affiliates outside the community.
type memberLookupBridge struct {
	members interface {
		FindByEmail(context.Context, string) (memberdomain.Member, error)
	}
}

func (bridge memberLookupBridge) IsMember(ctx context.Context, email string) (bool, error) {
	_, err := bridge.members.FindByEmail(ctx, email)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, membermongo.ErrMemberNotFound) {
		return false, nil
	}
	// Not knowing is not "no". An affiliate admitted because the member
	// directory was unreachable is a member being paid to recruit.
	return false, err
}
