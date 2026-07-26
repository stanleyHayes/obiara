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

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/services/api/internal/identity"
	identityapplication "github.com/stanleyHayes/obiara/services/api/internal/identity/application"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/member"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/health"
	apihttp "github.com/stanleyHayes/obiara/services/api/internal/platform/http"
	"github.com/stanleyHayes/obiara/services/api/internal/privacy"
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

	mux := http.NewServeMux()
	mux.Handle("GET /live", health.Live())
	mux.Handle("GET /ready", health.Ready(func(ctx context.Context) error {
		return client.Ping(ctx, readpref.Primary())
	}))
	apihttp.RegisterMemberRoutes(mux, memberModule.Register.Handle)
	apihttp.RegisterAuthRoutes(mux, identityModule.Registration)
	apihttp.RegisterVerificationRoutes(mux, verificationModule.Verification)
	apihttp.RegisterPrivacyRoutes(mux, privacyModule.Privacy)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apihttp.Correlation(mux),
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
