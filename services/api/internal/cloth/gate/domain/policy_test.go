package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func command(id, actor string, action Action, cap Capability, revision uint64) Command {
	return Command{id, actor, Fingerprint("policy", action, actor, cap, revision), revision, time.Now()}
}
func policy(t *testing.T) Policy {
	t.Helper()
	p, err := Open("policy", VersionV1, [2]string{key(1), key(2)}, command("open", key(1), ActionOpened, Capability{}, 0))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func TestDualConsentIntersectionAndImmediateRevoke(t *testing.T) {
	p := policy(t)
	cap := Capability{key(3), key(4), key(5)}
	if p.Allows(cap) {
		t.Fatal("default allowed")
	}
	p, err := p.Grant(cap, command("g1", key(1), ActionGranted, cap, 1))
	if err != nil {
		t.Fatal(err)
	}
	if p.Allows(cap) {
		t.Fatal("unilateral grant allowed")
	}
	p, err = p.Grant(cap, command("g2", key(2), ActionGranted, cap, 2))
	if err != nil || !p.Allows(cap) {
		t.Fatalf("dual grant=%v", err)
	}
	p, err = p.Revoke(cap, command("r1", key(1), ActionRevoked, cap, 3))
	if err != nil || p.Allows(cap) {
		t.Fatalf("revocation not immediate: %v", err)
	}
	if _, err = p.Grant(cap, command("outsider", key(9), ActionGranted, cap, 4)); !errors.Is(err, ErrNotMember) {
		t.Fatalf("outsider=%v", err)
	}
}
func FuzzNoSingleMemberCanGrantEffectiveAccess(f *testing.F) {
	f.Add(uint8(20))
	f.Fuzz(func(t *testing.T, count uint8) {
		p := policy(t)
		cap := Capability{key(3), key(4), key(5)}
		for i := 0; i < int(count%50); i++ {
			action := ActionGranted
			if i%2 == 1 {
				action = ActionRevoked
			}
			cmd := command(fmt.Sprintf("c%d", i), key(1), action, cap, p.Revision())
			var next Policy
			var err error
			if action == ActionGranted {
				next, err = p.Grant(cap, cmd)
			} else {
				next, err = p.Revoke(cap, cmd)
			}
			if err == nil {
				p = next
			}
			if p.Allows(cap) {
				t.Fatal("single member produced effective access")
			}
		}
	})
}
