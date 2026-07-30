package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/domain"
	admin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

type marketPacksStub struct {
	all []domain.MarketPack
}

func (stub marketPacksStub) All(context.Context, int) ([]domain.MarketPack, error) {
	return stub.all, nil
}
func (marketPacksStub) Published(context.Context) ([]domain.MarketPack, error) {
	return nil, nil
}
func (marketPacksStub) Draft(context.Context, domain.Market, string, map[string]bool, string) (domain.MarketPack, error) {
	panic("unexpected draft")
}
func (marketPacksStub) Publish(context.Context, string, string) (domain.MarketPack, error) {
	panic("unexpected publish")
}
func (marketPacksStub) Retire(context.Context, string, string) (domain.MarketPack, error) {
	panic("unexpected retire")
}

func TestAdminMarketPackRegisterProjectsNoActorIdentifiers(t *testing.T) {
	pack, err := domain.NewPack(
		"pack-1", domain.MarketGhanaTwi, "terminology:gh-tw:v4",
		map[string]bool{"fires": true}, "operator-secret", time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterMarketPackRoutes(mux, marketPacksStub{all: []domain.MarketPack{pack}}, func(*http.Request) (admin.Principal, error) {
		return admin.Principal{ActorID: "operator-secret", Scopes: []admin.Scope{admin.ScopeOperations}}, nil
	})
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/market-packs", nil)
	request.Header.Set("Authorization", "Bearer session")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if strings.Contains(body, "operator-secret") {
		t.Fatalf("actor identifier leaked: %s", body)
	}
	if !strings.Contains(body, `"proposedByMe":true`) || !strings.Contains(body, "terminology:gh-tw:v4") {
		t.Fatalf("safe governance projection missing: %s", body)
	}
}

func TestMarketPackMutationsRequireFreshMFA(t *testing.T) {
	for _, path := range []string{
		"/v1/admin/market-packs",
		"/v1/admin/market-packs/pack-1/publish",
		"/v1/admin/market-packs/pack-1/retire",
	} {
		t.Run(path, func(t *testing.T) {
			mux := http.NewServeMux()
			RegisterMarketPackRoutes(mux, marketPacksStub{}, func(*http.Request) (admin.Principal, error) {
				return admin.Principal{ActorID: "operator-1", Scopes: []admin.Scope{admin.ScopeOperations}}, nil
			})
			request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
			request.Header.Set("Authorization", "Bearer session")
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), "admin_step_up_required") {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}
