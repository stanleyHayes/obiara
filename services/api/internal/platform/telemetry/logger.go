package telemetry

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

const RedactedValue = "[REDACTED]"

var sensitiveKeyParts = []string{
	"address",
	"authorization",
	"biometric",
	"card",
	"content",
	"cookie",
	"credential",
	"date_of_birth",
	"dob",
	"email",
	"ghana",
	"key",
	"liveness",
	"message_body",
	"name",
	"otp",
	"password",
	"phone",
	"secret",
	"session",
	"token",
	"voice",
}

// LoggerConfig describes immutable deployment metadata safe for every record.
type LoggerConfig struct {
	Service     string
	Version     string
	Environment string
	Level       slog.Leveler
}

// NewJSONLogger writes one structured JSON object per record. All attributes,
// including attributes supplied through With, pass through deterministic
// sensitive-key redaction.
func NewJSONLogger(output io.Writer, config LoggerConfig) *slog.Logger {
	options := &slog.HandlerOptions{Level: config.Level}
	handler := newRedactingHandler(slog.NewJSONHandler(output, options))
	attrs := []any{slog.String("service", safeMetadata(config.Service))}
	if version := safeMetadata(config.Version); version != "" {
		attrs = append(attrs, slog.String("service_version", version))
	}
	if environment := safeMetadata(config.Environment); environment != "" {
		attrs = append(attrs, slog.String("environment", environment))
	}
	return slog.New(handler).With(attrs...)
}

// LogAttrs emits a context-enriched record through logger.
func LogAttrs(
	ctx context.Context,
	logger *slog.Logger,
	level slog.Level,
	event string,
	attrs ...slog.Attr,
) {
	if logger == nil {
		return
	}
	all := append(ContextAttrs(ctx), attrs...)
	logger.LogAttrs(ctx, level, safeEvent(event), all...)
}

type redactingHandler struct {
	next slog.Handler
}

func newRedactingHandler(next slog.Handler) slog.Handler {
	return redactingHandler{next: next}
}

func (handler redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.next.Enabled(ctx, level)
}

func (handler redactingHandler) Handle(ctx context.Context, record slog.Record) error {
	redacted := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		redacted.AddAttrs(redactAttr(attr))
		return true
	})
	return handler.next.Handle(ctx, redacted)
}

func (handler redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		redacted = append(redacted, redactAttr(attr))
	}
	return redactingHandler{next: handler.next.WithAttrs(redacted)}
}

func (handler redactingHandler) WithGroup(name string) slog.Handler {
	return redactingHandler{next: handler.next.WithGroup(name)}
}

func redactAttr(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	if sensitiveKey(attr.Key) {
		return slog.String(attr.Key, RedactedValue)
	}
	if attr.Value.Kind() != slog.KindGroup {
		return attr
	}
	children := attr.Value.Group()
	redacted := make([]slog.Attr, 0, len(children))
	for _, child := range children {
		redacted = append(redacted, redactAttr(child))
	}
	return slog.Group(attr.Key, attrsToAny(redacted)...)
}

func attrsToAny(attrs []slog.Attr) []any {
	values := make([]any, len(attrs))
	for index := range attrs {
		values[index] = attrs[index]
	}
	return values
}

func sensitiveKey(key string) bool {
	var normalizedBuilder strings.Builder
	for index, r := range key {
		if index > 0 && r >= 'A' && r <= 'Z' {
			normalizedBuilder.WriteByte('_')
		}
		normalizedBuilder.WriteRune(r)
	}
	normalized := strings.ToLower(strings.NewReplacer("-", "_", ".", "_").Replace(normalizedBuilder.String()))
	for _, part := range sensitiveKeyParts {
		if normalized == part ||
			strings.HasPrefix(normalized, part+"_") ||
			strings.HasSuffix(normalized, "_"+part) ||
			strings.Contains(normalized, "_"+part+"_") {
			return true
		}
	}
	return false
}

func safeEvent(event string) string {
	if event = safeOperation(event); event != "" {
		return event
	}
	return "telemetry.invalid_event"
}

func safeMetadata(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 96 {
		return ""
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}
	return value
}
