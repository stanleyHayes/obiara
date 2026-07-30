package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	flagapp "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/application"
	flagdomain "github.com/stanleyHayes/obiara/services/api/internal/platform/flagcontrol/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

const controlActorKey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type controlServiceStub struct {
	propose func(context.Context, flagapp.ProposeCommand) (flagdomain.Proposal, error)
}

func (stub controlServiceStub) Propose(ctx context.Context, command flagapp.ProposeCommand) (flagdomain.Proposal, error) {
	return stub.propose(ctx, command)
}
func (controlServiceStub) Approve(context.Context, string, string) (flagdomain.Proposal, error) {
	return flagdomain.Proposal{}, nil
}
func (controlServiceStub) Apply(context.Context, string, string) (flagdomain.Proposal, error) {
	return flagdomain.Proposal{}, nil
}

type controlReaderStub struct{ proposals []flagdomain.Proposal }

func (stub controlReaderStub) ListActive(context.Context, int64) ([]flagdomain.Proposal, error) {
	return stub.proposals, nil
}

type controlKeyerStub struct{}

func (controlKeyerStub) Key(string, string) (string, error) { return controlActorKey, nil }

func controlProposal(t *testing.T) flagdomain.Proposal {
	t.Helper()
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	proposal, err := flagdomain.NewProposal(
		"proposal:1", "command:1", controlActorKey, flagdomain.CapabilityAI,
		flagdomain.EnvironmentStaging, flagdomain.MarketGH, flagdomain.ActionKill,
		flagdomain.ReasonIncident, now, now.Add(flagdomain.MaxLifetime),
	)
	if err != nil {
		t.Fatal(err)
	}
	return proposal
}

func TestAdminControlListRequiresOperationsAndProjectsNoActorKey(t *testing.T) {
	proposal := controlProposal(t)
	for name, testCase := range map[string]struct {
		principal admin.Principal
		want      int
	}{
		"forbidden": {admin.Principal{ActorID: "admin-1", Scopes: []admin.Scope{admin.ScopeSafety}}, http.StatusForbidden},
		"allowed":   {admin.Principal{ActorID: "admin-1", Scopes: []admin.Scope{admin.ScopeOperations}}, http.StatusOK},
	} {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			RegisterAdminControlRoutes(mux, controlServiceStub{}, controlReaderStub{[]flagdomain.Proposal{proposal}}, controlKeyerStub{}, func(*http.Request) (admin.Principal, error) {
				return testCase.principal, nil
			})
			request := httptest.NewRequest(http.MethodGet, "/v1/admin/controls", nil)
			request.Header.Set("Authorization", "Bearer session-1")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != testCase.want {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if response.Code == http.StatusOK && (strings.Contains(response.Body.String(), controlActorKey) || !strings.Contains(response.Body.String(), `"proposedByMe":true`)) {
				t.Fatalf("unsafe projection: %s", response.Body.String())
			}
		})
	}
}

func TestAdminControlProposalRequiresFreshMFA(t *testing.T) {
	called := false
	service := controlServiceStub{propose: func(context.Context, flagapp.ProposeCommand) (flagdomain.Proposal, error) {
		called = true
		return controlProposal(t), nil
	}}
	mux := http.NewServeMux()
	RegisterAdminControlRoutes(mux, service, controlReaderStub{}, controlKeyerStub{}, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{ActorID: "admin-1", Scopes: []admin.Scope{admin.ScopeOperations}}, nil
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/admin/controls", strings.NewReader(`{"commandId":"command:1","capability":"ai","environment":"staging","market":"GH","action":"kill","reason":"incident"}`))
	request.Header.Set("Authorization", "Bearer session-1")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || called || !strings.Contains(response.Body.String(), "admin_step_up_required") {
		t.Fatalf("status=%d called=%v body=%s", response.Code, called, response.Body.String())
	}
}
