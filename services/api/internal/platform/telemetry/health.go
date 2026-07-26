package telemetry

import (
	"context"
	"log/slog"
	"time"
)

// InstrumentHealth wraps a dependency health check with bounded metrics and
// operational logging. Dependency names must be static configuration, not
// caller or member data.
func InstrumentHealth(
	dependency string,
	check func(context.Context) error,
	metrics Metrics,
	logger *slog.Logger,
) func(context.Context) error {
	name := safeOperation(dependency)
	if name == "" {
		name = "unknown"
	}
	if metrics == nil {
		metrics = NoopMetrics{}
	}
	return func(ctx context.Context) error {
		started := time.Now()
		err := check(ctx)
		status := "ready"
		if err != nil {
			status = "unavailable"
		}
		attributes := []Attribute{
			{Key: "dependency", Value: name},
			{Key: "status", Value: status},
		}
		metrics.Add(ctx, "health.checks", 1, attributes...)
		metrics.Observe(ctx, "health.check.duration", time.Since(started), attributes...)
		level := slog.LevelDebug
		if err != nil {
			level = slog.LevelWarn
		}
		LogAttrs(ctx, logger, level, "health.check",
			slog.String("dependency", name),
			slog.String("status", status),
		)
		return err
	}
}
