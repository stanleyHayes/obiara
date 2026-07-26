package telemetry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// RuntimeConfig is the vendor-neutral OTLP/HTTP composition contract.
// An empty endpoint deliberately selects local no-export mode.
type RuntimeConfig struct {
	Service     string
	Version     string
	Environment string
	Endpoint    string
	Insecure    bool
}

type Runtime struct {
	Logger  *slog.Logger
	Metrics Metrics
	traces  *sdktrace.TracerProvider
	meters  *sdkmetric.MeterProvider
}

func NewRuntime(ctx context.Context, output LoggerOutput, config RuntimeConfig) (*Runtime, error) {
	logger := NewJSONLogger(output, LoggerConfig{
		Service: config.Service, Version: config.Version, Environment: config.Environment,
	})
	runtime := &Runtime{Logger: logger, Metrics: NoopMetrics{}}
	if strings.TrimSpace(config.Endpoint) == "" {
		return runtime, nil
	}
	if err := validateEndpoint(config.Endpoint, config.Insecure); err != nil {
		return nil, err
	}

	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName(config.Service),
		semconv.ServiceVersion(config.Version),
		attribute.String("deployment.environment.name", config.Environment),
	))
	if err != nil {
		return nil, fmt.Errorf("build telemetry resource: %w", err)
	}
	traceOptions := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(config.Endpoint)}
	metricOptions := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpointURL(config.Endpoint)}
	if config.Insecure {
		traceOptions = append(traceOptions, otlptracehttp.WithInsecure())
		metricOptions = append(metricOptions, otlpmetrichttp.WithInsecure())
	}
	traceExporter, err := otlptracehttp.New(ctx, traceOptions...)
	if err != nil {
		return nil, fmt.Errorf("build OTLP trace exporter: %w", err)
	}
	metricExporter, err := otlpmetrichttp.New(ctx, metricOptions...)
	if err != nil {
		return nil, fmt.Errorf("build OTLP metric exporter: %w", err)
	}
	runtime.traces = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	runtime.meters = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetTracerProvider(runtime.traces)
	otel.SetMeterProvider(runtime.meters)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	runtime.Metrics = newOTelMetrics(runtime.meters.Meter("obiara/runtime"))
	return runtime, nil
}

// Shutdown flushes both providers and preserves every shutdown error.
func (runtime *Runtime) Shutdown(ctx context.Context) error {
	if runtime == nil {
		return nil
	}
	var errs []error
	if runtime.meters != nil {
		errs = append(errs, runtime.meters.Shutdown(ctx))
	}
	if runtime.traces != nil {
		errs = append(errs, runtime.traces.Shutdown(ctx))
	}
	return errors.Join(errs...)
}

// HTTP instruments server traces and bounded RED metrics. correlationID must
// return the already validated transport identifier.
func (runtime *Runtime) HTTP(next http.Handler, correlationID func(context.Context) string) http.Handler {
	if runtime == nil {
		return next
	}
	instrumented := http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := WithContextFields(request.Context(), ContextFields{
			CorrelationID: correlationID(request.Context()),
			Operation:     request.Method,
		})
		started := time.Now()
		runtime.Metrics.Add(ctx, "http.server.requests", 1,
			Attribute{Key: "method", Value: request.Method})
		next.ServeHTTP(w, request.WithContext(ctx))
		runtime.Metrics.Observe(ctx, "http.server.duration", time.Since(started),
			Attribute{Key: "method", Value: request.Method})
		LogAttrs(ctx, runtime.Logger, slog.LevelInfo, "http.request.completed",
			slog.String("method", request.Method),
			slog.String("route", routeClass(request.URL.Path)),
		)
	})
	return otelhttp.NewHandler(instrumented, "obiara.http",
		otelhttp.WithSpanNameFormatter(func(_ string, request *http.Request) string {
			return request.Method + " " + routeClass(request.URL.Path)
		}),
	)
}

func validateEndpoint(raw string, insecure bool) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must be an absolute http(s) URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must not contain credentials, query or fragment")
	}
	if parsed.Scheme == "http" && !insecure {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_INSECURE must be true for an http endpoint")
	}
	return nil
}

func routeClass(path string) string {
	switch path {
	case "/live", "/ready":
		return path
	default:
		return "/api"
	}
}

type otelMetrics struct {
	counter   metric.Int64Counter
	histogram metric.Float64Histogram
}

func newOTelMetrics(meter metric.Meter) Metrics {
	counter, _ := meter.Int64Counter("obiara.events")
	histogram, _ := meter.Float64Histogram("obiara.duration", metric.WithUnit("s"))
	return otelMetrics{counter: counter, histogram: histogram}
}

func (adapter otelMetrics) Add(ctx context.Context, name string, delta int64, attrs ...Attribute) {
	if !validMetricName(name) {
		return
	}
	adapter.counter.Add(ctx, delta, metric.WithAttributes(metricAttrs(name, attrs)...))
}

func (adapter otelMetrics) Observe(ctx context.Context, name string, value time.Duration, attrs ...Attribute) {
	if !validMetricName(name) || value < 0 {
		return
	}
	adapter.histogram.Record(ctx, value.Seconds(), metric.WithAttributes(metricAttrs(name, attrs)...))
}

func metricAttrs(name string, attrs []Attribute) []attribute.KeyValue {
	result := []attribute.KeyValue{attribute.String("operation", name)}
	for _, item := range attrs {
		if validDimension(item) {
			result = append(result, attribute.String(item.Key, item.Value))
		}
	}
	return result
}

func validDimension(item Attribute) bool {
	if !validMetricName(item.Key) || item.Value == "" || len(item.Value) > 32 {
		return false
	}
	for _, r := range item.Value {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') &&
			r != '_' && r != '-' && r != '.' && r != '/' {
			return false
		}
	}
	return true
}
