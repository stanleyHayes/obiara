package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string { return fmt.Sprintf("%064x", n) }
func TestExpiryNeverInventsMove(t *testing.T) {
	s, e := Create("session-1", k(3), []string{k(1), k(2)}, time.Hour, now, Command{ID: "create", At: now})
	if e != nil {
		t.Fatal(e)
	}
	before := s.Board().Houses()
	if _, e = s.Move(k(1), 0, now.Add(time.Hour), Command{ID: "late", ExpectedRevision: 1, At: now.Add(time.Hour)}); !errors.Is(e, ErrExpired) {
		t.Fatal(e)
	}
	expired, e := s.Expire(now.Add(time.Hour), Command{ID: "expire", ExpectedRevision: 1, At: now.Add(time.Hour)})
	if e != nil || expired.Board().Houses() != before || expired.Revision() != 2 {
		t.Fatal("expiry invented move")
	}
}
func TestLegalMoveProperty(t *testing.T) {
	property := func(pit uint8) bool {
		s, e := Create("session-1", k(3), []string{k(1), k(2)}, time.Hour, now, Command{ID: "create", At: now})
		if e != nil {
			return false
		}
		p := int(pit % 6)
		next, e := s.Move(k(1), p, now.Add(time.Second), Command{ID: "move", ExpectedRevision: 1, At: now.Add(time.Second)})
		return e == nil && next.Board().TotalSeeds() == 48 && next.Turn() != s.Turn()
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
