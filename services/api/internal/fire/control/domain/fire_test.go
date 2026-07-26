package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func cmd(id, actor, target string, action Action, revision uint64) Command {
	return Command{id, actor, "moderation.test", Fingerprint("control", action, actor, target, "moderation.test", revision), revision, time.Now()}
}
func fire(t *testing.T) Fire {
	t.Helper()
	f, err := Open("control", key(10), key(1), []string{key(2), key(3)}, cmd("open", key(1), key(1), ActionOpened, 0))
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func TestLeastPrivilegeAndTerminalEject(t *testing.T) {
	f := fire(t)
	if _, err := f.Promote(key(2), cmd("bad", key(3), key(2), ActionPromoted, 1)); !errors.Is(err, ErrDenied) {
		t.Fatalf("participant promoted=%v", err)
	}
	f, _ = f.Promote(key(2), cmd("promote", key(1), key(2), ActionPromoted, 1))
	if _, err := f.Mute(key(1), cmd("mute-host", key(2), key(1), ActionMuted, 2)); !errors.Is(err, ErrDenied) {
		t.Fatalf("cohost muted host=%v", err)
	}
	if _, err := f.Mute(key(2), cmd("mute-peer", key(1), key(2), ActionMuted, 2)); !errors.Is(err, ErrDenied) {
		t.Fatalf("host muted cohost=%v", err)
	}
	f, _ = f.Eject(key(3), cmd("eject", key(2), key(3), ActionEjected, 2))
	if _, err := f.Mute(key(3), cmd("again", key(1), key(3), ActionMuted, 3)); !errors.Is(err, ErrDenied) {
		t.Fatalf("ejected target acted on=%v", err)
	}
}
func FuzzCohostNeverControlsPrivilegedMember(f *testing.F) {
	f.Add(uint8(7))
	f.Fuzz(func(t *testing.T, n uint8) {
		state := fire(t)
		state, _ = state.Promote(key(2), cmd("promote", key(1), key(2), ActionPromoted, 1))
		target := key(1)
		if n%2 == 0 {
			target = key(2)
		}
		if _, err := state.Eject(target, cmd("eject", key(2), target, ActionEjected, 2)); !errors.Is(err, ErrDenied) {
			t.Fatalf("cohost ejected privileged target: %v", err)
		}
	})
}
