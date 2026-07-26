package visibility

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	trustapplication "github.com/stanleyHayes/obiara/services/api/internal/trust/application"
	"github.com/stanleyHayes/obiara/services/api/internal/trust/domain"
)

func TestHTTPContractReturnsBoundedPrivacySafeJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	requester := NewMockRequesterResolver(ctrl)
	edge := visibleEdge(t, "edge-ab", "member-a", "member-b", nil)
	service := NewService(
		staticProjector{projection: trustapplication.Projection{
			RootID: "member-a",
			Paths:  []trustapplication.Path{{Steps: []trustapplication.Step{stepFrom(edge)}}},
		}},
		NewDisclosurePolicy(mapEdgeReader{edge.ID(): edge}, allowAllConsent{}, endpointSet{
			"member-a": true, "member-b": true,
		}, func() time.Time { return visibilityTime().Add(time.Hour) }),
	)
	requester.EXPECT().RequesterID(gomock.Any()).Return("requester-1", nil)
	handler := NewHTTPHandler(service, requester)
	request := httptest.NewRequest(http.MethodGet, "/v1/members/member-a/trust-paths?depth=2&nodes=10", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	for _, expected := range []string{`"paths"`, `"targetId":"member-b"`, `"reason":"known_connection"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body %q missing %q", body, expected)
		}
	}
	if strings.Contains(body, "provenance-") || strings.Contains(body, "consent-") || strings.Contains(body, "edge-ab") {
		t.Fatalf("internal trust metadata leaked: %s", body)
	}
}

func TestHTTPContractCollapsesAuthorizationFailuresToNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	requester := NewMockRequesterResolver(ctrl)
	requester.EXPECT().RequesterID(gomock.Any()).Return("requester-1", nil)
	service := NewService(
		staticProjector{err: domain.ErrAccessDenied},
		NewDisclosurePolicy(nil, nil, nil, time.Now),
	)
	response := httptest.NewRecorder()
	NewHTTPHandler(service, requester).ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/v1/members/hidden-member/trust-paths", nil),
	)
	if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "hidden-member") {
		t.Fatalf("status = %d, body = %q", response.Code, response.Body.String())
	}
}
