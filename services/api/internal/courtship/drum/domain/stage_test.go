package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func TestVoiceFirstAndStrictAlternation(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	if _, err := Open("stage-1", []string{key(1), key(2)}, "", command("open", key(1), 0, now)); !errors.Is(err, ErrVoiceRequired) {
		t.Fatalf("text-only opening error=%v", err)
	}
	stage, err := Open("stage-1", []string{key(1), key(2)}, key(9), command("open", key(1), 0, now))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stage.Add(MediumText, key(10), command("double", key(1), 1, now.Add(time.Second))); !errors.Is(err, ErrNotTurn) {
		t.Fatalf("same actor error=%v", err)
	}
	stage, err = stage.Add(MediumText, key(10), command("reply", key(2), 1, now.Add(time.Second)))
	if err != nil || stage.NextActorKey() != key(1) {
		t.Fatalf("stage=%#v error=%v", stage, err)
	}
	replayed, err := stage.Add(MediumText, key(10), command("reply", key(2), 1, now.Add(time.Second)))
	if err != nil || replayed.Revision() != stage.Revision() {
		t.Fatalf("replay revision=%d error=%v", replayed.Revision(), err)
	}
}

func TestAlternationProperty(t *testing.T) {
	property := func(raw uint16) bool {
		count := int(raw%100) + 1
		now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
		stage, err := Open("stage-1", []string{key(1), key(2)}, key(9), command("open", key(1), 0, now))
		if err != nil {
			return false
		}
		for index := 0; index < count; index++ {
			actor := stage.NextActorKey()
			stage, err = stage.Add(MediumVoice, key(index+20), command(fmt.Sprintf("turn-%d", index), actor, stage.Revision(), now.Add(time.Duration(index+1)*time.Second)))
			if err != nil {
				return false
			}
			beats := stage.Beats()
			if beats[len(beats)-1].ActorKey == beats[len(beats)-2].ActorKey {
				return false
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id, actor string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorKey: actor, ReasonCode: "member_action", ExpectedRevision: revision, At: at}
}
