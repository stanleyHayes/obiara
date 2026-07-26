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

	notificationmongodb "github.com/stanleyHayes/obiara/internal/notifications/adapters/outbound/mongodb"
	notificationapplication "github.com/stanleyHayes/obiara/internal/notifications/application"
	ritualapplication "github.com/stanleyHayes/obiara/internal/notifications/ritual/application"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	privacymongodb "github.com/stanleyHayes/obiara/internal/privacy/adapters/outbound/mongodb"
	privacyapplication "github.com/stanleyHayes/obiara/internal/privacy/application"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
	privacyjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/privacy"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/relay"
	ritualjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/ritual"
	ritualmongodb "github.com/stanleyHayes/obiara/services/worker/internal/jobs/ritual/adapters/outbound/mongodb"
	workertelemetry "github.com/stanleyHayes/obiara/services/worker/internal/telemetry"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "worker startup failed:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	telemetryRuntime, err := workertelemetry.New(ctx, os.Stdout, workertelemetry.Config{
		Version:     envOrDefault("SERVICE_VERSION", "dev"),
		Environment: envOrDefault("APP_ENV", "development"),
		Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Insecure:    envOrDefault("OTEL_EXPORTER_OTLP_INSECURE", "false") == "true",
	})
	if err != nil {
		return fmt.Errorf("configure telemetry: %w", err)
	}
	logger := telemetryRuntime.Logger
	ctx = telemetryRuntime.Started(ctx, 1)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = telemetryRuntime.Shutdown(shutdownCtx)
	}()

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

	// Privacy processor (S4-005): executes export/deletion requests with
	// the FR-106 statutory clocks.
	privacyRequests := privacymongodb.NewRequestRepository(database)
	if err := privacyRequests.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure privacy indexes: %w", err)
	}
	privacyProcessor := privacyapplication.NewProcessor(
		privacyRequests,
		privacymongodb.NewArchiveAssembler(database, time.Now),
		privacymongodb.NewErasureRunner(database, time.Now),
		mongodb.NewSessionRevoker(database, time.Now),
		time.Now,
	)

	// Ritual dispatch (E13-S02): calendar rituals and fire heralds through
	// the E13-S01 preference boundary and the durable outbox.
	notificationRepository := notificationmongodb.NewRepository(database)
	decider := notificationapplication.NewNotificationService(notificationRepository, notificationRepository, time.Now)
	ritualSources := ritualmongodb.NewSources(database)
	ritualDispatcher := ritualapplication.NewDispatcher(
		ritualSources,
		ritualSources,
		decider,
		ritualSources,
		outboxStore,
		inbox.NewStore(database, time.Now),
		time.Now,
	)

	scheduler := application.NewScheduler([]application.Job{
		relay.NewOutboxJob(outboxStore, loggingPublisher{logger: logger}, 100, 5*time.Second),
		privacyjob.NewProcessorJob(privacyProcessor, 25, 60*time.Second),
		ritualjob.NewCalendarJob(ritualDispatcher, 5*time.Minute),
		ritualjob.NewHeraldJob(ritualDispatcher, 5*time.Minute),
	}, mongodb.NewDeadLetterStore(database, time.Now), logger, time.Now)

	logger.InfoContext(ctx, "worker started", slog.Int("jobs", 4))
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
