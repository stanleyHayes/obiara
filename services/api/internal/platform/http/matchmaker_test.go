package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type matchmakerStub struct {
	profiles []domain.LicensedProfile
}

func (stub matchmakerStub) Marketplace(context.Context) ([]domain.LicensedProfile, error) {
	return stub.profiles, nil
}
func (matchmakerStub) ForMember(context.Context, string) ([]domain.Engagement, error) {
	return nil, nil
}
func (matchmakerStub) FindForMember(context.Context, string, string) (domain.Engagement, error) {
	return domain.Engagement{}, nil
}
func (matchmakerStub) Book(context.Context, string, string, domain.Terms, string) (domain.Engagement, error) {
	return domain.Engagement{}, nil
}
func (matchmakerStub) Mutate(context.Context, string, string, func(domain.Engagement) (domain.Engagement, error)) (domain.Engagement, error) {
	return domain.Engagement{}, nil
}

type matchmakerKeyerStub struct{}

func (matchmakerKeyerStub) MemberKey(string) (string, error) {
	return "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil
}

func TestMatchmakerMarketplaceProjectsOnlyBoundedLicenseData(t *testing.T) {
	now := time.Now().UTC()
	profile := domain.LicensedProfile{
		License: domain.License{
			ID:            "license.ghana",
			MatchmakerKey: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Jurisdiction:  "ghana", Version: 2,
			ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour),
			MinimumFeePesewas: 8000, MaximumFeePesewas: 25000,
		},
		DisplayName: "Akosua Mensah", Languages: []string{"Twi"},
		Specialties: []string{"Consultation"}, RatingBasisPoints: 480,
	}
	mux := http.NewServeMux()
	RegisterMatchmakerRoutes(mux, matchmakerStub{profiles: []domain.LicensedProfile{profile}}, matchmakerKeyerStub{}, sessionAuthenticatorStub{
		authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return trustSession(t, "mem_matchmaker"), nil
		},
	})
	request := httptest.NewRequest(http.MethodGet, "/v1/matchmakers", nil)
	request.Header.Set("Authorization", "Bearer member-session")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data struct {
			Items []matchmakerProfileResponse `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data.Items) != 1 || envelope.Data.Items[0].DisplayName != profile.DisplayName {
		t.Fatalf("items=%+v", envelope.Data.Items)
	}
	if envelope.Data.Items[0].MinimumFeePesewas != 8000 || envelope.Data.Items[0].RatingBasisPoints != 480 {
		t.Fatalf("bounded projection=%+v", envelope.Data.Items[0])
	}
}
