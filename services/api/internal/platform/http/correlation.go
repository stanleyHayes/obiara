package apihttp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	CorrelationIDHeader = "X-Correlation-ID"
	maxCorrelationIDLen = 128
)

type correlationIDKey struct{}

// Correlation adds a safe request correlation identifier to the context and
// response. A valid upstream identifier is preserved across service calls;
// malformed untrusted input is replaced rather than reflected.
func Correlation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get(CorrelationIDHeader))
		if !validCorrelationID(id) {
			id = newCorrelationID()
		}

		w.Header().Set(CorrelationIDHeader, id)
		ctx := context.WithValue(r.Context(), correlationIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CorrelationID retrieves the identifier established by Correlation.
func CorrelationID(ctx context.Context) string {
	id, _ := ctx.Value(correlationIDKey{}).(string)
	return id
}

func validCorrelationID(value string) bool {
	if len(value) < 8 || len(value) > maxCorrelationIDLen {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return false
		}
	}
	return true
}

func newCorrelationID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}
	// crypto/rand failures are exceptional. This value remains safe to expose
	// and searchable, while deliberately making no uniqueness claim.
	return "correlation-id-unavailable"
}
