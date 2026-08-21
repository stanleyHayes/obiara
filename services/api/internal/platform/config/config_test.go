package config

import (
	"testing"
	"time"
)

func envWith(values map[string]string) func(string) string {
	return func(key string) string {
		return values[key]
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(envWith(nil))
	if err != nil {
		t.Fatalf("Load with defaults returned error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.MongoURI != "mongodb://localhost:27017" {
		t.Errorf("MongoURI = %q, want local default", cfg.MongoURI)
	}
	if cfg.MongoDatabase != "obiara" {
		t.Errorf("MongoDatabase = %q, want obiara", cfg.MongoDatabase)
	}
	if cfg.MongoConnectTimeout != 10*time.Second {
		t.Errorf("MongoConnectTimeout = %v, want 10s", cfg.MongoConnectTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 10s", cfg.ShutdownTimeout)
	}
}

func TestLoadOverrides(t *testing.T) {
	now := time.Now().UTC()
	cfg, err := Load(envWith(map[string]string{
		"PORT":                             "9090",
		"MONGODB_URI":                      "mongodb://mongo.internal:27017",
		"MONGODB_DATABASE":                 "obiara_staging",
		"MONGO_CONNECT_TIMEOUT":            "3s",
		"SHUTDOWN_TIMEOUT":                 "250ms",
		"OTEL_EXPORTER_OTLP_ENDPOINT":      "https://collector.example.test",
		"OTEL_EXPORTER_OTLP_INSECURE":      "false",
		"SERVICE_VERSION":                  "git-abc123",
		"APP_ENV":                          "staging",
		"MONGODB_URI_ROTATED_AT":           now.Format(time.RFC3339),
		"RESEND_WEBHOOK_SECRET":            "synthetic-test-only",
		"RESEND_WEBHOOK_SECRET_ROTATED_AT": now.Format(time.RFC3339),
		"LIVENESS_HMAC_SECRET":             "synthetic-liveness-secret-at-least-32-bytes",
		"LIVENESS_HMAC_SECRET_ROTATED_AT":  now.Format(time.RFC3339),
		"COMMERCE_HMAC_SECRET":             "synthetic-commerce-secret-at-least-32-bytes",
		"COMMERCE_HMAC_SECRET_ROTATED_AT":  now.Format(time.RFC3339),
		"ADMIN_HMAC_SECRET":                "synthetic-admin-secret-at-least-32-bytes",
		"ADMIN_HMAC_SECRET_ROTATED_AT":     now.Format(time.RFC3339),
		"NNOBOA_INVITE_SECRET":             "synthetic-nnoboa-secret-at-least-32-bytes",
		"NNOBOA_INVITE_SECRET_ROTATED_AT":  now.Format(time.RFC3339),
		"SEED_HMAC_SECRET":                 "synthetic-seed-secret-at-least-32-bytes",
		"SEED_HMAC_SECRET_ROTATED_AT":      now.Format(time.RFC3339),
		"CIRCLE_HMAC_SECRET":               "synthetic-circle-secret-at-least-32-bytes",
		"CIRCLE_HMAC_SECRET_ROTATED_AT":    now.Format(time.RFC3339),
		"OTP_PROVIDERS":                    "arkesel",
		"ARKESEL_API_KEY":                  "synthetic-test-only",
		"ARKESEL_SENDER_ID":                "Obiara",
		"EMAIL_PROVIDER":                   "resend",
		"RESEND_API_KEY":                   "synthetic-test-only",
		"RESEND_FROM_ADDRESS":              "no-reply@obiara.test",
		"WHATSAPP_PROVIDER":                "disabled",
		"IDENTITY_VERIFICATION_PROVIDER":   "manual",
		"LIVENESS_PROVIDER":                "manual",
	}))
	if err != nil {
		t.Fatalf("Load with overrides returned error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.MongoURI != "mongodb://mongo.internal:27017" {
		t.Errorf("MongoURI = %q, want override", cfg.MongoURI)
	}
	if cfg.MongoDatabase != "obiara_staging" {
		t.Errorf("MongoDatabase = %q, want override", cfg.MongoDatabase)
	}
	if cfg.MongoConnectTimeout != 3*time.Second {
		t.Errorf("MongoConnectTimeout = %v, want 3s", cfg.MongoConnectTimeout)
	}
	if cfg.ShutdownTimeout != 250*time.Millisecond {
		t.Errorf("ShutdownTimeout = %v, want 250ms", cfg.ShutdownTimeout)
	}
	if cfg.TelemetryEndpoint != "https://collector.example.test" || cfg.TelemetryInsecure ||
		cfg.ServiceVersion != "git-abc123" || cfg.Environment != "staging" {
		t.Errorf("telemetry config = %#v", cfg)
	}
}

func TestLoadStagingRejectsMissingOrStaleSecretMetadata(t *testing.T) {
	now := time.Date(2026, 7, 27, 6, 0, 0, 0, time.UTC)
	base := map[string]string{
		"APP_ENV": "staging", "MONGODB_URI": "synthetic-test-only",
		"RESEND_WEBHOOK_SECRET":            "synthetic-test-only",
		"LIVENESS_HMAC_SECRET":             "synthetic-liveness-secret-at-least-32-bytes",
		"MONGODB_URI_ROTATED_AT":           now.Format(time.RFC3339),
		"RESEND_WEBHOOK_SECRET_ROTATED_AT": now.Format(time.RFC3339),
		"LIVENESS_HMAC_SECRET_ROTATED_AT":  now.Format(time.RFC3339),
		"COMMERCE_HMAC_SECRET":             "synthetic-commerce-secret-at-least-32-bytes",
		"COMMERCE_HMAC_SECRET_ROTATED_AT":  now.Format(time.RFC3339),
		"ADMIN_HMAC_SECRET":                "synthetic-admin-secret-at-least-32-bytes",
		"ADMIN_HMAC_SECRET_ROTATED_AT":     now.Format(time.RFC3339),
		"NNOBOA_INVITE_SECRET":             "synthetic-nnoboa-secret-at-least-32-bytes",
		"NNOBOA_INVITE_SECRET_ROTATED_AT":  now.Format(time.RFC3339),
		"SEED_HMAC_SECRET":                 "synthetic-seed-secret-at-least-32-bytes",
		"SEED_HMAC_SECRET_ROTATED_AT":      now.Format(time.RFC3339),
		"CIRCLE_HMAC_SECRET":               "synthetic-circle-secret-at-least-32-bytes",
		"CIRCLE_HMAC_SECRET_ROTATED_AT":    now.Format(time.RFC3339),
		"OTP_PROVIDERS":                    "arkesel",
		"ARKESEL_API_KEY":                  "synthetic-test-only",
		"ARKESEL_SENDER_ID":                "Obiara",
		"EMAIL_PROVIDER":                   "resend",
		"RESEND_API_KEY":                   "synthetic-test-only",
		"RESEND_FROM_ADDRESS":              "no-reply@obiara.test",
		"WHATSAPP_PROVIDER":                "disabled",
		"IDENTITY_VERIFICATION_PROVIDER":   "manual",
		"LIVENESS_PROVIDER":                "manual",
	}
	if _, err := loadAt(envWith(base), now); err != nil {
		t.Fatal(err)
	}
	delete(base, "RESEND_WEBHOOK_SECRET_ROTATED_AT")
	if _, err := loadAt(envWith(base), now); err == nil {
		t.Fatal("staging accepted missing rotation metadata")
	}
}

func TestLoadInvalid(t *testing.T) {
	cases := map[string]map[string]string{
		"non-numeric port":     {"PORT": "abc"},
		"out-of-range port":    {"PORT": "70000"},
		"blank mongo uri":      {"MONGODB_URI": "   "},
		"blank database":       {"MONGODB_DATABASE": " "},
		"bad connect timeout":  {"MONGO_CONNECT_TIMEOUT": "soon"},
		"bad shutdown timeout": {"SHUTDOWN_TIMEOUT": "later"},
		"bad telemetry bool":   {"OTEL_EXPORTER_OTLP_INSECURE": "perhaps"},
	}
	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(envWith(env)); err == nil {
				t.Fatalf("Load(%v) succeeded, want error", env)
			}
		})
	}
}
