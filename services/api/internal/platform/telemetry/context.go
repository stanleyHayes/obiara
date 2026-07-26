// Package telemetry provides privacy-safe, transport-neutral observability
// primitives for Obiara services.
package telemetry

import (
	"context"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

type contextKey struct{}

// ContextFields are the bounded identifiers that may be attached to logs,
// traces, and metrics. Member PII and raw content do not belong here.
type ContextFields struct {
	RequestID     string
	CorrelationID string
	Operation     string
}

// WithContextFields returns a context carrying normalized telemetry fields.
// Invalid control characters are discarded rather than reaching log output.
func WithContextFields(ctx context.Context, fields ContextFields) context.Context {
	fields.RequestID = safeIdentifier(fields.RequestID)
	fields.CorrelationID = safeIdentifier(fields.CorrelationID)
	fields.Operation = safeOperation(fields.Operation)
	return context.WithValue(ctx, contextKey{}, fields)
}

// FieldsFromContext returns the telemetry fields previously attached to ctx.
func FieldsFromContext(ctx context.Context) ContextFields {
	fields, _ := ctx.Value(contextKey{}).(ContextFields)
	return fields
}

// ContextAttrs returns stable log attributes, including valid OpenTelemetry
// trace and span identifiers when a recording or remote span is present.
func ContextAttrs(ctx context.Context) []slog.Attr {
	fields := FieldsFromContext(ctx)
	attrs := make([]slog.Attr, 0, 5)
	if fields.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", fields.RequestID))
	}
	if fields.CorrelationID != "" {
		attrs = append(attrs, slog.String("correlation_id", fields.CorrelationID))
	}
	if fields.Operation != "" {
		attrs = append(attrs, slog.String("operation", fields.Operation))
	}
	span := trace.SpanContextFromContext(ctx)
	if span.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", span.TraceID().String()),
			slog.String("span_id", span.SpanID().String()),
		)
	}
	return attrs
}

func safeIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 {
		return ""
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return ""
		}
	}
	return value
}

func safeOperation(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 96 {
		return ""
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '.', r == '_', r == '-', r == '/':
		default:
			return ""
		}
	}
	return value
}
