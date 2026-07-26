package domain

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func command(id string, base uint64) Command {
	return Command{ID: id, DeviceKey: key(2), ActorKey: key(3), PayloadKey: key(4), Fingerprint: key(5), BaseSequence: base, At: time.Now()}
}
func TestSequenceAndStaleDevice(t *testing.T) {
	state, _ := Open(key(1))
	next, event, err := state.Accept(command("first", 0))
	if err != nil || event.Sequence != 1 {
		t.Fatal(event, err)
	}
	if _, _, err = next.Accept(command("stale", 0)); !errors.Is(err, ErrStaleDevice) {
		t.Fatalf("stale=%v", err)
	}
}
func FuzzAcceptedEventsAreStrictlyOrdered(f *testing.F) {
	f.Add(uint8(20))
	f.Fuzz(func(t *testing.T, count uint8) {
		state, _ := Open(key(1))
		limit := int(count % 100)
		for i := 0; i < limit; i++ {
			next, event, err := state.Accept(command(fmt.Sprintf("c-%d", i), state.Sequence))
			if err != nil {
				t.Fatal(err)
			}
			if event.Sequence != state.Sequence+1 {
				t.Fatalf("sequence=%d prior=%d", event.Sequence, state.Sequence)
			}
			state = next
		}
		if state.Sequence != uint64(limit) || state.Revision != uint64(limit) {
			t.Fatalf("state=%#v", state)
		}
	})
}
