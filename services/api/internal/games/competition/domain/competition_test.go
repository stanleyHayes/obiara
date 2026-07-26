package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string                  { return fmt.Sprintf("%064x", n) }
func cmd(id string, r uint64) Command { return Command{ID: id, ExpectedRevision: r, At: now} }
func TestDeterministicBracketAndReviewSeparation(t *testing.T) {
	x, e := Create("competition-1", k(9), []string{k(4), k(2), k(1), k(3)}, cmd("create", 0))
	if e != nil {
		t.Fatal(e)
	}
	m := x.Matches()
	if m[0].FirstKey != k(1) || m[0].SecondKey != k(2) {
		t.Fatal("not deterministic")
	}
	x, e = x.OpenReview("review-1", m[0].ID, k(8), k(1), now, cmd("review", 1))
	if e != nil || x.Reviews()[0].Decision != DecisionNone || x.Ladder()[0].Played != 0 {
		t.Fatal("review altered ladder")
	}
}
func TestVerifiedResultShapeProperty(t *testing.T) {
	property := func(second bool) bool {
		x, e := Create("competition-1", k(9), []string{k(1), k(2), k(3), k(4)}, cmd("create", 0))
		if e != nil {
			return false
		}
		m := x.Matches()[0]
		winner := m.FirstKey
		if second {
			winner = m.SecondKey
		}
		x, e = x.RecordResult(m.ID, winner, k(8), cmd("result", 1))
		if e != nil {
			return false
		}
		p := x.Project()
		for _, v := range p.Ladder {
			if v.MemberKey == winner {
				return v.Played == 1 && v.Wins == 1
			}
		}
		return false
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
