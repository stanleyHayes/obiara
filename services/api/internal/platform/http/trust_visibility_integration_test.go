//go:build integration

package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"

	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/trust"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestTrustVisibilityHTTPComposesMongoProjectionAndHidesRevocation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)
	container, err := testmongodb.Run(ctx, "mongo:8.0.13")
	if err != nil {
		t.Fatalf("start MongoDB Testcontainer (Docker/container runtime required): %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate MongoDB Testcontainer: %v", err)
		}
	})
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(uri, "?") {
		uri += "&directConnection=true"
	} else {
		uri += "?directConnection=true"
	}
	client, err := apimongo.Connect(ctx, uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	module, err := trust.NewModule(
		ctx,
		client.Database("obiara_trust_http_test"),
		integrationOwnerAuthorizer{},
		integrationTrustConsent{},
		integrationTrustEndpoints{},
	)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	edge, err := domain.Create(domain.Params{
		ID: "edge-ab", SourceID: "member-1", TargetID: "member-2", Type: domain.EdgeKnown,
		ProvenanceRef: "provenance-ab", ConsentRef: "consent-ab",
		Visibility: domain.VisibilityConsentedPath, CreatedAt: now,
	}, domain.Command{
		ID: "create-ab", ActorRef: "actor-1", Kind: "edge.create",
		Payload: "edge-ab", RecordedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := module.Edges.Save(ctx, edge, 0, "create-ab"); err != nil {
		t.Fatal(err)
	}
	session := trustSession(t, "member-1")
	handler := trustHTTPHandler(
		module.Visibility,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return session, nil
		}},
	)
	get := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/v1/members/member-1/trust-paths", nil)
		request.Header.Set("Authorization", "Bearer access")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	response := get()
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"targetId":"member-2"`) {
		t.Fatalf("visible status = %d, body = %s", response.Code, response.Body.String())
	}

	revoke := domain.Command{
		ID: "revoke-ab", ExpectedRevision: 1, ActorRef: "actor-1",
		Kind: "edge.revoke", Payload: "edge-ab", RecordedAt: now.Add(time.Minute),
	}
	revoked, err := edge.Revoke(revoke)
	if err != nil {
		t.Fatal(err)
	}
	if err := module.Edges.Save(ctx, revoked, 1, "revoke-ab"); err != nil {
		t.Fatal(err)
	}
	response = get()
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"paths":[]`) {
		t.Fatalf("revoked status = %d, body = %s", response.Code, response.Body.String())
	}
	for _, hidden := range []string{"edge-ab", "revoke", "consent-ab", "provenance-ab"} {
		if strings.Contains(response.Body.String(), hidden) {
			t.Fatalf("revocation detail %q leaked: %s", hidden, response.Body.String())
		}
	}
}

type integrationOwnerAuthorizer struct{}

func (integrationOwnerAuthorizer) CanProject(_ context.Context, requesterID, rootID string) (bool, error) {
	return requesterID == rootID, nil
}

type integrationTrustConsent struct{}

func (integrationTrustConsent) Allows(context.Context, string, string) (bool, error) {
	return true, nil
}

type integrationTrustEndpoints struct{}

func (integrationTrustEndpoints) CanReveal(context.Context, string, string) (bool, error) {
	return true, nil
}
