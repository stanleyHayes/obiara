package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func fresh(t *testing.T) Relay {
	r, e := New("relay-pair-01", []string{key(1), key(2)}, key(3))
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func submit(r Relay) (Relay, error) {
	return r.Submit(Command{ID: "submit-command-01", ActorKey: key(3), QuestionRef: "ref_questionabcdefghijklmnop", PromptRef: "ref_promptabcdefghijklmnop", At: time.Unix(100, 0)})
}
func TestExactlyTwoMembersOneReviewer(t *testing.T) {
	if _, e := New("relay-pair-01", []string{key(1)}, key(3)); e == nil {
		t.Fatal("accepted")
	}
	if _, e := New("relay-pair-01", []string{key(1), key(2)}, key(1)); e == nil {
		t.Fatal("member reviewer")
	}
}
func TestIntersectionAndImmediateRevoke(t *testing.T) {
	r, _ := submit(fresh(t))
	for i, m := range []string{key(1), key(2)} {
		r, _ = r.Grant(Command{ID: fmt.Sprintf("grant-command-0%d", i+1), ActorKey: m, QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_responseabcdefghijklmnop", ExpectedRevision: uint64(i + 1), At: time.Unix(101+int64(i), 0)})
		if _, e := r.Project(key(3), "ref_questionabcdefghijklmnop"); (i == 0) != (e == ErrDenied) {
			t.Fatalf("i=%d e=%v", i, e)
		}
	}
	r, _ = r.Revoke(Command{ID: "revoke-command-01", ActorKey: key(1), QuestionRef: "ref_questionabcdefghijklmnop", ExpectedRevision: 3, At: time.Unix(103, 0)})
	if _, e := r.Project(key(3), "ref_questionabcdefghijklmnop"); e != ErrDenied {
		t.Fatalf("%v", e)
	}
}
func TestNoUnilateralMismatchedOrUnauthorizedAccess(t *testing.T) {
	r, _ := submit(fresh(t))
	r, _ = r.Grant(Command{ID: "grant-command-01", ActorKey: key(1), QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_responseabcdefghijklmnop", ExpectedRevision: 1, At: time.Now()})
	r, _ = r.Grant(Command{ID: "grant-command-02", ActorKey: key(2), QuestionRef: "ref_questionabcdefghijklmnop", ResponseRef: "ref_otherabcdefghijklmnopqrst", ExpectedRevision: 2, At: time.Now()})
	for _, reviewer := range []string{key(3), key(9)} {
		if _, e := r.Project(reviewer, "ref_questionabcdefghijklmnop"); e != ErrDenied {
			t.Fatalf("%v", e)
		}
	}
}
func TestReplayCASImmutableAuditPrivacy(t *testing.T) {
	c := Command{ID: "submit-command-01", ActorKey: key(3), QuestionRef: "ref_questionabcdefghijklmnop", PromptRef: "ref_promptabcdefghijklmnop", At: time.Unix(100, 0)}
	r, _ := fresh(t).Submit(c)
	again, e := r.Submit(c)
	if e != nil || len(again.Audit()) != 1 {
		t.Fatalf("%v", e)
	}
	c.PromptRef = "ref_changedabcdefghijklmnop"
	if _, e = r.Submit(c); e != ErrCommandMismatch {
		t.Fatalf("%v", e)
	}
	wire := strings.ToLower(fmt.Sprintf("%+v", r))
	for _, bad := range []string{"raw content", "full-thread", "public", "reverse"} {
		if strings.Contains(wire, bad) {
			t.Fatal(bad)
		}
	}
}
