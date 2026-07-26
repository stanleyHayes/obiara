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
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/stanleyHayes/obiara/internal/notifications"
	"github.com/stanleyHayes/obiara/internal/notifications/email"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/privacy"
	"github.com/stanleyHayes/obiara/internal/safety"
	"github.com/stanleyHayes/obiara/services/api/internal/fire"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/ember"
	"github.com/stanleyHayes/obiara/services/api/internal/identity"
	identityapplication "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/member"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/health"
	apihttp "github.com/stanleyHayes/obiara/services/api/internal/platform/http"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/telemetry"
	"github.com/stanleyHayes/obiara/services/api/internal/profile"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening"
	"github.com/stanleyHayes/obiara/services/api/internal/trust"
	"github.com/stanleyHayes/obiara/services/api/internal/verification"
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

	connectCtx, cancel := context.WithTimeout(ctx, cfg.MongoConnectTimeout)
	defer cancel()
	client, err := apimongo.Connect(connectCtx, cfg.MongoURI)
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	// Modules are composed here at startup (agent_plan.md §7.2).
	memberModule, err := member.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build member module: %w", err)
	}
	// The identity module provides session issuance and phone OTP
	// registration (E03-S01).
	identityModule, err := identity.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build identity module: %w", err)
	}
	// Verification (E03-S03) promotes accounts through the identity tier
	// state machine via the composition-time bridge.
	verificationModule, err := verification.NewModule(ctx, client.Database(cfg.MongoDatabase), tierBridge{tiers: identityModule.Tiers})
	if err != nil {
		return fmt.Errorf("build verification module: %w", err)
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
	// Fire scheduling and attendance (E09-S01).
	fireModule, err := fire.NewModule(ctx, client.Database(cfg.MongoDatabase))
	if err != nil {
		return fmt.Errorf("build fire module: %w", err)
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
	// Transactional email (E13-S04): Resend channel with signed delivery
	// webhooks.
	emailModule, err := email.NewModule(ctx, client.Database(cfg.MongoDatabase), os.Getenv("RESEND_WEBHOOK_SECRET"))
	if err != nil {
		return fmt.Errorf("build email module: %w", err)
	}
	// Safety intake (E12-S01): reports ride the durable outbox to queue
	// processors.
	safetyOutbox := outbox.NewStore(client.Database(cfg.MongoDatabase), time.Now)
	if err := safetyOutbox.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure outbox indexes: %w", err)
	}
	safetyModule, err := safety.NewModule(ctx, client.Database(cfg.MongoDatabase), safetyOutbox)
	if err != nil {
		return fmt.Errorf("build safety module: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /live", health.Live())
	mux.Handle("GET /ready", health.Ready(func(ctx context.Context) error {
		return client.Ping(ctx, readpref.Primary())
	}))
	apihttp.RegisterMemberRoutes(mux, memberModule.Register.Handle)
	apihttp.RegisterAuthRoutes(mux, identityModule.Registration)
	apihttp.RegisterVerificationRoutes(mux, verificationModule.Verification)
	apihttp.RegisterPrivacyRoutes(mux, privacyModule.Privacy)
	apihttp.RegisterTrustVisibilityRoutes(mux, trustModule.Visibility, identityModule.Sessions)
	apihttp.RegisterDoorwayRoutes(mux, profileModule.Doorway, profileModule.Vault)
	apihttp.RegisterListeningRoutes(mux, listeningModule.Listening)
	apihttp.RegisterFireRoutes(mux, fireModule.Fires)
	apihttp.RegisterEmberRoutes(mux, emberModule.Embers)
	apihttp.RegisterNotificationRoutes(mux, notificationModule.Notifications)
	apihttp.RegisterSafetyRoutes(mux, safetyModule.Safety)
	apihttp.RegisterResendWebhookRoute(mux, emailModule.Webhook, inbox.NewStore(client.Database(cfg.MongoDatabase), time.Now))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apihttp.Correlation(telemetryRuntime.HTTP(mux, apihttp.CorrelationID)),
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
