package invariants_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	allowance "github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
	garden "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
	pod "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	safety "github.com/stanleyHayes/obiara/services/api/internal/seed/safety/domain"
	sprout "github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
	water "github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
)

func opaque(n int) string { return fmt.Sprintf("%064x", n) }
func podCommand(id string, actor string, revision uint64, at time.Time) pod.Command {
	return pod.Command{ID: id, ActorKey: actor, ReasonCode: "invariant.test", ExpectedRevision: revision, At: at}
}
func waterCommand(id, actor string, revision uint64, at time.Time) water.Command {
	return water.Command{ID: id, ActorKey: actor, ReasonCode: "invariant.test", ExpectedRevision: revision, At: at}
}

func TestCapsMutualityAlternationAndDeclineCannotBeBypassed(t *testing.T) {
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	recipients := make([]string, pod.MaxRecipients+1)
	for i := range recipients {
		recipients[i] = opaque(i + 10)
	}
	if _, err := pod.Create("pod", opaque(1), opaque(2), recipients, now.Add(time.Hour), podCommand("create", opaque(1), 0, now)); !errors.Is(err, pod.ErrInvalidPod) {
		t.Fatalf("recipient cap bypass=%v", err)
	}

	members := []string{opaque(3), opaque(4)}
	watering, err := water.Start("water", members, waterCommand("first", members[0], 0, now))
	if err != nil {
		t.Fatal(err)
	}
	if watering.Status() != water.StatusAwaiting || watering.RoomKey() != "" {
		t.Fatal("unilateral water created room")
	}
	if _, err = watering.Water(waterCommand("repeat", members[0], 1, now), opaque(9)); !errors.Is(err, water.ErrAlreadyWatered) {
		t.Fatalf("repeat water=%v", err)
	}
	watering, err = watering.Water(waterCommand("second", members[1], 1, now), opaque(9))
	if err != nil {
		t.Fatal(err)
	}
	if watering.Status() != water.StatusRoomCreated {
		t.Fatal("mutual water did not create room")
	}

	doorway, err := sprout.Open("doorway", members[0], members[1], now)
	if err != nil {
		t.Fatal(err)
	}
	doorway, _, err = doorway.Exchange(members[0], opaque(20), "x1", "fp1", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = doorway.Exchange(members[0], opaque(21), "x2", "fp2", now); !errors.Is(err, sprout.ErrOutOfTurn) {
		t.Fatalf("alternation bypass=%v", err)
	}
	doorway, _, _ = doorway.Exchange(members[1], opaque(22), "x2", "fp2", now)
	doorway, _, _ = doorway.Exchange(members[0], opaque(23), "x3", "fp3", now)
	if _, _, err = doorway.Exchange(members[1], opaque(24), "x4", "fp4", now); !errors.Is(err, sprout.ErrDoorwaySealed) {
		t.Fatalf("exchange cap bypass=%v", err)
	}

	item, err := garden.New(opaque(30), opaque(31), now.Add(time.Hour), now)
	if err != nil {
		t.Fatal(err)
	}
	item, err = item.Transition(garden.StateDeclined, item.Revision, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []garden.State{garden.StateDelivered, garden.StateHeard, garden.StateSprouted} {
		if _, transitionErr := item.Transition(state, item.Revision, now.Add(2*time.Minute)); !errors.Is(transitionErr, garden.ErrInvalidTransition) {
			t.Fatalf("decline resurrected into %s: %v", state, transitionErr)
		}
	}
}

func TestSafetyThrottleCannotBeRacedWithStaleRevision(t *testing.T) {
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	bucket, err := safety.New(opaque(1), now)
	if err != nil {
		t.Fatal(err)
	}
	expected := bucket.Revision
	first, decision, err := bucket.Evaluate(safety.OperationSow, expected, now)
	if err != nil || !decision.Allowed {
		t.Fatal(err)
	}
	if _, _, err = bucket.Evaluate(safety.OperationSow, expected+1, now); !errors.Is(err, safety.ErrStaleRevision) {
		t.Fatalf("stale copy bypass=%v", err)
	}
	for i := 1; i < 6; i++ {
		first, decision, err = first.Evaluate(safety.OperationSow, first.Revision, now)
		if err != nil || !decision.Allowed {
			t.Fatalf("allow %d: %v", i, err)
		}
	}
	first, decision, err = first.Evaluate(safety.OperationSow, first.Revision, now)
	if err != nil || decision.Allowed {
		t.Fatalf("seventh sow allowed: %#v %v", decision, err)
	}
}

func FuzzCrossContextStateMachineNeverBypassesTerminalInvariants(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6})
	f.Fuzz(func(t *testing.T, operations []byte) {
		if len(operations) > 256 {
			t.Skip()
		}
		now := time.Unix(1_800_000_000, 0).UTC()
		ledger, err := allowance.Issue(opaque(1), 6, now, now, "issue", "issue-fp")
		if err != nil {
			t.Fatal(err)
		}
		doorway, _ := sprout.Open("doorway", opaque(2), opaque(3), now)
		item, _ := garden.New(opaque(4), opaque(5), now.Add(time.Hour), now)
		for index, operation := range operations {
			command := fmt.Sprintf("command-%d", index)
			switch operation % 6 {
			case 0:
				next, changed, _ := ledger.Spend(int64(operation%8)+1, now, now, command, "fp-"+command)
				if changed {
					ledger = next
				}
			case 1, 2:
				actor := opaque(2)
				if operation%2 == 0 {
					actor = opaque(3)
				}
				next, changed, _ := doorway.Exchange(actor, opaque(10+index), command, "fp-"+command, now)
				if changed {
					doorway = next
				}
			case 3:
				next, transitionErr := item.Transition(garden.StateDeclined, item.Revision, now.Add(time.Minute))
				if transitionErr == nil {
					item = next
				}
			case 4:
				next, transitionErr := item.Transition(garden.StateDelivered, item.Revision, now.Add(time.Minute))
				if transitionErr == nil {
					item = next
				}
			case 5:
				next, transitionErr := item.Transition(garden.StateSprouted, item.Revision, now.Add(2*time.Minute))
				if transitionErr == nil {
					item = next
				}
			}
			if ledger.Balance() < 0 {
				t.Fatalf("negative allowance after %d operations", index)
			}
			exchanges := doorway.Exchanges()
			if len(exchanges) > sprout.MaxExchanges {
				t.Fatalf("doorway exceeded cap: %d", len(exchanges))
			}
			for turn := 1; turn < len(exchanges); turn++ {
				if exchanges[turn].ActorKey == exchanges[turn-1].ActorKey {
					t.Fatal("doorway accepted repeated actor")
				}
			}
			if item.State == garden.StateDeclined || item.State == garden.StateSprouted || item.State == garden.StateExpired {
				before := item
				if next, transitionErr := item.Transition(garden.StateDelivered, item.Revision, now.Add(3*time.Minute)); transitionErr == nil || next != (garden.Item{}) {
					t.Fatalf("terminal state %s resurrected from %#v to %#v", before.State, before, next)
				}
			}
		}
	})
}
