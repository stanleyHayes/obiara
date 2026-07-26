package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func cmd(id string, actor string, revision uint64, action Action, at time.Time) Command {
	return Command{id, actor, Fingerprint(id, action, actor, revision), revision, at}
}
func proposal(t *testing.T) (Proposal, time.Time) {
	t.Helper()
	now := time.Now().UTC()
	p, err := Create("proposal", TypeExclusivity, key(1), key(2), key(3), now.Add(time.Hour), cmd("create", key(1), 0, ActionCreated, now))
	if err != nil {
		t.Fatal(err)
	}
	return p, now
}
func TestRolesExpiryTerminalAndReplay(t *testing.T) {
	p, now := proposal(t)
	if _, err := p.Accept(cmd("wrong", key(1), 1, ActionAccepted, now)); !errors.Is(err, ErrNotParticipant) {
		t.Fatalf("sender accepted: %v", err)
	}
	accepted, err := p.Accept(cmd("accept", key(2), 1, ActionAccepted, now))
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := accepted.Accept(cmd("accept", key(2), 1, ActionAccepted, now))
	if err != nil || replayed.Revision() != 2 {
		t.Fatalf("replay=%v", err)
	}
	if _, err = accepted.Reject(cmd("reject", key(2), 2, ActionRejected, now)); !errors.Is(err, ErrTerminal) {
		t.Fatalf("terminal changed=%v", err)
	}
	p, _ = proposal(t)
	if _, err = p.Accept(cmd("late", key(2), 1, ActionAccepted, now.Add(2*time.Hour))); !errors.Is(err, ErrExpired) {
		t.Fatalf("late=%v", err)
	}
}
func FuzzOnlyRequiredRoleCanDecide(f *testing.F) {
	f.Add(uint8(7))
	f.Fuzz(func(t *testing.T, n uint8) {
		p, now := proposal(t)
		actor := key(int(n%20) + 10)
		if actor == key(2) {
			t.Skip()
		}
		if _, err := p.Accept(cmd("accept", actor, 1, ActionAccepted, now)); !errors.Is(err, ErrNotParticipant) {
			t.Fatalf("actor accepted: %v", err)
		}
	})
}
