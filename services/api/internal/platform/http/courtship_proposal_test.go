package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/application"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/proposal/domain"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type proposalsStub struct {
	created  application.CreateCommand
	decided  application.DecisionCommand
	result   application.Result
	err      error
	lastCall string
}

func (stub *proposalsStub) Create(_ context.Context, c application.CreateCommand) (application.Result, error) {
	stub.created, stub.lastCall = c, "create"
	return stub.result, stub.err
}
func (stub *proposalsStub) Accept(_ context.Context, c application.DecisionCommand) (application.Result, error) {
	stub.decided, stub.lastCall = c, "accept"
	return stub.result, stub.err
}
func (stub *proposalsStub) Reject(_ context.Context, c application.DecisionCommand) (application.Result, error) {
	stub.decided, stub.lastCall = c, "reject"
	return stub.result, stub.err
}
func (stub *proposalsStub) Withdraw(_ context.Context, c application.DecisionCommand) (application.Result, error) {
	stub.decided, stub.lastCall = c, "withdraw"
	return stub.result, stub.err
}

type sessionStub struct{ memberID string }

func (stub sessionStub) Authenticate(context.Context, string) (identitydomain.Session, error) {
	if stub.memberID == "" {
		return identitydomain.Session{}, context.Canceled
	}
	return identitydomain.Reconstitute(
		"ses_1", stub.memberID, "device_1", identitydomain.StatusActive,
		"", time.Now().Add(time.Hour), "", "", time.Now().Add(24*time.Hour),
		1, time.Now(), time.Now(),
	), nil
}

func proposalHandler(proposals Proposals, memberID string) http.Handler {
	mux := http.NewServeMux()
	RegisterCourtshipProposalRoutes(mux, proposals, sessionStub{memberID: memberID})
	return Correlation(mux)
}

func postProposal(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func validCreateBody() string {
	return `{"commandId":"cmd_1","recipientId":"mem_other","kind":"call","detail":"a call this evening","expiresAt":"` +
		time.Now().Add(24*time.Hour).UTC().Format(time.RFC3339) + `"}`
}

func TestCreateProposalUsesTheSessionAsSender(t *testing.T) {
	stub := &proposalsStub{}
	response := postProposal(t, proposalHandler(stub, "mem_sender"), "/v1/courtship/proposals", validCreateBody())

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	// The sender must come from the session; taking it from the body would
	// let a member send proposals as somebody else.
	if stub.created.SenderID != "mem_sender" {
		t.Errorf("SenderID = %q, want the session's member", stub.created.SenderID)
	}
	if stub.created.RecipientID != "mem_other" || stub.created.Kind != domain.TypeCall {
		t.Errorf("command = %+v", stub.created)
	}
}

func TestCreateProposalRejectsSelfProposal(t *testing.T) {
	stub := &proposalsStub{}
	body := strings.Replace(validCreateBody(), "mem_other", "mem_sender", 1)
	response := postProposal(t, proposalHandler(stub, "mem_sender"), "/v1/courtship/proposals", body)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
	if stub.lastCall != "" {
		t.Error("a self-proposal reached the service")
	}
}

func TestCreateProposalValidates(t *testing.T) {
	cases := map[string]string{
		"no command id":   `{"commandId":"","recipientId":"mem_o","kind":"call","detail":"x","expiresAt":"2099-01-01T00:00:00Z"}`,
		"unknown kind":    `{"commandId":"c","recipientId":"mem_o","kind":"marriage","detail":"x","expiresAt":"2099-01-01T00:00:00Z"}`,
		"empty detail":    `{"commandId":"c","recipientId":"mem_o","kind":"call","detail":"","expiresAt":"2099-01-01T00:00:00Z"}`,
		"bad expiry":      `{"commandId":"c","recipientId":"mem_o","kind":"call","detail":"x","expiresAt":"tomorrow"}`,
		"detail too long": `{"commandId":"c","recipientId":"mem_o","kind":"call","detail":"` + strings.Repeat("x", 501) + `","expiresAt":"2099-01-01T00:00:00Z"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			stub := &proposalsStub{}
			response := postProposal(t, proposalHandler(stub, "mem_sender"), "/v1/courtship/proposals", body)
			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", response.Code)
			}
			if stub.lastCall != "" {
				t.Error("an invalid request reached the service")
			}
		})
	}
}

func TestDecisionsCarryTheActorAndRevision(t *testing.T) {
	for _, action := range []string{"accept", "reject", "withdraw"} {
		t.Run(action, func(t *testing.T) {
			stub := &proposalsStub{}
			response := postProposal(t, proposalHandler(stub, "mem_actor"),
				"/v1/courtship/proposals/prop_1/"+action, `{"commandId":"cmd_2","expectedRevision":3}`)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
			}
			if stub.lastCall != action {
				t.Errorf("called %q, want %q", stub.lastCall, action)
			}
			if stub.decided.ActorID != "mem_actor" || stub.decided.ProposalID != "prop_1" {
				t.Errorf("command = %+v", stub.decided)
			}
			// Without the expected revision two devices could race a decision.
			if stub.decided.ExpectedRevision != 3 {
				t.Errorf("ExpectedRevision = %d, want 3", stub.decided.ExpectedRevision)
			}
		})
	}
}

// TestUnknownAndForeignProposalsAreIndistinguishable stops the endpoint from
// becoming a way to enumerate other members' private negotiations.
func TestUnknownAndForeignProposalsAreIndistinguishable(t *testing.T) {
	var payloads []string
	for _, cause := range []error{application.ErrNotAvailable, domain.ErrNotParticipant} {
		stub := &proposalsStub{err: cause}
		response := postProposal(t, proposalHandler(stub, "mem_actor"),
			"/v1/courtship/proposals/prop_1/accept", `{"commandId":"cmd_2"}`)
		if response.Code != http.StatusNotFound {
			t.Fatalf("status = %d for %v, want 404", response.Code, cause)
		}
		// The correlation id is unique per request by design, so only the
		// error itself is compared.
		var envelope struct {
			Error APIError `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode: %v", err)
		}
		payloads = append(payloads, envelope.Error.Code+"|"+envelope.Error.Message)
	}
	if payloads[0] != payloads[1] {
		t.Errorf("responses differ (%q vs %q), letting a caller tell a missing proposal from someone else's",
			payloads[0], payloads[1])
	}
}

func TestProposalConflictsAreMapped(t *testing.T) {
	cases := map[error]int{
		domain.ErrTerminal:              http.StatusConflict,
		domain.ErrExpired:               http.StatusConflict,
		domain.ErrStaleRevision:         http.StatusConflict,
		application.ErrConcurrentChange: http.StatusConflict,
		domain.ErrCommandMismatch:       http.StatusConflict,
		domain.ErrInvalid:               http.StatusUnprocessableEntity,
		application.ErrUnavailable:      http.StatusServiceUnavailable,
	}
	for cause, want := range cases {
		stub := &proposalsStub{err: cause}
		response := postProposal(t, proposalHandler(stub, "mem_actor"),
			"/v1/courtship/proposals/prop_1/accept", `{"commandId":"cmd_2"}`)
		if response.Code != want {
			t.Errorf("%v mapped to %d, want %d", cause, response.Code, want)
		}
	}
}

// TestReplayedCommandIsNotACreate keeps a retried tap on a flaky network from
// reading as a second proposal.
func TestReplayedCommandReportsReplay(t *testing.T) {
	stub := &proposalsStub{result: application.Result{Replayed: true}}
	response := postProposal(t, proposalHandler(stub, "mem_sender"), "/v1/courtship/proposals", validCreateBody())

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for a replay", response.Code)
	}
	var envelope struct {
		Data proposalResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Data.Replayed {
		t.Error("response does not report the replay")
	}
}

func TestProposalRoutesRequireASession(t *testing.T) {
	stub := &proposalsStub{}
	handler := proposalHandler(stub, "")
	for _, path := range []string{
		"/v1/courtship/proposals",
		"/v1/courtship/proposals/prop_1/accept",
	} {
		response := postProposal(t, handler, path, `{"commandId":"c"}`)
		if response.Code != http.StatusUnauthorized {
			t.Errorf("%s status = %d, want 401", path, response.Code)
		}
	}
}
