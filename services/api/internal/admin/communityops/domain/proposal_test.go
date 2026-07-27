package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func fixture(now time.Time) (Density, HostEligibility, Notice) {
	d := Density{key(1), key(2), 40, 60, 7, now}
	h := HostEligibility{key(3), 4, 5, now.Add(time.Hour), now.Add(time.Hour), true, true}
	n := Notice{"fire.host_change", 3, "en-gh", 40, NoticeDigest("fire.host_change", 3, "en-gh", 40)}
	return d, h, n
}
func TestProposalRequiresCurrentHostAndExactAcknowledgement(t *testing.T) {
	now := time.Date(2026, 7, 27, 4, 0, 0, 0, time.UTC)
	d, h, n := fixture(now)
	p, e := Propose(key(4), key(5), ReplaceHost, ReasonHostUnavailable, key(6), d, &h, n, "propose-1", now)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.AcknowledgeNotice(key(5), key(9), "ack-1", now); e == nil {
		t.Fatal("accepted wrong preview digest")
	}
	next, e := p.AcknowledgeNotice(key(5), n.Digest, "ack-1", now)
	if e != nil {
		t.Fatal(e)
	}
	if !next.ReadyForHumanReview() || len(next.State().Audit) != 2 {
		t.Fatal("missing acknowledged review/audit state")
	}
	if _, e = next.AcknowledgeNotice(key(5), n.Digest, "ack-2", now); e == nil {
		t.Fatal("accepted second acknowledgement")
	}
}
func TestProposalRejectsExpiredOrUntrainedHostAndCapacityBypass(t *testing.T) {
	now := time.Now().UTC()
	d, h, n := fixture(now)
	h.CertifiedUntil = now
	if _, e := Propose(key(4), key(5), AssignHost, ReasonCertificationExpired, key(6), d, &h, n, "propose-1", now); e == nil {
		t.Fatal("accepted expired certification")
	}
	d.Participants = 61
	if _, e := Propose(key(4), key(5), CancelFire, ReasonSafetyCapacity, key(6), d, nil, n, "propose-2", now); e == nil {
		t.Fatal("accepted over capacity density")
	}
}
func TestDensityBoundsProperty(t *testing.T) {
	now := time.Now().UTC()
	if e := quick.Check(func(participants, capacity uint16) bool {
		d := Density{key(1), key(2), int(participants), int(capacity), 1, now}
		want := capacity > 0 && capacity <= MaxCapacity && participants <= capacity
		return validDensity(d) == want
	}, nil); e != nil {
		t.Fatal(e)
	}
}
func FuzzNoticeDigestBinding(f *testing.F) {
	f.Add("fire.host_change", uint64(2), "en-gh", 40)
	f.Fuzz(func(t *testing.T, template string, version uint64, locale string, audience int) {
		n := Notice{template, version, locale, audience, NoticeDigest(template, version, locale, audience)}
		if validNotice(n, audience) && n.Digest == NoticeDigest(template, version+1, locale, audience) {
			t.Fatal("version not bound")
		}
	})
}
