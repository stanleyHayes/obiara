// Command worker is the composition root for the Obiara background worker
// (agent_plan.md §7.1): durable jobs, outbox relay, expiry/state
// transitions and notifications build on this scheduler (E02-S09).
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	notificationmongodb "github.com/stanleyHayes/obiara/internal/notifications/adapters/outbound/mongodb"
	notificationapplication "github.com/stanleyHayes/obiara/internal/notifications/application"
	inappmongodb "github.com/stanleyHayes/obiara/internal/notifications/inapp/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/internal/notifications/push"
	pushdisabled "github.com/stanleyHayes/obiara/internal/notifications/push/adapters/outbound/disabled"
	pushexpo "github.com/stanleyHayes/obiara/internal/notifications/push/adapters/outbound/expo"
	pushapp "github.com/stanleyHayes/obiara/internal/notifications/push/application"
	ritualapplication "github.com/stanleyHayes/obiara/internal/notifications/ritual/application"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/adapters/outbound/inappsender"
	"github.com/stanleyHayes/obiara/internal/notifications/routing/adapters/outbound/pushsender"
	routingapplication "github.com/stanleyHayes/obiara/internal/notifications/routing/application"
	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/platform/retention"
	"github.com/stanleyHayes/obiara/internal/platform/secrets"
	privacymongodb "github.com/stanleyHayes/obiara/internal/privacy/adapters/outbound/mongodb"
	privacyapplication "github.com/stanleyHayes/obiara/internal/privacy/application"
	safetymongodb "github.com/stanleyHayes/obiara/internal/safety/adapters/outbound/mongodb"
	safetyapplication "github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
	privacyjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/privacy"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/relay"
	retentionjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/retention"
	ritualjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/ritual"
	ritualmongodb "github.com/stanleyHayes/obiara/services/worker/internal/jobs/ritual/adapters/outbound/mongodb"
	safetyjob "github.com/stanleyHayes/obiara/services/worker/internal/jobs/safety"
	"github.com/stanleyHayes/obiara/services/worker/internal/jobs/safety/reactivation"
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
	environment := envOrDefault("APP_ENV", "development")
	if err := secrets.ValidateRuntime(secrets.Worker, environment, os.Getenv, time.Now().UTC()); err != nil {
		return fmt.Errorf("runtime secret policy: %w", err)
	}
	telemetryRuntime, err := workertelemetry.New(ctx, os.Stdout, workertelemetry.Config{
		Version:     envOrDefault("SERVICE_VERSION", "dev"),
		Environment: environment,
		Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		Insecure:    envOrDefault("OTEL_EXPORTER_OTLP_INSECURE", "false") == "true",
	})
	if err != nil {
		return fmt.Errorf("configure telemetry: %w", err)
	}
	logger := telemetryRuntime.Logger
	ctx = telemetryRuntime.Started(ctx, 1)
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), telemetryShutdownTimeout)
		defer shutdownCancel()
		_ = telemetryRuntime.Shutdown(shutdownCtx)
	}()

	// The worker honours the same connect and shutdown budgets as the API.
	// These were hardcoded to ten seconds, which a cold Atlas cluster
	// routinely exceeds; on Render that is a crash loop with no knob to
	// turn.
	connectTimeout, err := durationOrDefault("MONGO_CONNECT_TIMEOUT", 10*time.Second)
	if err != nil {
		return err
	}
	shutdownTimeout, err := durationOrDefault("SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return err
	}

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	client, err := apimongo.Connect(connectCtx, envOrDefault("MONGODB_URI", "mongodb://localhost:27017"))
	if err != nil {
		return err
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), shutdownTimeout)
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

	// T&S case builder (E12-S02): filed reports become tiered cases with
	// SLA deadlines.
	safetyReportRepository := safetymongodb.NewRepository(database)
	safetyCaseRepository := safetymongodb.NewCaseRepository(database)
	safetyCaseService := safetyapplication.NewCaseService(safetyCaseRepository, time.Now, newID)
	safetyBuilder := safetyjob.NewCaseBuilder(outboxStore, safetyReportRepository, safetyCaseService, inbox.NewStore(database, time.Now))

	// Notification routing (E13-S03): the outbox relay delivers member-facing
	// events through the channel ladder. Push is not provisioned, so the
	// ritual ladder falls through to the always-available in-app inbox; the
	// router skips channels it has no sender for.
	inAppStore := inappmongodb.NewStore(database)
	if err := inAppStore.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure in-app notification indexes: %w", err)
	}
	pushSender, err := workerPushSender()
	if err != nil {
		return err
	}
	pushModule, err := push.NewModule(ctx, database, pushSender)
	if err != nil {
		return fmt.Errorf("build push module: %w", err)
	}
	// Ladder order follows domain.LadderFor: push first, in-app as the
	// channel that always works. A member with no registered device fails the
	// push rung and falls through rather than losing the notification.
	router := routingapplication.NewRouter(
		[]routingapplication.ChannelSender{
			pushsender.New(pushModule.Push),
			inappsender.New(inAppStore, time.Now),
		},
		decider,
		time.Now,
	)

	scheduler := application.NewScheduler([]application.Job{
		relay.NewOutboxJob(outboxStore, relay.NewNotifyingPublisher(router, logger), 100, 5*time.Second),
		privacyjob.NewProcessorJob(privacyProcessor, 25, 60*time.Second),
		ritualjob.NewCalendarJob(ritualDispatcher, 5*time.Minute),
		ritualjob.NewHeraldJob(ritualDispatcher, 5*time.Minute),
		safetyjob.NewBuilderJob(safetyBuilder, 50, 30*time.Second),
		reactivation.NewJob(reactivation.NewStore(database, time.Now), 5*time.Minute),
		retentionjob.NewJob(retention.NewRunner(database, retention.BindingPolicies, time.Now), 6*time.Hour),
	}, mongodb.NewDeadLetterStore(database, time.Now), logger, time.Now)

	logger.InfoContext(ctx, "worker started", slog.Int("jobs", 7))
	if err := jobs.NewModule(scheduler).Run(ctx); err != nil {
		return err
	}
	logger.InfoContext(ctx, "worker stopped")
	return nil
}

func newID() string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return "case_" + base64.RawURLEncoding.EncodeToString(id)
}

// workerPushSender selects the push adapter.
//
// The worker cannot reuse the API's delivery builder, which lives under
// services/api/internal and is therefore closed to it, so selection is
// repeated here over the same environment variable. Like every other
// channel, the simulator is refused outside development: a channel that
// reports delivery without delivering is the failure this codebase has
// already paid for once.
func workerPushSender() (pushapp.Sender, error) {
	provider := strings.ToLower(envOrDefault("PUSH_PROVIDER", "disabled"))
	environment := strings.ToLower(strings.TrimSpace(envOrDefault("APP_ENV", "development")))
	simulatorsAllowed := environment == "development" || environment == "test" || environment == "local"

	switch provider {
	case "expo":
		return pushexpo.NewSender(pushexpo.Config{
			AccessToken: strings.TrimSpace(os.Getenv("EXPO_ACCESS_TOKEN")),
			BaseURL:     strings.TrimSpace(os.Getenv("EXPO_BASE_URL")),
		}, &http.Client{Timeout: 10 * time.Second})
	case "disabled":
		return pushdisabled.NewSender(), nil
	case "simulator":
		if !simulatorsAllowed {
			return nil, fmt.Errorf(
				"PUSH_PROVIDER may not be %q outside development; use %q to take the channel out of service",
				"simulator", "disabled")
		}
		return pushdisabled.NewSender(), nil
	default:
		return nil, fmt.Errorf("PUSH_PROVIDER must be \"expo\", \"disabled\" or \"simulator\", got %q", provider)
	}
}

// telemetryShutdownTimeout bounds exporter flush at exit. It is fixed
// because it runs before the configured shutdown budget is in scope.
const telemetryShutdownTimeout = 10 * time.Second

// durationOrDefault reads a Go duration from the environment, failing loudly
// on a malformed value rather than silently falling back.
func durationOrDefault(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a Go duration such as 30s, got %q", key, value)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be positive, got %q", key, value)
	}
	return parsed, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
