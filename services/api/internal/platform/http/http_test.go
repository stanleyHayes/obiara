package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/member/application"
	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

func TestRegisterMemberSuccessEnvelope(t *testing.T) {
	createdAt := time.Date(2026, time.July, 26, 12, 30, 0, 0, time.UTC)
	var received application.RegisterMemberCommand
	handler := testHandler(func(_ context.Context, command application.RegisterMemberCommand) (domain.Member, error) {
		received = command
		return domain.NewMember(command.ID, command.Email, createdAt)
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/members", strings.NewReader(`{"id":"member-1","email":"MEMBER@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "register-1")
	request.Header.Set(CorrelationIDHeader, "request-1234")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusCreated, response.Body.String())
	}
	if got := response.Header().Get("Location"); got != "/v1/members/member-1" {
		t.Fatalf("Location = %q", got)
	}
	if got := response.Header().Get(CorrelationIDHeader); got != "request-1234" {
		t.Fatalf("correlation header = %q", got)
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if received.Email != "member@example.com" || received.IdempotencyKey != "register-1" {
		t.Fatalf("command = %#v", received)
	}

	var envelope struct {
		Data memberResponse `json:"data"`
		Meta metadata       `json:"meta"`
	}
	decodeResponse(t, response, &envelope)
	if envelope.Data.ID != "member-1" || envelope.Data.CreatedAt != createdAt {
		t.Fatalf("data = %#v", envelope.Data)
	}
	if envelope.Meta.CorrelationID != "request-1234" {
		t.Fatalf("meta correlationId = %q", envelope.Meta.CorrelationID)
	}
}

func TestRegisterMemberRejectsInvalidRequests(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		key         string
		status      int
		code        string
	}{
		{name: "content type", contentType: "text/plain", body: `{}`, key: "request-1", status: http.StatusUnsupportedMediaType, code: "unsupported_media_type"},
		{name: "malformed JSON", contentType: "application/json", body: `{"id":`, key: "request-1", status: http.StatusBadRequest, code: "invalid_json"},
		{name: "unknown field", contentType: "application/json", body: `{"id":"member-1","email":"a@example.com","role":"admin"}`, key: "request-1", status: http.StatusBadRequest, code: "invalid_json"},
		{name: "multiple objects", contentType: "application/json", body: `{"id":"member-1","email":"a@example.com"} {}`, key: "request-1", status: http.StatusBadRequest, code: "invalid_json"},
		{name: "invalid fields", contentType: "application/json", body: `{"id":"bad id","email":"not-email"}`, status: http.StatusUnprocessableEntity, code: "validation_failed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			handler := testHandler(func(context.Context, application.RegisterMemberCommand) (domain.Member, error) {
				called = true
				return domain.Member{}, nil
			})
			request := httptest.NewRequest(http.MethodPost, "/v1/members", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			request.Header.Set("Idempotency-Key", test.key)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.status, response.Body.String())
			}
			var envelope errorEnvelope
			decodeResponse(t, response, &envelope)
			if envelope.Error.Code != test.code {
				t.Fatalf("code = %q, want %q", envelope.Error.Code, test.code)
			}
			if called {
				t.Fatal("application handler called for invalid request")
			}
		})
	}
}

func TestRegisterMemberMapsApplicationFailureWithoutLeakingIt(t *testing.T) {
	handler := testHandler(func(context.Context, application.RegisterMemberCommand) (domain.Member, error) {
		return domain.Member{}, errors.New("mongo dial tcp secret.internal:27017")
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/members", strings.NewReader(`{"id":"member-1","email":"a@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "request-1")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "mongo") || strings.Contains(response.Body.String(), "secret.internal") {
		t.Fatalf("response leaked internal error: %s", response.Body.String())
	}
	var envelope errorEnvelope
	decodeResponse(t, response, &envelope)
	if envelope.Error.Code != "internal_error" {
		t.Fatalf("code = %q", envelope.Error.Code)
	}
}

func TestCorrelationReplacesUnsafeInput(t *testing.T) {
	handler := Correlation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeSuccess(w, r, http.StatusOK, map[string]string{"status": "ok"})
	}))
	request := httptest.NewRequest(http.MethodGet, "/live", nil)
	request.Header.Set(CorrelationIDHeader, "unsafe\r\nvalue")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	id := response.Header().Get(CorrelationIDHeader)
	if !validCorrelationID(id) || id == "unsafe\r\nvalue" {
		t.Fatalf("generated correlation ID = %q", id)
	}
	var envelope successEnvelope
	decodeResponse(t, response, &envelope)
	if envelope.Meta.CorrelationID != id {
		t.Fatalf("meta correlationId = %q, header = %q", envelope.Meta.CorrelationID, id)
	}
}

func TestRegisterMemberLimitsRequestBody(t *testing.T) {
	handler := testHandler(func(context.Context, application.RegisterMemberCommand) (domain.Member, error) {
		t.Fatal("application handler called")
		return domain.Member{}, nil
	})
	body := `{"id":"member-1","email":"` + strings.Repeat("a", maxRequestBodyBytes) + `@example.com"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/members", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "request-1")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func testHandler(register RegisterMember) http.Handler {
	mux := http.NewServeMux()
	RegisterMemberRoutes(mux, register)
	return Correlation(mux)
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if got := response.Header().Get("Content-Type"); got != contentTypeJSON {
		t.Fatalf("Content-Type = %q", got)
	}
	if err := json.Unmarshal(response.Body.Bytes(), destination); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, response.Body.String())
	}
}
