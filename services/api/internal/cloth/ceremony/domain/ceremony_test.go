package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func cmd(id, actor string, action Action, destination string, revision uint64) Command {
	return Command{id, actor, Fingerprint("ceremony", action, actor, destination, revision), revision, time.Now()}
}
func ceremony(t *testing.T) Ceremony {
	t.Helper()
	x, err := Open("ceremony", [2]string{key(1), key(2)}, cmd("open", key(1), ActionOpened, "", 0))
	if err != nil {
		t.Fatal(err)
	}
	return x
}
func TestDualCompletionAndFreshDualAnnouncementConsent(t *testing.T) {
	x := ceremony(t)
	x, _ = x.Confirm(cmd("c1", key(1), ActionConfirmed, "", 1))
	if x.Complete() {
		t.Fatal("unilateral completion")
	}
	if _, err := x.ProposeAnnouncement(key(3), cmd("early", key(1), ActionAnnouncementProposed, key(3), 2)); !errors.Is(err, ErrNotComplete) {
		t.Fatalf("early=%v", err)
	}
	x, _ = x.Confirm(cmd("c2", key(2), ActionConfirmed, "", 2))
	if !x.Complete() {
		t.Fatal("dual ceremony incomplete")
	}
	x, _ = x.ProposeAnnouncement(key(3), cmd("propose", key(1), ActionAnnouncementProposed, key(3), 3))
	if x.AnnouncementReady() {
		t.Fatal("proposal auto-ready")
	}
	x, _ = x.ConsentAnnouncement(cmd("a1", key(1), ActionAnnouncementConsented, key(3), 4))
	if x.AnnouncementReady() {
		t.Fatal("single fresh consent ready")
	}
	x, _ = x.ConsentAnnouncement(cmd("a2", key(2), ActionAnnouncementConsented, key(3), 5))
	if !x.AnnouncementReady() {
		t.Fatal("dual fresh consent not ready")
	}
}
func FuzzCeremonyNeverCompletesFromOneMember(f *testing.F) {
	f.Add(uint8(30))
	f.Fuzz(func(t *testing.T, count uint8) {
		x := ceremony(t)
		for i := 0; i < int(count%100); i++ {
			next, err := x.Confirm(cmd(fmt.Sprintf("c%d", i), key(1), ActionConfirmed, "", x.Revision()))
			if err == nil {
				x = next
			}
			if x.Complete() {
				t.Fatal("one member completed ceremony")
			}
		}
	})
}
