package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func room(t *testing.T) Closure {
	t.Helper()
	c, err := New("closure-room-01", []string{key(2), key(1)}, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestExactlyTwoOpaqueMembers(t *testing.T) {
	for _, members := range [][]string{{key(1)}, {key(1), key(2), key(3)}, {key(1), "member@example.com"}} {
		if _, err := New("closure-room-01", members, time.Now()); err == nil {
			t.Fatalf("accepted %#v", members)
		}
	}
}
func TestEitherMemberClosesImmediatelyWithNeutralEvent(t *testing.T) {
	for _, actor := range []string{key(1), key(2)} {
		c, err := room(t).Close(Command{"close-command-01", actor, 0, time.Unix(101, 0)})
		if err != nil || c.Status() != StatusClosed {
			t.Fatalf("status=%s err=%v", c.Status(), err)
		}
		event := c.Event()
		if event.Kind != KindMember || event.CommandID == "" || !event.At.Equal(time.Unix(101, 0)) {
			t.Fatalf("%+v", event)
		}
		if strings.Contains(fmt.Sprintf("%+v", event), actor) {
			t.Fatal("event exposes actor")
		}
	}
}
func TestInactivityUsesServerTimeAndNoActor(t *testing.T) {
	c := room(t)
	if _, err := c.CloseInactive(Command{"timeout-command-01", "", 0, time.Unix(159, 0)}, time.Minute); err != ErrNotDue {
		t.Fatalf("%v", err)
	}
	closed, err := c.CloseInactive(Command{"timeout-command-01", "", 0, time.Unix(160, 0)}, time.Minute)
	if err != nil || closed.Event().Kind != KindInactive {
		t.Fatalf("%+v %v", closed.Event(), err)
	}
	if _, err = c.CloseInactive(Command{"timeout-command-02", key(1), 0, time.Unix(160, 0)}, time.Minute); err != ErrDenied {
		t.Fatalf("%v", err)
	}
}
func TestTerminalReplayAndCAS(t *testing.T) {
	cmd := Command{"close-command-01", key(1), 0, time.Unix(101, 0)}
	c, _ := room(t).Close(cmd)
	replayed, err := c.Close(cmd)
	if err != nil || replayed.Revision() != 1 {
		t.Fatalf("%d %v", replayed.Revision(), err)
	}
	if _, err = c.Close(Command{"close-command-02", key(2), 1, time.Unix(102, 0)}); err != ErrDenied {
		t.Fatalf("%v", err)
	}
	if _, err = room(t).Close(Command{"close-command-01", key(2), 0, time.Unix(101, 0)}); err != nil {
		t.Fatalf("independent command allowed: %v", err)
	}
	if _, err = c.Close(Command{"close-command-01", key(2), 1, time.Unix(102, 0)}); err != ErrCommandMismatch {
		t.Fatalf("%v", err)
	}
	if _, err = room(t).Close(Command{"close-command-01", key(1), 7, time.Unix(101, 0)}); err != ErrStaleRevision {
		t.Fatalf("%v", err)
	}
}
func TestBoundaryPropertyNoReasonsOrReputation(t *testing.T) {
	c, _ := room(t).Close(Command{"close-command-01", key(1), 0, time.Unix(101, 0)})
	wire := strings.ToLower(fmt.Sprintf("%+v", State{ID: c.ID(), Members: c.Members(), Status: c.Status(), LastActivity: c.LastActivity(), ClosedAt: c.ClosedAt(), Revision: c.Revision(), Event: c.Event(), Commands: c.Commands()}))
	for _, bad := range []string{"because", "reason", "accusation", "readreceipt", "score", "rank", "public", "reputation"} {
		if strings.Contains(wire, bad) {
			t.Fatalf("leaked %q", bad)
		}
	}
}
