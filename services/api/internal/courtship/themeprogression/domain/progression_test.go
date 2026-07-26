package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func TestThemesUnlockStrictlyAfterPriorSimultaneousReveal(t *testing.T) {
	progression, now := opened(t)
	if _, err := progression.Submit(ThemeThree, key(20), command("skip", key(1), 1, now)); !errors.Is(err, ErrLocked) {
		t.Fatalf("skip error=%v", err)
	}
	progression = submit(t, progression, ThemeTwo, key(1), key(20), "two-a", now.Add(time.Second))
	if state(progression, ThemeTwo).Revealed || len(state(progression, ThemeTwo).Submissions) != 0 ||
		state(progression, ThemeThree).Unlocked {
		t.Fatalf("first response leaked or unlocked next: %#v", progression.Projection())
	}
	progression = submit(t, progression, ThemeTwo, key(2), key(21), "two-b", now.Add(2*time.Second))
	if !state(progression, ThemeTwo).Revealed || !state(progression, ThemeThree).Unlocked ||
		state(progression, ThemeFour).Unlocked {
		t.Fatalf("theme two reveal/unlock invalid: %#v", progression.Projection())
	}
	if _, err := progression.Submit(ThemeTwo, key(22), command("overwrite", key(1), progression.Revision(), now.Add(3*time.Second))); !errors.Is(err, ErrInvalid) {
		t.Fatalf("revealed overwrite error=%v", err)
	}
}

func TestProgressionPropertyNeverSkipsOrRevealsOneResponse(t *testing.T) {
	property := func(reverseMask uint8) bool {
		progression, now := opened(t)
		revisionTime := 1
		for theme := ThemeTwo; theme <= ThemeFour; theme++ {
			future := theme + 1
			if future <= ThemeFour {
				_, err := progression.Submit(future, key(100+int(theme)), command(
					fmt.Sprintf("skip-%d", theme), key(1), progression.Revision(), now.Add(time.Duration(revisionTime)*time.Second),
				))
				if !errors.Is(err, ErrLocked) {
					return false
				}
			}
			first, second := key(1), key(2)
			if reverseMask&(1<<theme) != 0 {
				first, second = second, first
			}
			progression = submit(t, progression, theme, first, key(200+int(theme)), fmt.Sprintf("%d-a", theme), now.Add(time.Duration(revisionTime)*time.Second))
			revisionTime++
			current := state(progression, theme)
			if current.Revealed || len(current.Submissions) != 0 {
				return false
			}
			progression = submit(t, progression, theme, second, key(300+int(theme)), fmt.Sprintf("%d-b", theme), now.Add(time.Duration(revisionTime)*time.Second))
			revisionTime++
			if !state(progression, theme).Revealed || len(state(progression, theme).Submissions) != 2 {
				return false
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
}

func opened(t *testing.T) (Progression, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	progression, err := Open("progression-1", []string{key(1), key(2)}, key(9), command("open", key(1), 0, now))
	if err != nil {
		t.Fatal(err)
	}
	return progression, now
}
func submit(t *testing.T, progression Progression, theme ThemeNumber, actor, content, id string, at time.Time) Progression {
	t.Helper()
	next, err := progression.Submit(theme, content, command(id, actor, progression.Revision(), at))
	if err != nil {
		t.Fatal(err)
	}
	return next
}
func state(progression Progression, number ThemeNumber) ThemeState {
	for _, state := range progression.Projection().Themes {
		if state.Number == number {
			return state
		}
	}
	return ThemeState{}
}
func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id, actor string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorKey: actor, ReasonCode: "member_action", ExpectedRevision: revision, At: at}
}
