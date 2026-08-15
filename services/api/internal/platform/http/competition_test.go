package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	competitionapp "github.com/stanleyHayes/obiara/services/api/internal/games/competition/application"
	competitiondomain "github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type competitionReviewDeskStub struct {
	resolveCalls int
	appealCalls  int
}

func (competitionReviewDeskStub) ViewForReviewer(context.Context, competitionapp.Command) (competitionapp.PrivateProjection, error) {
	return competitionapp.PrivateProjection{}, nil
}

func (stub *competitionReviewDeskStub) ResolveReviewPrivate(context.Context, competitionapp.Command, string, competitiondomain.Decision) (competitionapp.PrivateProjection, error) {
	stub.resolveCalls++
	return competitionapp.PrivateProjection{}, nil
}

func (stub *competitionReviewDeskStub) ResolveAppealPrivate(context.Context, competitionapp.Command, string, competitiondomain.Decision) (competitionapp.PrivateProjection, error) {
	stub.appealCalls++
	return competitionapp.PrivateProjection{}, nil
}

func TestCompetitionReviewResolutionRequiresSteppedUpOperationsAdmin(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		principal admin.Principal
		want      int
		calls     int
	}{
		{
			"resolve rejects scope-only principal",
			"/v1/admin/game-cohorts/coh_1/competitions/comp_1/reviews/rev_1/resolve",
			admin.Principal{ActorID: "ops", Scopes: []admin.Scope{admin.ScopeOperations}},
			http.StatusForbidden, 0,
		},
		{
			"resolve-appeal rejects scope-only principal",
			"/v1/admin/game-cohorts/coh_1/competitions/comp_1/reviews/rev_1/resolve-appeal",
			admin.Principal{ActorID: "ops", Scopes: []admin.Scope{admin.ScopeOperations}},
			http.StatusForbidden, 0,
		},
		{
			"stepped-up operations admin resolves",
			"/v1/admin/game-cohorts/coh_1/competitions/comp_1/reviews/rev_1/resolve",
			admin.Principal{ActorID: "ops", Scopes: []admin.Scope{admin.ScopeOperations}, MFAVerified: true},
			http.StatusOK, 1,
		},
		{
			"stepped-up operations admin resolves an appeal",
			"/v1/admin/game-cohorts/coh_1/competitions/comp_1/reviews/rev_1/resolve-appeal",
			admin.Principal{ActorID: "ops", Scopes: []admin.Scope{admin.ScopeOperations}, MFAVerified: true},
			http.StatusOK, 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stub := &competitionReviewDeskStub{}
			resolve := func(*http.Request) (admin.Principal, error) {
				return test.principal, nil
			}
			mux := http.NewServeMux()
			mux.Handle("POST /v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve", resolveCompetitionReviewHandler(stub, resolve, false))
			mux.Handle("POST /v1/admin/game-cohorts/{cohortId}/competitions/{competitionId}/reviews/{reviewId}/resolve-appeal", resolveCompetitionReviewHandler(stub, resolve, true))
			request := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(`{"decision":"rules_action","expectedRevision":0}`))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Idempotency-Key", "resolve-1")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			calls := stub.resolveCalls + stub.appealCalls
			if response.Code != test.want || calls != test.calls {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, calls, response.Body.String())
			}
		})
	}
}
