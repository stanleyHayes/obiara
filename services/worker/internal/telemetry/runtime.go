package telemetry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

type Config struct {
	Version, Environment, Endpoint string
	Insecure                       bool
}

type Runtime struct {
	Logger *slog.Logger
	meter  metric.Meter
	traces *sdktrace.TracerProvider
	meters *sdkmetric.MeterProvider
}

func New(ctx context.Context, output io.Writer, config Config) (*Runtime, error) {
	logger := slog.New(slog.NewJSONHandler(output, nil)).With(
		slog.String("service", "obiara-worker"),
		slog.String("service_version", config.Version),
		slog.String("environment", config.Environment),
	)
	result := &Runtime{Logger: logger, meter: otel.GetMeterProvider().Meter("obiara/worker")}
	if strings.TrimSpace(config.Endpoint) == "" {
		return result, nil
	}
	parsed, err := url.Parse(config.Endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must be an absolute credential-free http(s) URL")
	}
	if parsed.Scheme == "http" && !config.Insecure {
		return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_INSECURE must be true for an http endpoint")
	}
	traceOptions := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(config.Endpoint)}
	metricOptions := []otlpmetrichttp.Option{otlpmetrichttp.WithEndpointURL(config.Endpoint)}
	if config.Insecure {
		traceOptions = append(traceOptions, otlptracehttp.WithInsecure())
		metricOptions = append(metricOptions, otlpmetrichttp.WithInsecure())
	}
	traceExporter, err := otlptracehttp.New(ctx, traceOptions...)
	if err != nil {
		return nil, fmt.Errorf("build worker trace exporter: %w", err)
	}
	metricExporter, err := otlpmetrichttp.New(ctx, metricOptions...)
	if err != nil {
		return nil, fmt.Errorf("build worker metric exporter: %w", err)
	}
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName("obiara-worker"),
		semconv.ServiceVersion(config.Version),
		attribute.String("deployment.environment.name", config.Environment),
	))
	if err != nil {
		return nil, fmt.Errorf("build worker telemetry resource: %w", err)
	}
	result.traces = sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(res))
	result.meters = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetTracerProvider(result.traces)
	otel.SetMeterProvider(result.meters)
	result.meter = result.meters.Meter("obiara/worker")
	return result, nil
}

func (runtime *Runtime) Started(ctx context.Context, jobs int64) context.Context {
	tracer := otel.GetTracerProvider().Tracer("obiara/worker")
	ctx, span := tracer.Start(ctx, "worker.run")
	span.SetAttributes(attribute.Int64("worker.jobs", jobs))
	counter, _ := runtime.meter.Int64Counter("worker.starts")
	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "ready")))
	span.End()
	return ctx
}

func (runtime *Runtime) Shutdown(ctx context.Context) error {
	var errs []error
	if runtime != nil && runtime.meters != nil {
		errs = append(errs, runtime.meters.Shutdown(ctx))
	}
	if runtime != nil && runtime.traces != nil {
		errs = append(errs, runtime.traces.Shutdown(ctx))
	}
	return errors.Join(errs...)
}
