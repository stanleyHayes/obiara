package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adminapp "github.com/stanleyHayes/obiara/services/api/internal/admin/application"
	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	verificationadmin "github.com/stanleyHayes/obiara/services/api/internal/verification/admin/application"
)

var notifyNow = time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

// notifyAdminStub implements the Admin port. Only the two methods the inbox
// touches carry behaviour; the rest exist to satisfy the interface and panic
// so a future caller cannot silently rely on a fake.
type notifyAdminStub struct {
	session   admindomain.Session
	principal admindomain.Principal
	authErr   error
	changes   []admindomain.RoleChange
	changeErr error
}

func (stub notifyAdminStub) Authenticate(context.Context, string) (admindomain.Session, admindomain.Principal, error) {
	return stub.session, stub.principal, stub.authErr
}

func (stub notifyAdminStub) ListPendingRoleChanges(context.Context, string) ([]admindomain.RoleChange, error) {
	return stub.changes, stub.changeErr
}

func (notifyAdminStub) Enroll(context.Context, string, string, []admindomain.Role) (admindomain.Principal, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) ListPrincipals(context.Context, string) ([]admindomain.Principal, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) ChangeStatus(context.Context, string, string, admindomain.Status) (admindomain.Principal, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) ChangeRoles(context.Context, string, string, []admindomain.Role, *int64) (admindomain.Principal, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) ProposeAdminRoleChange(context.Context, string, string, []admindomain.Role, string) (admindomain.RoleChange, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) ApproveAdminRoleChange(context.Context, string, string) (admindomain.Principal, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) StartLogin(context.Context, string, string) error {
	panic("not used by the inbox")
}
func (notifyAdminStub) CompleteLogin(context.Context, string, string) (admindomain.Session, error) {
	panic("not used by the inbox")
}
func (notifyAdminStub) StepUpStart(context.Context, string) error { panic("not used by the inbox") }
func (notifyAdminStub) StepUpComplete(context.Context, string, string) (admindomain.Session, error) {
	panic("not used by the inbox")
}

type notifyVerificationStub struct {
	cases []verificationadmin.CaseSummary
	err   error
}

func (stub notifyVerificationStub) ListQueue(context.Context, verificationadmin.Principal, int) ([]verificationadmin.CaseSummary, error) {
	return stub.cases, stub.err
}
func (notifyVerificationStub) Detail(context.Context, verificationadmin.Principal, string) (verificationadmin.CaseDetail, error) {
	panic("not used by the inbox")
}
func (notifyVerificationStub) OpenEvidence(context.Context, verificationadmin.Principal, string, string, string, string) (verificationadmin.Evidence, error) {
	panic("not used by the inbox")
}
func (notifyVerificationStub) Decide(context.Context, verificationadmin.Principal, string, verificationadmin.Outcome, string, string, string, int64) (verificationadmin.DecisionResult, error) {
	panic("not used by the inbox")
}

type notifyMarksStub struct {
	seenAt time.Time
	saved  time.Time
	err    error
}

func (stub *notifyMarksStub) SeenAt(context.Context, string) (time.Time, error) {
	return stub.seenAt, stub.err
}
func (stub *notifyMarksStub) MarkSeen(_ context.Context, _ string, at time.Time) error {
	stub.saved = at
	return stub.err
}

func notifySession(t *testing.T) (admindomain.Session, admindomain.Principal) {
	t.Helper()
	principal, err := admindomain.NewPrincipal("adm_1", "ops@obiara.com",
		[]admindomain.Role{admindomain.RoleAdmin}, notifyNow)
	if err != nil {
		t.Fatal(err)
	}
	return admindomain.NewSession("sess_1", principal.ID(), principal.Roles(), notifyNow), principal
}

func readInbox(t *testing.T, handler http.Handler) (int, adminNotificationsResponse) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/v1/admin/notifications", nil)
	request.Header.Set("Authorization", "Bearer sess_1")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	var envelope struct {
		Data adminNotificationsResponse `json:"data"`
	}
	_ = json.Unmarshal(recorder.Body.Bytes(), &envelope)
	return recorder.Code, envelope.Data
}

func passingResolver(principalID string) AdminPrincipalResolver {
	return func(*http.Request) (verificationadmin.Principal, error) {
		return verificationadmin.Principal{ActorID: principalID, Scopes: []verificationadmin.Scope{adminOperationsScope}}, nil
	}
}

func TestInboxProjectsBothQueuesAndCountsUnread(t *testing.T) {
	session, principal := notifySession(t)
	admin := notifyAdminStub{session: session, principal: principal, changes: pendingChange(t)}
	verification := notifyVerificationStub{cases: []verificationadmin.CaseSummary{
		{ID: "IDV-1", SubmittedAt: notifyNow.Add(-2 * time.Hour)},
		{ID: "IDV-2", SubmittedAt: notifyNow.Add(-30 * time.Minute)},
	}}
	// Acknowledged an hour ago: the newer verification and the role change
	// are unread, the older verification does not make its own item.
	marks := &notifyMarksStub{seenAt: notifyNow.Add(-time.Hour)}

	handler := adminNotificationsHandler(admin, verification, marks, passingResolver("adm_1"))
	code, body := readInbox(t, handler)
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	if len(body.Items) != 2 {
		t.Fatalf("items = %#v, want verification and role-change entries", body.Items)
	}
	if body.Items[0].Key != "verification_queue" || body.Items[0].Count != 2 {
		t.Fatalf("verification item = %#v", body.Items[0])
	}
	// The item's timestamp must be the NEWEST in the queue, or an inbox
	// would look stale the moment an old case lingers.
	if !body.Items[0].LatestAt.Equal(notifyNow.Add(-30 * time.Minute)) {
		t.Fatalf("latestAt = %s, want the newest case", body.Items[0].LatestAt)
	}
	if body.UnreadCount != 2 {
		t.Fatalf("unreadCount = %d, want 2", body.UnreadCount)
	}
}

func TestInboxOmitsSourcesTheOperatorMayNotSee(t *testing.T) {
	session, principal := notifySession(t)
	// A verifier has no business seeing admin-role proposals, and an
	// operator without the verification scope has none seeing that queue.
	// Both refusals must drop the source silently rather than fail the
	// whole inbox or leak that the source exists.
	admin := notifyAdminStub{session: session, principal: principal, changeErr: adminapp.ErrNotAdmin}
	verification := notifyVerificationStub{err: verificationadmin.ErrForbidden}
	marks := &notifyMarksStub{}

	code, body := readInbox(t, adminNotificationsHandler(admin, verification, marks, passingResolver("adm_1")))
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200 with an empty inbox", code)
	}
	if len(body.Items) != 0 || body.UnreadCount != 0 {
		t.Fatalf("body = %#v, want an empty inbox", body)
	}
}

func TestInboxTreatsANeverAcknowledgedInboxAsFullyUnread(t *testing.T) {
	session, principal := notifySession(t)
	admin := notifyAdminStub{session: session, principal: principal}
	verification := notifyVerificationStub{cases: []verificationadmin.CaseSummary{
		{ID: "IDV-1", SubmittedAt: notifyNow.Add(-90 * 24 * time.Hour)},
	}}

	_, body := readInbox(t, adminNotificationsHandler(admin, verification, &notifyMarksStub{}, passingResolver("adm_1")))
	if body.SeenAt != nil {
		t.Fatalf("seenAt = %v, want null for an operator who never acknowledged", body.SeenAt)
	}
	if body.UnreadCount != 1 || !body.Items[0].Unread {
		t.Fatalf("body = %#v, want the item unread", body)
	}
}

func TestInboxSurfacesUnexpectedFailuresInsteadOfAnEmptyInbox(t *testing.T) {
	session, principal := notifySession(t)
	// A real fault must not read as "nothing needs your attention" — that
	// is the one wrong answer, because it looks exactly like success.
	admin := notifyAdminStub{session: session, principal: principal, changeErr: adminapp.ErrPrincipalConflict}
	code, _ := readInbox(t, adminNotificationsHandler(admin, notifyVerificationStub{}, &notifyMarksStub{}, passingResolver("adm_1")))
	if code == http.StatusOK {
		t.Fatal("an unexpected repository error was reported as an empty inbox")
	}
}

func TestInboxRequiresASession(t *testing.T) {
	admin := notifyAdminStub{authErr: adminapp.ErrSessionNotFound}
	code, _ := readInbox(t, adminNotificationsHandler(admin, notifyVerificationStub{}, &notifyMarksStub{}, passingResolver("adm_1")))
	if code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", code)
	}
}

func TestMarkSeenRecordsTheWatermarkForTheAuthenticatedOperator(t *testing.T) {
	session, principal := notifySession(t)
	marks := &notifyMarksStub{}
	handler := adminNotificationsSeenHandler(
		notifyAdminStub{session: session, principal: principal}, marks, func() time.Time { return notifyNow })

	request := httptest.NewRequest(http.MethodPost, "/v1/admin/notifications/seen", nil)
	request.Header.Set("Authorization", "Bearer sess_1")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if !marks.saved.Equal(notifyNow) {
		t.Fatalf("saved = %s, want %s", marks.saved, notifyNow)
	}
}

func pendingChange(t *testing.T) []admindomain.RoleChange {
	t.Helper()
	change, err := admindomain.NewRoleChange("rc_1", "adm_2", 3,
		[]admindomain.Role{admindomain.RoleAdmin}, "promote for coverage", "adm_1", notifyNow.Add(-10*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return []admindomain.RoleChange{change}
}
