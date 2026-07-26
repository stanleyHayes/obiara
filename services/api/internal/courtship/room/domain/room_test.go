package domain

import (
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func k(n int) string { return fmt.Sprintf("%064x", n) }
func c(id string, a int, r uint64) Command {
	return Command{ID: id, ActorKey: k(a), ReasonCode: "member_action", ExpectedRevision: r, At: now}
}
func TestExactlyTwoAndDeterministicProjection(t *testing.T) {
	r, e := Open("room-1", []string{k(2), k(1)}, c("open", 1, 0))
	if e != nil {
		t.Fatal(e)
	}
	r, e = r.Message(k(9), c("message", 2, 1))
	if e != nil {
		t.Fatal(e)
	}
	r, e = r.Close(c("close", 1, 2))
	if e != nil {
		t.Fatal(e)
	}
	if r.Projection().Status != StatusClosed || r.Projection().MessageCount != 1 {
		t.Fatal(r.Projection())
	}
	p, e := Project(r.Events())
	if e != nil || p != r.Projection() {
		t.Fatalf("%+v %v", p, e)
	}
}
func TestEventRoundTripProperty(t *testing.T) {
	property := func(n uint8) bool {
		r, e := Open("room-1", []string{k(1), k(2)}, c("open", 1, 0))
		if e != nil {
			return false
		}
		count := int(n) % 40
		for i := 0; i < count; i++ {
			r, e = r.Message(k(i+10), c(fmt.Sprintf("m-%d", i), i%2+1, uint64(i+1)))
			if e != nil {
				return false
			}
		}
		x, e := Rehydrate(State{ID: r.ID(), Members: r.Members(), Events: r.Events(), Commands: r.Commands()})
		return e == nil && x.Projection() == r.Projection()
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 300}); e != nil {
		t.Fatal(e)
	}
}
