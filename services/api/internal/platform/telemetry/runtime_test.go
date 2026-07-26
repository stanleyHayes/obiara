package telemetry

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPRuntimeCorrelatesLogTraceAndREDMetrics(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	traces := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	reader := sdkmetric.NewManualReader()
	meters := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	originalTracer := otel.GetTracerProvider()
	originalMeter := otel.GetMeterProvider()
	otel.SetTracerProvider(traces)
	otel.SetMeterProvider(meters)
	t.Cleanup(func() {
		otel.SetTracerProvider(originalTracer)
		otel.SetMeterProvider(originalMeter)
	})

	var logs bytes.Buffer
	runtime := &Runtime{
		Logger:  NewJSONLogger(&logs, LoggerConfig{Service: "test-api"}),
		Metrics: newOTelMetrics(meters.Meter("test")),
	}
	handler := runtime.HTTP(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), func(context.Context) string { return "correlation-12345678" })
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ready", nil))

	spans := recorder.Ended()
	if len(spans) != 1 || spans[0].Name() != "GET /ready" {
		t.Fatalf("spans = %#v", spans)
	}
	if got := logs.String(); !strings.Contains(got, `"correlation_id":"correlation-12345678"`) ||
		!strings.Contains(got, `"trace_id":"`) ||
		strings.Contains(got, "member") {
		t.Fatalf("unexpected correlated log: %s", got)
	}
	var metrics metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &metrics); err != nil {
		t.Fatal(err)
	}
	if len(metrics.ScopeMetrics) == 0 || len(metrics.ScopeMetrics[0].Metrics) != 2 {
		t.Fatalf("metrics = %#v", metrics.ScopeMetrics)
	}
}

func TestRuntimeRejectsUnsafeExporterEndpoint(t *testing.T) {
	_, err := NewRuntime(context.Background(), &bytes.Buffer{}, RuntimeConfig{
		Service: "api", Endpoint: "http://user:secret@example.test?token=value",
		Insecure: true,
	})
	if err == nil {
		t.Fatal("expected unsafe endpoint rejection")
	}
}
