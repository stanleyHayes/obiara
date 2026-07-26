package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func TestContentIsConcealedUntilSimultaneousReveal(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	theme, err := Open("theme-1", []string{key(1), key(2)}, command("open", key(1), 0, now))
	if err != nil {
		t.Fatal(err)
	}
	first, err := theme.Submit(key(10), command("first", key(1), 1, now.Add(time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	projection := first.Projection()
	if projection.Revealed || projection.SubmittedCount != 1 || len(projection.Submissions) != 0 {
		t.Fatalf("first projection leaked content: %#v", projection)
	}
	if _, err := first.Submit(key(11), command("overwrite", key(1), 2, now.Add(2*time.Second))); !errors.Is(err, ErrAlreadySubmitted) {
		t.Fatalf("overwrite error=%v", err)
	}
	second, err := first.Submit(key(12), command("second", key(2), 2, now.Add(2*time.Second)))
	if err != nil {
		t.Fatal(err)
	}
	projection = second.Projection()
	if !projection.Revealed || projection.SubmittedCount != 2 || len(projection.Submissions) != 2 {
		t.Fatalf("simultaneous reveal missing: %#v", projection)
	}
	if projection.PromptRef != ThemeOnePromptRef || projection.PromptVersion != ThemeOnePromptVersion {
		t.Fatalf("prompt drift: %#v", projection)
	}
}

func TestConcealRevealPropertyIsIndependentOfSubmissionOrder(t *testing.T) {
	property := func(firstIsA bool, aRaw, bRaw uint16) bool {
		now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
		theme, err := Open("theme-1", []string{key(1), key(2)}, command("open", key(1), 0, now))
		if err != nil {
			return false
		}
		firstActor, secondActor := key(1), key(2)
		if !firstIsA {
			firstActor, secondActor = secondActor, firstActor
		}
		firstContent, secondContent := key(int(aRaw)+20), key(int(bRaw)+70000)
		theme, err = theme.Submit(firstContent, command("first", firstActor, 1, now.Add(time.Second)))
		if err != nil || theme.Projection().Revealed || len(theme.Projection().Submissions) != 0 {
			return false
		}
		theme, err = theme.Submit(secondContent, command("second", secondActor, 2, now.Add(2*time.Second)))
		return err == nil && theme.Projection().Revealed && len(theme.Projection().Submissions) == 2
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 2_000}); err != nil {
		t.Fatal(err)
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
func command(id, actor string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorKey: actor, ReasonCode: "member_action", ExpectedRevision: revision, At: at}
}
