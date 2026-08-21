// Package config loads the api service configuration from the environment.
// Secrets arrive only via environment variables; none are ever committed
// (agent_plan.md §12 deployment gates).
package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/internal/platform/secrets"
)

// Config is the runtime configuration for the api composition root.
type Config struct {
	Port                string
	MongoURI            string
	MongoDatabase       string
	MongoConnectTimeout time.Duration
	ShutdownTimeout     time.Duration
	TelemetryEndpoint   string
	TelemetryInsecure   bool
	ServiceVersion      string
	Environment         string
	LivenessHMACSecret  string
	CommerceHMACSecret  string
	AdminHMACSecret     string
	NnoboaInviteSecret  string
	SeedHMACSecret      string
	CircleHMACSecret    string
	// Notifications selects the outbound OTP, WhatsApp and email adapters.
	Notifications NotificationsConfig
	// Verification selects the identity and liveness provider adapters.
	Verification VerificationConfig
}

// simulatorsAllowed reports whether the environment may run without real
// delivery providers. Only local development and automated tests may.
func simulatorsAllowed(environment string) bool {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "development", "test", "local":
		return true
	default:
		return false
	}
}

// Load reads configuration using getenv (os.Getenv in production) and
// validates it. Defaults target local development; staging and production
// must set every value explicitly through the environment matrix.
func Load(getenv func(string) string) (Config, error) {
	return loadAt(getenv, time.Now().UTC())
}

func loadAt(getenv func(string) string, now time.Time) (Config, error) {
	cfg := Config{
		Port:              valueOrDefault(getenv("PORT"), "8080"),
		MongoURI:          valueOrDefault(getenv("MONGODB_URI"), "mongodb://localhost:27017"),
		MongoDatabase:     valueOrDefault(getenv("MONGODB_DATABASE"), "obiara"),
		TelemetryEndpoint: strings.TrimSpace(getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		ServiceVersion:    valueOrDefault(getenv("SERVICE_VERSION"), "dev"),
		Environment:       valueOrDefault(getenv("APP_ENV"), "development"),
		LivenessHMACSecret: valueOrDefault(
			getenv("LIVENESS_HMAC_SECRET"),
			"obiara-local-liveness-key-change-before-production",
		),
		CommerceHMACSecret: valueOrDefault(
			getenv("COMMERCE_HMAC_SECRET"),
			"obiara-local-commerce-key-change-before-production",
		),
		AdminHMACSecret: valueOrDefault(
			getenv("ADMIN_HMAC_SECRET"),
			"obiara-local-admin-key-change-before-production",
		),
		NnoboaInviteSecret: valueOrDefault(
			getenv("NNOBOA_INVITE_SECRET"),
			"obiara-local-nnoboa-invite-key-change-before-production",
		),
		SeedHMACSecret: valueOrDefault(
			getenv("SEED_HMAC_SECRET"),
			"obiara-local-seed-key-change-before-production",
		),
		CircleHMACSecret: valueOrDefault(
			getenv("CIRCLE_HMAC_SECRET"),
			"obiara-local-circle-key-change-before-production",
		),
	}

	var err error
	if cfg.MongoConnectTimeout, err = durationOrDefault(getenv("MONGO_CONNECT_TIMEOUT"), 10*time.Second); err != nil {
		return Config{}, fmt.Errorf("MONGO_CONNECT_TIMEOUT: %w", err)
	}
	if cfg.ShutdownTimeout, err = durationOrDefault(getenv("SHUTDOWN_TIMEOUT"), 10*time.Second); err != nil {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT: %w", err)
	}
	if cfg.TelemetryInsecure, err = boolOrDefault(getenv("OTEL_EXPORTER_OTLP_INSECURE"), false); err != nil {
		return Config{}, fmt.Errorf("OTEL_EXPORTER_OTLP_INSECURE: %w", err)
	}

	port, convErr := strconv.Atoi(cfg.Port)
	if convErr != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be a number between 1 and 65535, got %q", cfg.Port)
	}
	if strings.TrimSpace(cfg.MongoURI) == "" {
		return Config{}, fmt.Errorf("MONGODB_URI must not be empty")
	}
	if strings.TrimSpace(cfg.MongoDatabase) == "" {
		return Config{}, fmt.Errorf("MONGODB_DATABASE must not be empty")
	}
	if len(cfg.LivenessHMACSecret) < 32 {
		return Config{}, fmt.Errorf("LIVENESS_HMAC_SECRET must contain at least 32 bytes")
	}
	if len(cfg.CommerceHMACSecret) < 32 {
		return Config{}, fmt.Errorf("COMMERCE_HMAC_SECRET must contain at least 32 bytes")
	}
	if len(cfg.AdminHMACSecret) < 32 {
		return Config{}, fmt.Errorf("ADMIN_HMAC_SECRET must contain at least 32 bytes")
	}
	if len(cfg.NnoboaInviteSecret) < 32 {
		return Config{}, fmt.Errorf("NNOBOA_INVITE_SECRET must contain at least 32 bytes")
	}
	if len(cfg.SeedHMACSecret) < 32 {
		return Config{}, fmt.Errorf("SEED_HMAC_SECRET must contain at least 32 bytes")
	}
	if len(cfg.CircleHMACSecret) < 32 {
		return Config{}, fmt.Errorf("CIRCLE_HMAC_SECRET must contain at least 32 bytes")
	}
	if err := secrets.ValidateRuntime(secrets.API, cfg.Environment, getenv, now); err != nil {
		return Config{}, fmt.Errorf("runtime secret policy: %w", err)
	}

	allowSimulators := simulatorsAllowed(cfg.Environment)
	cfg.Notifications = loadNotifications(getenv)
	if err := validateNotifications(cfg.Notifications, allowSimulators); err != nil {
		return Config{}, fmt.Errorf("notification providers: %w", err)
	}
	cfg.Verification = loadVerification(getenv)
	if err := validateVerification(cfg.Verification, allowSimulators); err != nil {
		return Config{}, fmt.Errorf("verification providers: %w", err)
	}

	return cfg, nil
}

func boolOrDefault(value string, fallback bool) (bool, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("must be true or false, got %q", value)
	}
	return parsed, nil
}

func valueOrDefault(value, fallback string) string {
	// Only an unset (empty) variable falls back; an explicitly set
	// whitespace-only value is a configuration mistake and must fail
	// validation rather than silently becoming the default.
	if value != "" {
		return value
	}
	return fallback
}

func durationOrDefault(value string, fallback time.Duration) (time.Duration, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("must be a Go duration such as 10s, got %q", value)
	}
	return parsed, nil
}
