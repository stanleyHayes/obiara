package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string { return fmt.Sprintf("%064x", n) }
func TestBoundedMappingProperty(t *testing.T) {
	events := []GameEvent{EventAbandoned, EventInactivity, EventModerationRemoval, EventRespectfulCompletion}
	property := func(i uint8) bool {
		s, e := Record("signal-1", k(1), k(2), k(3), events[int(i)%len(events)], now, Command{ID: "record", At: now})
		if e != nil {
			return false
		}
		p := s.Project()
		return p.Reason != "" && p.Provenance != "" && (p.Kind == KindConcern || p.Kind == KindAffirmation)
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
func TestAppealLifecycle(t *testing.T) {
	s, _ := Record("signal-1", k(1), k(2), k(3), EventAbandoned, now, Command{ID: "record", At: now})
	s, e := s.Appeal(now, Command{ID: "appeal", ExpectedRevision: 1, At: now})
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Resolve(AppealOverturned, now, Command{ID: "resolve", ExpectedRevision: 2, At: now})
	if e != nil || s.Project().Appeal != AppealOverturned {
		t.Fatal(e)
	}
}
