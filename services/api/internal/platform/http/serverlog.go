package apihttp

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
)

// serverLogger is the sink for the causes behind 5xx responses.
//
// Error envelopes are deliberately opaque to callers: they carry a code and a
// safe message, never the underlying fault. Without a server-side sink that
// meant a 500 left no trace anywhere, and an operator seeing "The request
// could not be completed." had nothing to act on. This records the cause with
// the correlation id the caller was shown, so the two can be joined during
// triage.
var serverLogger atomic.Pointer[slog.Logger]

// SetServerLogger installs the logger used for 5xx causes. The composition
// root calls it with the telemetry runtime's redacting logger. Until then
// slog.Default() is used, so nothing is lost in tests.
func SetServerLogger(logger *slog.Logger) {
	if logger != nil {
		serverLogger.Store(logger)
	}
}

func logger() *slog.Logger {
	if stored := serverLogger.Load(); stored != nil {
		return stored
	}
	return slog.Default()
}

// maxFaultLength bounds what a single fault can write to the log.
const maxFaultLength = 300

var (
	// An unmapped error can wrap anything, including a driver message that
	// quotes the offending document. These patterns mask the identifiers
	// that would most commonly ride along.
	emailPattern = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)
	digitsRun    = regexp.MustCompile(`\+?\d{7,}`)
	longToken    = regexp.MustCompile(`\b[A-Za-z0-9_\-]{24,}\b`)
)

// sanitizeFault renders an error safe to log.
//
// The runtime logger redacts the "error" and "err" keys outright, because an
// arbitrary error chain can quote member data. Dropping the cause entirely
// left 5xx responses undiagnosable, so instead of bypassing that control
// this masks the identifiers that make errors unsafe — addresses, phone-like
// digit runs, and token-shaped strings — and bounds the length.
func sanitizeFault(cause error) string {
	text := cause.Error()
	text = emailPattern.ReplaceAllString(text, "[email]")
	text = digitsRun.ReplaceAllString(text, "[digits]")
	text = longToken.ReplaceAllString(text, "[token]")
	text = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, text)
	if len(text) > maxFaultLength {
		text = text[:maxFaultLength] + "…"
	}
	return text
}

// routePattern returns the matched ServeMux template, falling back to the
// method alone when a request failed before routing.
func routePattern(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	return r.Method
}

// logServerError records the cause of a 5xx response under the "fault" key,
// which carries a sanitized summary rather than the raw error chain.
func logServerError(ctx context.Context, r *http.Request, status int, code string, cause error) {
	if cause == nil {
		return
	}
	logger().ErrorContext(ctx, "request failed",
		slog.Int("status", status),
		slog.String("code", code),
		// r.Pattern is the matched route template ("POST /v1/admin/{id}"),
		// so it never carries the identifiers a raw path would.
		slog.String("route", routePattern(r)),
		slog.String("method", r.Method),
		slog.String("correlationId", CorrelationID(ctx)),
		slog.String("fault", sanitizeFault(cause)),
	)
}
