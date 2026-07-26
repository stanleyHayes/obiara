package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func TestLoggerRedactsSensitiveFieldsAndAddsContext(t *testing.T) {
	var output bytes.Buffer
	logger := NewJSONLogger(&output, LoggerConfig{
		Service:     "obiara-api",
		Version:     "test",
		Environment: "unit",
	})
	logger = logger.With(
		slog.String("authorization", "Bearer top-secret"),
		slog.Group("member",
			slog.String("email_address", "ama@example.test"),
			slog.String("tier", "two"),
		),
	)
	traceID, err := trace.TraceIDFromHex("70f5d3e6c96a3f6c4d8b0f7bb35c6512")
	if err != nil {
		t.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("3f6c4d8b0f7bb35c")
	if err != nil {
		t.Fatal(err)
	}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
	ctx = WithContextFields(ctx, ContextFields{
		RequestID:     "request-12345678",
		CorrelationID: "correlation-12345678",
		Operation:     "member.register",
	})

	LogAttrs(ctx, logger, slog.LevelInfo, "member.registered",
		slog.String("phone-number", "+233550000101"),
		slog.String("result", "accepted"),
	)

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode JSON log: %v\n%s", err, output.String())
	}
	assertValue(t, record, "msg", "member.registered")
	assertValue(t, record, "service", "obiara-api")
	assertValue(t, record, "request_id", "request-12345678")
	assertValue(t, record, "correlation_id", "correlation-12345678")
	assertValue(t, record, "trace_id", traceID.String())
	assertValue(t, record, "span_id", spanID.String())
	assertValue(t, record, "authorization", RedactedValue)
	assertValue(t, record, "phone-number", RedactedValue)
	assertValue(t, record, "result", "accepted")
	member, ok := record["member"].(map[string]any)
	if !ok {
		t.Fatalf("member group = %#v", record["member"])
	}
	assertValue(t, member, "email_address", RedactedValue)
	assertValue(t, member, "tier", "two")

	for _, forbidden := range []string{
		"top-secret",
		"ama@example.test",
		"+233550000101",
	} {
		if bytes.Contains(output.Bytes(), []byte(forbidden)) {
			t.Errorf("log contains sensitive value %q: %s", forbidden, output.String())
		}
	}
}

func TestContextRejectsUntrustedControlCharacters(t *testing.T) {
	ctx := WithContextFields(context.Background(), ContextFields{
		RequestID:     "request\nforged",
		CorrelationID: "safe-12345678",
		Operation:     "member register",
	})
	fields := FieldsFromContext(ctx)
	if fields.RequestID != "" || fields.Operation != "" {
		t.Fatalf("unsafe fields survived: %#v", fields)
	}
	if fields.CorrelationID != "safe-12345678" {
		t.Fatalf("correlation ID = %q", fields.CorrelationID)
	}
}

func TestMemoryMetricsIsRaceSafeAndSnapshotsAreIndependent(t *testing.T) {
	metrics := NewMemoryMetrics()
	const workers = 32
	const iterations = 200
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for range iterations {
				metrics.Add(context.Background(), "http.requests", 1)
				metrics.Observe(context.Background(), "http.duration", time.Millisecond)
			}
		}()
	}
	group.Wait()

	counters, observations := metrics.Snapshot()
	if counters["http.requests"] != workers*iterations {
		t.Fatalf("counter = %d", counters["http.requests"])
	}
	if len(observations["http.duration"]) != workers*iterations {
		t.Fatalf("observations = %d", len(observations["http.duration"]))
	}
	counters["http.requests"] = 0
	again, _ := metrics.Snapshot()
	if again["http.requests"] != workers*iterations {
		t.Fatal("snapshot mutated source")
	}
}

func TestInstrumentHealthRecordsStatusWithoutLeakingError(t *testing.T) {
	var output bytes.Buffer
	logger := NewJSONLogger(&output, LoggerConfig{Service: "api"})
	metrics := NewMemoryMetrics()
	want := errors.New("dial failed for ama@example.test with secret")
	check := InstrumentHealth("mongodb", func(context.Context) error {
		return want
	}, metrics, logger)

	if got := check(context.Background()); !errors.Is(got, want) {
		t.Fatalf("error = %v", got)
	}
	counters, observations := metrics.Snapshot()
	if counters["health.checks"] != 1 {
		t.Fatalf("checks = %d", counters["health.checks"])
	}
	if len(observations["health.check.duration"]) != 1 {
		t.Fatalf("durations = %d", len(observations["health.check.duration"]))
	}
	if bytes.Contains(output.Bytes(), []byte("ama@example.test")) {
		t.Fatalf("health log leaked dependency error: %s", output.String())
	}
}

func TestInvalidMetricNamesAreIgnored(t *testing.T) {
	metrics := NewMemoryMetrics()
	metrics.Add(context.Background(), "Member ID", 1)
	metrics.Observe(context.Background(), "valid.metric", -time.Second)
	counters, observations := metrics.Snapshot()
	if len(counters) != 0 || len(observations) != 0 {
		t.Fatalf("invalid metrics recorded: %#v %#v", counters, observations)
	}
}

func assertValue(t *testing.T, values map[string]any, key string, want any) {
	t.Helper()
	if got := values[key]; got != want {
		t.Fatalf("%s = %#v, want %#v", key, got, want)
	}
}
