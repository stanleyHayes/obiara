package domain

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/quick"
	"time"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func cmd(id string, r uint64, actor int, at time.Time) Command {
	return Command{ID: id, ActorKey: key(actor), ReasonCode: "user_requested", ExpectedRevision: r, At: at}
}
func create(t *testing.T, recipients []string) Pod {
	t.Helper()
	p, e := Create("pod-1", key(1), key(2), recipients, now.Add(time.Hour), cmd("create-1", 0, 1, now))
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestCapDeterminismAndPlaybackExpiry(t *testing.T) {
	p := create(t, []string{key(4), key(3), key(3)})
	if !slices.Equal(p.RecipientKeys(), []string{key(3), key(4)}) {
		t.Fatal(p.RecipientKeys())
	}
	too := make([]string, MaxRecipients+1)
	for i := range too {
		too[i] = key(i + 10)
	}
	if _, e := Create("pod-1", key(1), key(2), too, now.Add(time.Hour), cmd("create-1", 0, 1, now)); !errors.Is(e, ErrInvalidPod) {
		t.Fatal(e)
	}
	if _, e := p.Play(cmd("play-x", 1, 9, now)); !errors.Is(e, ErrRecipientDenied) {
		t.Fatal(e)
	}
	if _, e := p.Play(cmd("play-late", 1, 3, now.Add(time.Hour))); !errors.Is(e, ErrInvalidTransition) {
		t.Fatal(e)
	}
}
func TestReplayAndRoundTripProperty(t *testing.T) {
	property := func(n uint8) bool {
		count := int(n)%MaxRecipients + 1
		rs := make([]string, count)
		for i := range rs {
			rs[i] = key(i + 3)
		}
		p := create(t, rs)
		played, e := p.Play(cmd("play-1", 1, 3, now.Add(time.Minute)))
		if e != nil {
			return false
		}
		replay, e := played.Play(cmd("play-1", 1, 3, now.Add(time.Minute)))
		if e != nil || replay.Revision() != 2 {
			return false
		}
		x, e := Rehydrate(State{ID: played.ID(), OwnerKey: played.OwnerKey(), MediaKey: played.MediaKey(), RecipientKeys: played.RecipientKeys(), Status: played.Status(), ExpiresAt: played.ExpiresAt(), EndedAt: played.EndedAt(), Revision: played.Revision(), Events: played.Events(), Commands: played.Commands()})
		return e == nil && x.Revision() == played.Revision()
	}
	if e := quick.Check(property, &quick.Config{MaxCount: 300}); e != nil {
		t.Fatal(e)
	}
}
