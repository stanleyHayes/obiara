package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	promotionapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/application"
	promotiondomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/domain"
)

type promotionStub struct {
	promotion promotiondomain.Promotion
	list      []promotiondomain.Promotion
	err       error
	issued    promotionapp.IssueCommand
	calls     int
}

func (s *promotionStub) Issue(
	_ context.Context, command promotionapp.IssueCommand,
) (promotiondomain.Promotion, error) {
	s.calls++
	s.issued = command
	return s.promotion, s.err
}

func (s *promotionStub) Withdraw(context.Context, string, string) (promotiondomain.Promotion, error) {
	s.calls++
	return s.promotion, s.err
}

func (s *promotionStub) ListByIssuer(context.Context, string, int) ([]promotiondomain.Promotion, error) {
	s.calls++
	return s.list, s.err
}

func promotionFixture(t *testing.T) promotiondomain.Promotion {
	t.Helper()
	opened := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	promotion, err := promotiondomain.Issue(
		"promo_1", "ASHESI26", "org_1", "sku_membership",
		promotiondomain.ShapePercentage, 50, opened, opened.AddDate(0, 1, 0), 50,
		promotiondomain.Command{ID: "cmd_1", At: opened})
	if err != nil {
		t.Fatal(err)
	}
	return promotion
}

func promotionRequest(
	t *testing.T, stub *promotionStub, resolve AdminPrincipalResolver,
	method, path, body, key string,
) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterAdminPromotionRoutes(mux, stub, resolve)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	return response
}

func TestIssuingACodeNamesTheOrganizationFromThePath(t *testing.T) {
	stub := &promotionStub{promotion: promotionFixture(t)}
	response := promotionRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/promotions",
		`{"code":"ashesi26","skuId":"sku_membership","shape":"percentage","amount":50,`+
			`"startsAt":"2026-09-01T00:00:00Z","endsAt":"2026-10-01T00:00:00Z","cap":50}`,
		"cmd-1")

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.issued.IssuerID != "org_1" {
		t.Fatalf("issuer = %q", stub.issued.IssuerID)
	}
	if stub.issued.Cap != 50 || stub.issued.Amount != 50 {
		t.Fatalf("issued %#v", stub.issued)
	}
}

func TestIssuingACodeNeedsAStepUp(t *testing.T) {
	// A code is money coming off real revenue.
	stub := &promotionStub{}
	response := promotionRequest(t, stub, withoutStepUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/promotions",
		`{"code":"ASHESI26","skuId":"sku","shape":"fixed","amount":100,`+
			`"startsAt":"2026-09-01T00:00:00Z","endsAt":"2026-10-01T00:00:00Z","cap":5}`,
		"cmd-1")

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if stub.calls != 0 {
		t.Fatal("a code was issued without a step-up")
	}
}

func TestACodeForASuspendedOrganizationIsRefusedPlainly(t *testing.T) {
	stub := &promotionStub{err: promotionapp.ErrIssuerNotIssuing}
	response := promotionRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/promotions",
		`{"code":"ASHESI26","skuId":"sku","shape":"fixed","amount":100,`+
			`"startsAt":"2026-09-01T00:00:00Z","endsAt":"2026-10-01T00:00:00Z","cap":5}`,
		"cmd-1")

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "organization_not_issuing") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestTheCodeListSaysHowManyAndNeverWho(t *testing.T) {
	// Counting is the whole reporting story an organization gets.
	stub := &promotionStub{list: []promotiondomain.Promotion{promotionFixture(t)}}
	response := promotionRequest(t, stub, withoutStepUp("adm_1"), http.MethodGet,
		"/v1/admin/organizations/org_1/promotions", "", "")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"redeemed"`) {
		t.Fatalf("the count is missing: %s", body)
	}
	// No member appears anywhere in the projection. Matched on the field
	// names that would carry one rather than on a substring: "members" also
	// occurs inside sku_membership, which is not a leak.
	for _, leak := range []string{`"redeemedKeys"`, `"memberKey"`, `"memberId"`, `"redeemedBy"`} {
		if strings.Contains(body, leak) {
			t.Fatalf("the list named who used a code (%s): %s", leak, body)
		}
	}
}

func TestABadTimestampIsRefusedAsInputRatherThanAsAFault(t *testing.T) {
	stub := &promotionStub{}
	response := promotionRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/promotions",
		`{"code":"ASHESI26","skuId":"sku","shape":"fixed","amount":100,`+
			`"startsAt":"next tuesday","endsAt":"2026-10-01T00:00:00Z","cap":5}`,
		"cmd-1")

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if stub.calls != 0 {
		t.Fatal("an unparseable window reached the service")
	}
}
