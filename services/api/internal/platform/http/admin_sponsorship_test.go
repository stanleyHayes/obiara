package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/purchase"
	sponsorshipapp "github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/application"
	sponsorshipdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/sponsorship/domain"
)

type fundsStub struct {
	fund      sponsorshipdomain.Fund
	list      []sponsorshipdomain.Fund
	err       error
	deposited sponsorshipapp.DepositCommand
	calls     int
}

func (s *fundsStub) Deposit(
	_ context.Context, command sponsorshipapp.DepositCommand,
) (sponsorshipdomain.Fund, error) {
	s.calls++
	s.deposited = command
	return s.fund, s.err
}

func (s *fundsStub) FindByOrganization(
	context.Context, string,
) (sponsorshipdomain.Fund, error) {
	s.calls++
	return s.fund, s.err
}

func (s *fundsStub) List(context.Context, int) ([]sponsorshipdomain.Fund, error) {
	s.calls++
	return s.list, s.err
}

// hashKeyer stands in for the operator keyer.
type hashKeyer struct{}

func (hashKeyer) Key(namespace, value string) (string, error) {
	return strings.Repeat("c", 64), nil
}

func fundFixture(t *testing.T, deposited, drawn int64) sponsorshipdomain.Fund {
	t.Helper()
	at := time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)
	fund, err := sponsorshipdomain.Open("fund_1", "org_1", sponsorshipdomain.Command{
		ID: "cmd_1", ActorKey: strings.Repeat("a", 64),
		ReasonCode: "partner_agreement", At: at,
	})
	if err != nil {
		t.Fatal(err)
	}
	if deposited > 0 {
		fund, err = fund.Deposit(deposited, sponsorshipdomain.Command{
			ID: "cmd_2", ActorKey: strings.Repeat("a", 64),
			ReasonCode: "bank_transfer_received", At: at,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if drawn > 0 {
		fund, err = fund.Draw("purchase_1", drawn, sponsorshipdomain.Command{
			ID: "cmd_3", ActorKey: strings.Repeat("a", 64),
			ReasonCode: "seat_taken", At: at,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return fund
}

func sponsorshipRequest(
	t *testing.T, stub *fundsStub, resolve AdminPrincipalResolver,
	method, path, body, key string,
) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterAdminSponsorshipRoutes(mux, stub, hashKeyer{}, resolve)
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

func TestRecordingADepositKeysTheOperatorAndCarriesTheReason(t *testing.T) {
	// Recording money that did not arrive would let seats be given away, so
	// there is a name against every deposit — a digest, like everywhere else.
	stub := &fundsStub{fund: fundFixture(t, 100_000, 0)}
	response := sponsorshipRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/sponsorship/deposits",
		`{"amountPesewas":100000,"reasonCode":"bank_transfer_received"}`, "cmd-1")

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.deposited.OrganizationID != "org_1" || stub.deposited.AmountPesewas != 100_000 {
		t.Fatalf("recorded %#v", stub.deposited)
	}
	if len(stub.deposited.OperatorKey) != 64 {
		t.Fatalf("operator recorded as %q", stub.deposited.OperatorKey)
	}
	if stub.deposited.ReasonCode != "bank_transfer_received" {
		t.Fatalf("reason = %q", stub.deposited.ReasonCode)
	}
}

func TestADepositNeedsAStepUp(t *testing.T) {
	stub := &fundsStub{}
	response := sponsorshipRequest(t, stub, withoutStepUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/sponsorship/deposits",
		`{"amountPesewas":100000,"reasonCode":"bank_transfer_received"}`, "cmd-1")

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
	if stub.calls != 0 {
		t.Fatal("money was recorded without a step-up")
	}
}

func TestADepositWithoutARetryKeyIsRefused(t *testing.T) {
	// An operator whose connection dropped will record the same transfer
	// again. Without a key that doubles an organization's balance.
	stub := &fundsStub{}
	response := sponsorshipRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/sponsorship/deposits",
		`{"amountPesewas":100000,"reasonCode":"bank_transfer_received"}`, "")

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
	if stub.calls != 0 {
		t.Fatal("a deposit with no retry key reached the service")
	}
}

func TestAFundSaysHowManySeatsAndNeverWho(t *testing.T) {
	// An organization is told a count. Naming the members who used its seats
	// would hand an outside party a list of people.
	stub := &fundsStub{fund: fundFixture(t, 100_000, 5_000)}
	response := sponsorshipRequest(t, stub, withoutStepUp("adm_1"), http.MethodGet,
		"/v1/admin/organizations/org_1/sponsorship", "", "")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"seats":1`) {
		t.Fatalf("the count is missing or wrong: %s", body)
	}
	if !strings.Contains(body, `"balancePesewas":95000`) {
		t.Fatalf("balance is wrong: %s", body)
	}
	for _, leak := range []string{`"drawnSeats"`, `"memberId"`, `"memberKey"`, "purchase_1"} {
		if strings.Contains(body, leak) {
			t.Fatalf("the fund named who took a seat (%s): %s", leak, body)
		}
	}
}

func TestFundingASuspendedOrganizationIsRefusedPlainly(t *testing.T) {
	stub := &fundsStub{err: sponsorshipapp.ErrIssuerNotIssuing}
	response := sponsorshipRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/sponsorship/deposits",
		`{"amountPesewas":100000,"reasonCode":"bank_transfer_received"}`, "cmd-1")

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "organization_not_issuing") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAnUnavailableSponsorshipTellsTheMemberTheyCanStillBuy(t *testing.T) {
	// The sponsorship is refused, not the purchase. The message says so
	// rather than reading as the member's account being at fault.
	stub := &purchaseStub{startErr: purchase.ErrSponsorshipUnavailable}
	response := purchaseCall(t, stub, "/v1/membership/purchases",
		`{"skuId":"sku","skuVersion":1,"phone":"0200000000","network":"mtn","code":"SEATS"}`,
		"cmd-1")

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	body := strings.ToLower(response.Body.String())
	if !strings.Contains(body, "sponsorship_unavailable") {
		t.Fatalf("body = %s", response.Body.String())
	}
	if !strings.Contains(body, "still buy") {
		t.Fatalf("the member was not told they can still buy: %s", response.Body.String())
	}
}
