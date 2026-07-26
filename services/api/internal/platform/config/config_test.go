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
	cfg, err := Load(envWith(map[string]string{
		"PORT":                        "9090",
		"MONGODB_URI":                 "mongodb://mongo.internal:27017",
		"MONGODB_DATABASE":            "obiara_staging",
		"MONGO_CONNECT_TIMEOUT":       "3s",
		"SHUTDOWN_TIMEOUT":            "250ms",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "https://collector.example.test",
		"OTEL_EXPORTER_OTLP_INSECURE": "false",
		"SERVICE_VERSION":             "git-abc123",
		"APP_ENV":                     "staging",
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
