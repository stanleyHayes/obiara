package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type adminCatalogStub struct {
	put func(context.Context, domain.LicensedProfile, uint64, string) error
}

func (stub adminCatalogStub) ListAll(context.Context) ([]domain.LicensedProfile, error) {
	return nil, nil
}
func (stub adminCatalogStub) Put(ctx context.Context, profile domain.LicensedProfile, expected uint64, actor string) error {
	return stub.put(ctx, profile, expected, actor)
}

func TestAdminMatchmakerCreateRequiresMFAAndAuditedActor(t *testing.T) {
	payload := `{"licenseId":"license.ghana","jurisdiction":"ghana","expectedVersion":0,"validFrom":"2026-07-30T00:00:00Z","validUntil":"2027-07-30T00:00:00Z","minimumFeePesewas":8000,"maximumFeePesewas":25000,"displayName":"Akosua Mensah","languages":["Twi"],"specialties":["Consultation"],"completedEngagements":0,"ratingBasisPoints":0}`
	called := false
	catalog := adminCatalogStub{put: func(_ context.Context, profile domain.LicensedProfile, expected uint64, actor string) error {
		called = true
		if actor != "adm_licensing" || expected != 0 || profile.License.Version != 1 {
			t.Fatalf("actor=%q expected=%d profile=%+v", actor, expected, profile)
		}
		return nil
	}}
	mux := http.NewServeMux()
	RegisterAdminMatchmakerRoutes(mux, catalog, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{ActorID: "adm_licensing", Scopes: []admin.Scope{admin.ScopeOperations}}, nil
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/admin/matchmakers", strings.NewReader(payload))
	request.Header.Set("Authorization", "Bearer admin")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || called {
		t.Fatalf("without MFA status=%d called=%v body=%s", response.Code, called, response.Body.String())
	}

	mux = http.NewServeMux()
	RegisterAdminMatchmakerRoutes(mux, catalog, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{ActorID: "adm_licensing", Scopes: []admin.Scope{admin.ScopeOperations}, MFAVerified: true}, nil
	})
	request = httptest.NewRequest(http.MethodPost, "/v1/admin/matchmakers", strings.NewReader(payload))
	request.Header.Set("Authorization", "Bearer admin")
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusCreated || !called {
		t.Fatalf("with MFA status=%d called=%v body=%s", response.Code, called, response.Body.String())
	}
}
