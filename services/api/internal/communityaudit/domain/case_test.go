package domain

import (
	"strings"
	"testing"
	"time"
)

func TestReasonedDecisionIsAuditedAndReplaySafe(t *testing.T) {
	n := time.Now()
	c, _ := New("case:1", KindCircle, "circle:key", "evidence:key", "legitimacy.review", n)
	a := strings.Repeat("a", 64)
	k := strings.Repeat("b", 64)
	x, e := c.Decide(true, "cmd:1", a, "evidence_verified", k, n, 1)
	if e != nil || x.Status() != StatusApproved || len(x.Audit()) != 1 {
		t.Fatal(x, e)
	}
	r, e := x.Decide(true, "cmd:1", a, "evidence_verified", k, n, 1)
	if e != nil || r.Version() != 2 {
		t.Fatal("replay")
	}
}
func FuzzReasonCodesMustBeBounded(f *testing.F) {
	f.Add("valid_reason")
	f.Fuzz(func(t *testing.T, s string) {
		n := time.Now()
		c, _ := New("case:1", KindVouch, "vouch:key", "evidence:key", "vouch.review", n)
		_, e := c.Decide(true, "cmd:1", strings.Repeat("a", 64), s, strings.Repeat("b", 64), n, 1)
		if e == nil && len(s) > 64 {
			t.Fatal("unbounded reason")
		}
	})
}
