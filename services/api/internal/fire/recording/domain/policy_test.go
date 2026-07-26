package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func cmd(id, actor, subject string, action Action, p Purpose, r time.Duration, rev uint64) Command {
	return Command{id, actor, Fingerprint("policy", action, actor, subject, p, r, rev), rev, time.Now()}
}
func policy(t *testing.T) Policy {
	t.Helper()
	p, err := Open("policy", key(9), key(1), []string{key(1), key(2)}, cmd("open", key(1), key(1), ActionOpened, "", 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestAllCurrentConsentJoinRevokeAndRetention(t *testing.T) {
	p := policy(t)
	if _, err := p.Propose(PurposeArchive, 31*24*time.Hour, cmd("bad", key(1), key(1), ActionProposed, PurposeArchive, 31*24*time.Hour, 1)); !errors.Is(err, ErrDenied) {
		t.Fatalf("retention=%v", err)
	}
	p, _ = p.Propose(PurposeArchive, 24*time.Hour, cmd("propose", key(1), key(1), ActionProposed, PurposeArchive, 24*time.Hour, 1))
	p, _ = p.OptIn(cmd("o1", key(1), key(1), ActionOptedIn, "", 0, 2))
	if _, err := p.Start(key(8), cmd("early", key(1), key(1), ActionStarted, "", 0, 3)); !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("early=%v", err)
	}
	p, _ = p.OptIn(cmd("o2", key(2), key(2), ActionOptedIn, "", 0, 3))
	p, _ = p.Start(key(8), cmd("start", key(1), key(1), ActionStarted, "", 0, 4))
	p, _ = p.Join(key(3), cmd("join", key(1), key(3), ActionJoined, "", 0, 5))
	if p.State().Active {
		t.Fatal("join did not stop")
	}
	if _, err := p.Start(key(8), cmd("restart", key(1), key(1), ActionStarted, "", 0, 6)); !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("joiner bypass=%v", err)
	}
}
func FuzzRecordingNeverStartsWithoutAllParticipants(f *testing.F) {
	f.Add(uint8(10))
	f.Fuzz(func(t *testing.T, n uint8) {
		p := policy(t)
		p, _ = p.Propose(PurposeReflection, time.Hour, cmd("p", key(1), key(1), ActionProposed, PurposeReflection, time.Hour, 1))
		if n%2 == 0 {
			p, _ = p.OptIn(cmd("o", key(1), key(1), ActionOptedIn, "", 0, 2))
		}
		if _, err := p.Start(key(8), cmd("s", key(1), key(1), ActionStarted, "", 0, p.Revision())); err == nil {
			t.Fatal("partial consent started")
		}
	})
}
