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
func segments() []Segment {
	return []Segment{{Type: TypeTalk, TitleCode: "welcome", PlannedDuration: time.Minute}, {Type: TypeGame, TitleCode: "game-one", PlannedDuration: 2 * time.Minute, CapabilityRef: AllowedGameCapabilities()[0]}, {Type: TypeClose, TitleCode: "close", PlannedDuration: time.Minute}}
}
func TestTimerNeverTransitions(t *testing.T) {
	r, e := Create("sheet-1", k(8), k(9), 1, segments(), cmd("create", 0))
	if e != nil {
		t.Fatal(e)
	}
	r, e = r.Start(now, cmd("start", 1))
	if e != nil {
		t.Fatal(e)
	}
	p := r.Project(now.Add(time.Hour))
	if p.Status != StatusRunning || p.CurrentIndex != 0 || p.Remaining != 0 || r.Revision() != 2 {
		t.Fatalf("%+v", p)
	}
}
func TestExplicitOrderProperty(t *testing.T) {
	property := func(extra uint8) bool {
		r, e := Create("sheet-1", k(8), k(9), 1, segments(), cmd("create", 0))
		if e != nil {
			return false
		}
		r, e = r.Start(now, cmd("start", 1))
		if e != nil {
			return false
		}
		if extra%2 == 0 {
			r, e = r.Extend(time.Duration(extra%60+1)*time.Minute, cmd("extend", 2))
			if e != nil {
				return false
			}
		}
		return r.Current() == 0 && r.Status() == StatusRunning
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 1000}); e != nil {
		t.Fatal(e)
	}
}
