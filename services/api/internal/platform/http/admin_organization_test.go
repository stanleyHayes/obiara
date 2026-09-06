package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	organizationapp "github.com/stanleyHayes/obiara/services/api/internal/organization/application"
	organizationdomain "github.com/stanleyHayes/obiara/services/api/internal/organization/domain"
	verificationadmin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

// organizationStub records what the surface asked for.
type organizationStub struct {
	result     organizationapp.Result
	list       []organizationdomain.Organization
	err        error
	registered organizationapp.RegisterCommand
	changed    organizationapp.ChangeCommand
	calls      int
}

func (s *organizationStub) Register(
	_ context.Context, command organizationapp.RegisterCommand,
) (organizationapp.Result, error) {
	s.calls++
	s.registered = command
	return s.result, s.err
}

func (s *organizationStub) Suspend(
	_ context.Context, command organizationapp.ChangeCommand,
) (organizationapp.Result, error) {
	s.calls++
	s.changed = command
	return s.result, s.err
}

func (s *organizationStub) Restore(
	_ context.Context, command organizationapp.ChangeCommand,
) (organizationapp.Result, error) {
	s.calls++
	s.changed = command
	return s.result, s.err
}

func (s *organizationStub) Rename(
	_ context.Context, command organizationapp.ChangeCommand,
) (organizationapp.Result, error) {
	s.calls++
	s.changed = command
	return s.result, s.err
}

func (s *organizationStub) List(context.Context, int) ([]organizationdomain.Organization, error) {
	s.calls++
	return s.list, s.err
}

// steppedUp is an operator who has re-asserted who they are. withoutStepUp is
// the same operator who has not.
func steppedUp(actorID string) AdminPrincipalResolver {
	return func(*http.Request) (verificationadmin.Principal, error) {
		return verificationadmin.Principal{
			ActorID: actorID, Scopes: []verificationadmin.Scope{adminOperationsScope},
			MFAVerified: true,
		}, nil
	}
}

func withoutStepUp(actorID string) AdminPrincipalResolver {
	return func(*http.Request) (verificationadmin.Principal, error) {
		return verificationadmin.Principal{
			ActorID: actorID, Scopes: []verificationadmin.Scope{adminOperationsScope},
		}, nil
	}
}

func organizationFixture(t *testing.T) organizationdomain.Organization {
	t.Helper()
	organization, err := organizationdomain.Register(
		"org_1", "Ashesi University", "billing@ashesi.edu.gh",
		organizationdomain.Command{
			ID: "cmd_1", ActorKey: strings.Repeat("a", 64), ReasonCode: "operator_request",
			At: time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC),
		})
	if err != nil {
		t.Fatal(err)
	}
	return organization
}

func organizationRequest(
	t *testing.T, stub *organizationStub, resolve AdminPrincipalResolver,
	method, path, body, key string,
) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	RegisterAdminOrganizationRoutes(mux, stub, resolve)
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
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

func TestRegisteringAnOrganizationPassesTheOperatorAndTheReason(t *testing.T) {
	// The audit trail is the point of this record. What the surface passes on
	// is what ends up in it.
	stub := &organizationStub{result: organizationapp.Result{Organization: organizationFixture(t)}}
	response := organizationRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations",
		`{"name":"Ashesi University","billingEmail":"billing@ashesi.edu.gh","reasonCode":"partner_agreement"}`,
		"cmd_1")

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", response.Code, response.Body.String())
	}
	if stub.registered.OperatorID != "adm_1" || stub.registered.ReasonCode != "partner_agreement" {
		t.Fatalf("recorded %#v", stub.registered)
	}
	if stub.registered.CommandID != "cmd_1" {
		t.Fatalf("command id = %q, want the idempotency key", stub.registered.CommandID)
	}
}

func TestAnOperatorWhoHasNotSteppedUpRegistersNothing(t *testing.T) {
	// A commercial relationship recorded by somebody who had not re-asserted
	// who they were is an audit entry worth less than no entry.
	stub := &organizationStub{}
	response := organizationRequest(t, stub, withoutStepUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations",
		`{"name":"Ashesi","billingEmail":"b@ashesi.edu.gh","reasonCode":"partner_agreement"}`,
		"cmd_1")

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", response.Code, response.Body.String())
	}
	if stub.calls != 0 {
		t.Fatal("the request reached the service without a step-up")
	}
}

func TestReadingTheRosterNeedsNoStepUp(t *testing.T) {
	// It holds no member data at all — an organization is a body, not a
	// person. Requiring a fresh MFA to look at a list of universities would
	// train operators to step up out of habit, which is what makes step-up
	// mean nothing.
	stub := &organizationStub{list: []organizationdomain.Organization{organizationFixture(t)}}
	response := organizationRequest(t, stub, withoutStepUp("adm_1"), http.MethodGet,
		"/v1/admin/organizations", "", "")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Ashesi University") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAChangeWithoutARetryKeyIsRefused(t *testing.T) {
	// The key is what makes a retried suspension one suspension. Generating
	// one here would put a second entry in the trail for one decision.
	stub := &organizationStub{}
	response := organizationRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/suspend", `{"reasonCode":"unpaid_invoice","expectedRevision":1}`, "")

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	if stub.calls != 0 {
		t.Fatal("a change with no retry key reached the service")
	}
}

func TestSuspendingCarriesTheRevisionTheOperatorDecidedAgainst(t *testing.T) {
	// Without it two operators acting at once would both succeed and the
	// second would erase the first's audit entry.
	stub := &organizationStub{result: organizationapp.Result{Organization: organizationFixture(t)}}
	response := organizationRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations/org_1/suspend",
		`{"reasonCode":"unpaid_invoice","expectedRevision":3}`, "cmd_2")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	if stub.changed.ExpectedRevision != 3 || stub.changed.OrganizationID != "org_1" {
		t.Fatalf("changed %#v", stub.changed)
	}
}

func TestADuplicateNameIsRefusedWithItsOwnCode(t *testing.T) {
	// Two bodies with one name make the trail ambiguous about which of them a
	// code was issued for.
	stub := &organizationStub{err: organizationapp.ErrNameTaken}
	response := organizationRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations",
		`{"name":"Ashesi","billingEmail":"b@ashesi.edu.gh","reasonCode":"partner_agreement"}`,
		"cmd_1")

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "organization_name_taken") {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAReplayedRegistrationAnswersTwoHundred(t *testing.T) {
	// The operator asked for one organization and got it. Answering the retry
	// with a 201 would suggest a second one was made.
	stub := &organizationStub{result: organizationapp.Result{
		Organization: organizationFixture(t), Replayed: true,
	}}
	response := organizationRequest(t, stub, steppedUp("adm_1"), http.MethodPost,
		"/v1/admin/organizations",
		`{"name":"Ashesi","billingEmail":"b@ashesi.edu.gh","reasonCode":"partner_agreement"}`,
		"cmd_1")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", response.Code, response.Body.String())
	}
}
