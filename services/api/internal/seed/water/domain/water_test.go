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
func c(id string, actor int, revision uint64) Command {
	return Command{ID: id, ActorKey: k(actor), ReasonCode: "member_watered", ExpectedRevision: revision, At: now}
}
func first(t *testing.T) Water {
	w, e := Start("water-1", []string{k(1), k(2)}, c("water-a", 1, 0))
	if e != nil {
		t.Fatal(e)
	}
	return w
}
func TestMutualityAndSingleRoom(t *testing.T) {
	w := first(t)
	if w.Status() != StatusAwaiting || w.RoomKey() != "" {
		t.Fatal("room before mutuality")
	}
	if _, e := w.Water(c("again", 1, 1), k(9)); !errors.Is(e, ErrAlreadyWatered) {
		t.Fatal(e)
	}
	done, e := w.Water(c("water-b", 2, 1), k(9))
	if e != nil || done.Status() != StatusRoomCreated || done.RoomKey() != k(9) {
		t.Fatalf("%+v %v", done, e)
	}
	replay, e := done.Water(c("water-b", 2, 1), k(9))
	if e != nil || replay.Revision() != 2 {
		t.Fatal(e)
	}
}
func TestHistoryProperty(t *testing.T) {
	property := func(swap bool) bool {
		a, b := 1, 2
		if swap {
			a, b = b, a
		}
		w, e := Start("water-1", []string{k(1), k(2)}, c("first", a, 0))
		if e != nil {
			return false
		}
		w, e = w.Water(c("second", b, 1), k(9))
		if e != nil {
			return false
		}
		x, e := Rehydrate(State{ID: w.ID(), Members: w.Members(), Watered: w.Watered(), RoomKey: w.RoomKey(), Status: w.Status(), Revision: w.Revision(), Events: w.Events(), Commands: w.Commands()})
		return e == nil && x.RoomKey() == k(9) && x.Revision() == 2
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 300}); e != nil {
		t.Fatal(e)
	}
}
