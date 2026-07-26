// Command worker is the composition root for the Obiara background worker
// (agent_plan.md §7.1): durable jobs, outbox relay, expiry/state
// transitions and notifications build on this scheduler (E02-S09).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/relay"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker startup failed:", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := apimongo.Connect(connectCtx, envOrDefault("MONGODB_URI", "mongodb://localhost:27017"))
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer disconnectCancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	database := client.Database(envOrDefault("MONGODB_DATABASE", "obiara"))
	outboxStore := outbox.NewStore(database, time.Now)
	if err := outboxStore.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure outbox indexes: %w", err)
	}

	scheduler := application.NewScheduler([]application.Job{
		relay.NewOutboxJob(outboxStore, loggingPublisher{logger: logger}, 100, 5*time.Second),
	}, mongodb.NewDeadLetterStore(database, time.Now), logger, time.Now)

	logger.InfoContext(ctx, "worker started", slog.Int("jobs", 1))
	if err := jobs.NewModule(scheduler).Run(ctx); err != nil {
		return err
	}
	logger.InfoContext(ctx, "worker stopped")
	return nil
}

// loggingPublisher is the placeholder publisher until notification and
// projection consumers land; it proves the relay loop end-to-end without
// leaking payloads (only bounded identifiers are logged).
type loggingPublisher struct {
	logger *slog.Logger
}

func (publisher loggingPublisher) Publish(ctx context.Context, record outbox.Record) error {
	publisher.logger.InfoContext(ctx, "outbox record published",
		slog.String("eventType", record.EventType),
		slog.String("aggregateType", record.AggregateType))
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
