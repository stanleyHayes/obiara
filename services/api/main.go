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
	"github.com/stanleyHayes/obiara/services/api/internal/member"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/health"
	apihttp "github.com/stanleyHayes/obiara/services/api/internal/platform/http"
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

	mux := http.NewServeMux()
	mux.Handle("GET /live", health.Live())
	mux.Handle("GET /ready", health.Ready(func(ctx context.Context) error {
		return client.Ping(ctx, readpref.Primary())
	}))
	apihttp.RegisterMemberRoutes(mux, memberModule.Register.Handle)
	apihttp.RegisterAuthRoutes(mux, identityModule.Registration)

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
