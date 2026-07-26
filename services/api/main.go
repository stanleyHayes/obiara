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

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/stanleyHayes/obiara/services/api/internal/member"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/config"
	"github.com/stanleyHayes/obiara/services/api/internal/platform/health"
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

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	connectCtx, cancel := context.WithTimeout(ctx, cfg.MongoConnectTimeout)
	defer cancel()
	if err := client.Ping(connectCtx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping mongo (start a local MongoDB or set MONGODB_URI): %w", err)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	// Modules are composed here at startup (agent_plan.md §7.2). Inbound
	// HTTP adapters for module use cases arrive with the API envelope task
	// (S1-003); until then the member module is built but not routed.
	if _, err := member.NewModule(ctx, client.Database(cfg.MongoDatabase)); err != nil {
		return fmt.Errorf("build member module: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /live", health.Live())
	mux.Handle("GET /ready", health.Ready(func(ctx context.Context) error {
		return client.Ping(ctx, readpref.Primary())
	}))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
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
