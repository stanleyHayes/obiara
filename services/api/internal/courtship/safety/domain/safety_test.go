package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func aggregate(t *testing.T) Safety {
	t.Helper()
	s, e := New("safety-room-01", []string{key(2), key(1)})
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestExactlyTwoOpaqueMembers(t *testing.T) {
	for _, m := range [][]string{{key(1)}, {key(1), key(2), key(3)}, {key(1), "raw-member"}} {
		if _, e := New("safety-room-01", m); e == nil {
			t.Fatalf("accepted %#v", m)
		}
	}
}
func TestEitherMemberBlocksTerminally(t *testing.T) {
	for _, actor := range []string{key(1), key(2)} {
		s, e := aggregate(t).Block(Command{ID: "block-command-01", ActorKey: actor, At: time.Unix(100, 0)})
		if e != nil || !s.Blocked() {
			t.Fatalf("%v %v", s.Blocked(), e)
		}
		for _, m := range s.Members() {
			if e = s.CanContact(m); e != ErrContactBlocked {
				t.Fatalf("%v", e)
			}
		}
		if _, e = s.Block(Command{ID: "block-command-02", ActorKey: key(2), ExpectedRevision: 1, At: time.Unix(101, 0)}); e != ErrContactBlocked {
			t.Fatalf("%v", e)
		}
	}
}
func TestReportBoundedImmutableAndReplayChecked(t *testing.T) {
	for _, category := range []Category{CategoryHarassment, CategoryIdentity, CategoryThreat, CategoryOther} {
		cmd := Command{ID: "report-command-01", ActorKey: key(1), Category: category, EvidenceRef: "enc_abcdefghijklmnopqrstuvwxyz", At: time.Unix(100, 0)}
		s, e := aggregate(t).Report(cmd)
		if e != nil || len(s.Reviews()) != 1 {
			t.Fatalf("%v %v", s.Reviews(), e)
		}
		replay, e := s.Report(cmd)
		if e != nil || len(replay.Reviews()) != 1 || replay.Revision() != 1 {
			t.Fatalf("%v %v", replay.Reviews(), e)
		}
		cmd.Category = Category("custom")
		if _, e = s.Report(cmd); e != ErrInvalid && e != ErrCommandMismatch {
			t.Fatalf("%v", e)
		}
	}
}
func TestRejectsFreeTextAndUnencryptedEvidence(t *testing.T) {
	for _, cmd := range []Command{{ID: "report-command-01", ActorKey: key(1), Category: "because they lied", EvidenceRef: "enc_abcdefghijklmnopqrstuvwxyz", At: time.Now()}, {ID: "report-command-01", ActorKey: key(1), Category: CategoryOther, EvidenceRef: "screenshot.png", At: time.Now()}} {
		if _, e := aggregate(t).Report(cmd); e != ErrInvalid {
			t.Fatalf("%v", e)
		}
	}
}
func TestCASFingerprintAndPrivacyBoundary(t *testing.T) {
	s := aggregate(t)
	if _, e := s.Block(Command{ID: "block-command-01", ActorKey: key(1), ExpectedRevision: 2, At: time.Now()}); e != ErrStaleRevision {
		t.Fatalf("%v", e)
	}
	cmd := Command{ID: "report-command-01", ActorKey: key(1), Category: CategoryThreat, EvidenceRef: "enc_abcdefghijklmnopqrstuvwxyz", At: time.Unix(100, 0)}
	s, _ = s.Report(cmd)
	cmd.EvidenceRef = "enc_zyxwvutsrqponmlkjihgfedcba"
	if _, e := s.Report(cmd); e != ErrCommandMismatch {
		t.Fatalf("%v", e)
	}
	wire := strings.ToLower(fmt.Sprintf("%+v", State{ID: s.ID(), Members: s.Members(), Blocked: s.Blocked(), BlockedAt: s.BlockedAt(), Revision: s.Revision(), Events: s.Events(), Reviews: s.Reviews(), Commands: s.Commands()}))
	for _, bad := range []string{"raw-member", "reason", "accusation", "score", "public", "reverse", "because"} {
		if strings.Contains(wire, bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
