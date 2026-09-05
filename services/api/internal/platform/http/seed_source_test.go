package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	sourceapp "github.com/stanleyHayes/obiara/services/api/internal/seed/source/application"
	sourcedomain "github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

type sourceStub struct {
	open     func(context.Context, sourceapp.Command, sourceapp.Proposal) (sourceapp.Result, error)
	withdraw func(context.Context, sourceapp.Command) (sourceapp.Result, error)
	find     func(context.Context, string, string) (sourcedomain.Request, error)
}

func (s sourceStub) Open(ctx context.Context, c sourceapp.Command, p sourceapp.Proposal) (sourceapp.Result, error) {
	return s.open(ctx, c, p)
}
func (s sourceStub) Withdraw(ctx context.Context, c sourceapp.Command) (sourceapp.Result, error) {
	return s.withdraw(ctx, c)
}
func (s sourceStub) FindForRequester(ctx context.Context, id, requester string) (sourcedomain.Request, error) {
	return s.find(ctx, id, requester)
}

func sourceRequest(t *testing.T, candidates []string) sourcedomain.Request {
	t.Helper()
	at := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)
	key := strings.Repeat("a", 64)
	request, err := sourcedomain.Open(
		"src_1", key,
		sourcedomain.Source{Type: sourcedomain.SourceCircle, Key: strings.Repeat("b", 64)},
		candidates, at.Add(time.Hour),
		sourcedomain.Command{
			ID: "cmd_open", ActorKey: key, ReasonCode: "member_request",
			At: at, ExpectedRevision: 0,
		})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return request
}

func sourceMux(t *testing.T, stub sourceStub, memberID string) http.Handler {
	t.Helper()
	at := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)
	session := identitydomain.Reconstitute(
		"session-1", memberID, "device-1", identitydomain.StatusActive,
		"access", at.Add(time.Hour), "refresh", "", at.Add(24*time.Hour), 1, at, at,
	)
	mux := http.NewServeMux()
	RegisterSeedSourceRoutes(mux, stub,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}}, verifiedGate())
	return Correlation(mux)
}

func TestOpeningASourceReturnsACountAndNotTheCandidates(t *testing.T) {
	// Candidates are keyed before storage so who reached toward whom is not
	// legible at rest. Handing the keys back would undo exactly that: a caller
	// could correlate the same person across requests without learning a name.
	keys := []string{strings.Repeat("c", 64), strings.Repeat("d", 64)}
	stub := sourceStub{
		open: func(_ context.Context, command sourceapp.Command, proposal sourceapp.Proposal) (sourceapp.Result, error) {
			if proposal.RequesterID != "member-1" {
				t.Fatalf("requester = %q; the session must decide it", proposal.RequesterID)
			}
			if proposal.SourceType != sourcedomain.SourceCircle {
				t.Fatalf("source type = %q", proposal.SourceType)
			}
			if command.ActorID != "member-1" {
				t.Fatalf("actor = %q", command.ActorID)
			}
			return sourceapp.Result{Request: sourceRequest(t, keys)}, nil
		},
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sources",
		strings.NewReader(`{"circleId":"circle_1"}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "cmd-open-1")
	response := httptest.NewRecorder()
	sourceMux(t, stub, "member-1").ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, key := range keys {
		if strings.Contains(body, key) {
			t.Fatal("a candidate key was returned to the client")
		}
	}
	var envelope struct{ Data seedSourceResponse }
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Candidates != 2 {
		t.Fatalf("candidateCount = %d, want 2", envelope.Data.Candidates)
	}
}

func TestOpeningASourceNeedsAnIdempotencyKey(t *testing.T) {
	stub := sourceStub{open: func(context.Context, sourceapp.Command, sourceapp.Proposal) (sourceapp.Result, error) {
		t.Fatal("the service must not be reached without a command id")
		return sourceapp.Result{}, nil
	}}
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sources", strings.NewReader(`{"circleId":"c"}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	sourceMux(t, stub, "member-1").ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
}

func TestACircleTheMemberIsNotInReadsAsNotFound(t *testing.T) {
	// The authorizer refuses with ErrNotAvailable. Answering 403 would confirm
	// the circle is real, which is the disclosure this surface avoids.
	stub := sourceStub{open: func(context.Context, sourceapp.Command, sourceapp.Proposal) (sourceapp.Result, error) {
		return sourceapp.Result{}, sourceapp.ErrNotAvailable
	}}
	request := httptest.NewRequest(http.MethodPost, "/v1/seed/sources",
		strings.NewReader(`{"circleId":"someone_elses_circle"}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "cmd-1")
	response := httptest.NewRecorder()
	sourceMux(t, stub, "member-1").ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestAnotherMembersRequestReadsAsAbsent(t *testing.T) {
	stub := sourceStub{find: func(context.Context, string, string) (sourcedomain.Request, error) {
		return sourcedomain.Request{}, sourceapp.ErrNotFound
	}}
	request := httptest.NewRequest(http.MethodGet, "/v1/seed/sources/src_1", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	sourceMux(t, stub, "member-1").ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.Code)
	}
}

func TestWithdrawingCarriesTheMemberAsTheActor(t *testing.T) {
	stub := sourceStub{withdraw: func(_ context.Context, command sourceapp.Command) (sourceapp.Result, error) {
		if command.ActorID != "member-1" || command.RequestID != "src_1" {
			t.Fatalf("command = %+v", command)
		}
		return sourceapp.Result{Request: sourceRequest(t, nil)}, nil
	}}
	request := httptest.NewRequest(http.MethodDelete, "/v1/seed/sources/src_1", nil)
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Idempotency-Key", "cmd-withdraw-1")
	response := httptest.NewRecorder()
	sourceMux(t, stub, "member-1").ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
