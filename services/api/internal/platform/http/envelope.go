// Package apihttp owns the public HTTP transport contract for the API.
//
// Domain and persistence types must not cross this boundary. Handlers map
// application results into the stable JSON envelope described by ADR-0004.
package apihttp

import (
	"encoding/json"
	"net/http"
)

const contentTypeJSON = "application/json; charset=utf-8"

type metadata struct {
	CorrelationID string `json:"correlationId"`
}

type successEnvelope struct {
	Data any      `json:"data"`
	Meta metadata `json:"meta"`
}

type errorEnvelope struct {
	Error APIError `json:"error"`
	Meta  metadata `json:"meta"`
}

// APIError is the stable, client-safe error representation. Code is intended
// for programmatic handling; Message is safe presentation fallback text.
type APIError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError identifies invalid request input without echoing its value.
type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func writeSuccess(w http.ResponseWriter, r *http.Request, status int, data any) {
	writeJSON(w, status, successEnvelope{
		Data: data,
		Meta: metadata{CorrelationID: CorrelationID(r.Context())},
	})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, apiError APIError) {
	writeJSON(w, status, errorEnvelope{
		Error: apiError,
		Meta:  metadata{CorrelationID: CorrelationID(r.Context())},
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	// All envelope values are composed from controlled transport values. If
	// encoding ever fails, the connection has already received its status and
	// there is no safer secondary response to write.
	_ = json.NewEncoder(w).Encode(value)
}
