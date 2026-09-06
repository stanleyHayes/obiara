package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/water/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func key(n int) string          { return fmt.Sprintf("%064x", n) }
func TestSecondWaterCreatesOnlyOpaqueRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, c, k := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockPairConsent(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	w, _ := domain.Start("water-1", []string{key(1), key(2)}, domain.Command{ID: "first", ActorKey: key(1), ReasonCode: "member_watered", At: now})
	s := NewService(r, a, c, k, fixedID("water-x"), fixedID("raw-room-id"), func() time.Time { return now })
	r.EXPECT().Find(gomock.Any(), "water-1").Return(w, nil)
	a.EXPECT().Require(gomock.Any(), "member-b", "seed.water.mutual", "water-1")
	k.EXPECT().Key("seed-water:member", "member-b").Return(key(2), nil)
	k.EXPECT().Key("seed-water:member", "member-a").Return(key(1), nil)
	// Raw ids, not the water's keys. See TestTheConsentCheckIsAskedAboutPeople.
	c.EXPECT().Revalidate(gomock.Any(), "member-b", "member-a")
	k.EXPECT().Key("seed-water:room", "raw-room-id").Return(key(9), nil)
	r.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "second").DoAndReturn(func(_ context.Context, w domain.Water, _ uint64, _ string) error {
		if w.RoomKey() != key(9) {
			t.Fatal("room reference was not opaque")
		}
		return nil
	})
	x, e := s.Water(context.Background(), Command{ID: "second", WaterID: "water-1", ActorID: "member-b", CounterpartID: "member-a", ReasonCode: "member_watered", ExpectedRevision: 1})
	if e != nil || x.Water.RoomKey() != key(9) {
		t.Fatalf("%+v %v", x, e)
	}
}

// recordingConsent remembers what it was asked, so a test can say whether the
// question was about people or about opaque keys.
type recordingConsent struct {
	first, second string
	calls         int
	err           error
}

func (c *recordingConsent) Revalidate(_ context.Context, a, b string) error {
	c.calls++
	c.first, c.second = a, b
	return c.err
}

func TestTheConsentCheckIsAskedAboutPeople(t *testing.T) {
	// Start passed raw member ids; Water passed the aggregate's keys. Any
	// PairConsent that keys its own arguments — which is every one of them,
	// because that is how a context looks a member up — would compare a key
	// against a key-of-a-key and find nothing. So a block placed after the
	// water started was never honoured on the step that opens the room.
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockKeyer(ctrl)
	consent := &recordingConsent{}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	w, _ := domain.Start("water-1", []string{key(1), key(2)},
		domain.Command{ID: "first", ActorKey: key(1), ReasonCode: "member_watered", At: now})

	r.EXPECT().Find(gomock.Any(), "water-1").Return(w, nil)
	a.EXPECT().Require(gomock.Any(), "member-b", "seed.water.mutual", "water-1")
	k.EXPECT().Key("seed-water:member", "member-b").Return(key(2), nil)
	k.EXPECT().Key("seed-water:member", "member-a").Return(key(1), nil)
	k.EXPECT().Key("seed-water:room", gomock.Any()).Return(key(9), nil)
	r.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "second").Return(nil)

	s := NewService(r, a, consent, k, fixedID("water-x"), fixedID("room"), func() time.Time { return now })
	if _, e := s.Water(context.Background(), Command{
		ID: "second", WaterID: "water-1", ActorID: "member-b",
		CounterpartID: "member-a", ReasonCode: "member_watered", ExpectedRevision: 1,
	}); e != nil {
		t.Fatal(e)
	}
	if consent.first != "member-b" || consent.second != "member-a" {
		t.Fatalf("consent was asked about %q and %q, want the two people", consent.first, consent.second)
	}
}

func TestARefusedPairNeverWaters(t *testing.T) {
	// A block placed between the first water and the second closes it. The
	// room is never keyed and nothing is appended: the mocks carry no
	// expectation for either, so reaching them fails this.
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockKeyer(ctrl)
	consent := &recordingConsent{err: errors.New("blocked")}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	w, _ := domain.Start("water-1", []string{key(1), key(2)},
		domain.Command{ID: "first", ActorKey: key(1), ReasonCode: "member_watered", At: now})

	r.EXPECT().Find(gomock.Any(), "water-1").Return(w, nil)
	a.EXPECT().Require(gomock.Any(), "member-b", "seed.water.mutual", "water-1")
	k.EXPECT().Key("seed-water:member", "member-b").Return(key(2), nil)
	k.EXPECT().Key("seed-water:member", "member-a").Return(key(1), nil)

	s := NewService(r, a, consent, k, fixedID("water-x"), fixedID("room"), func() time.Time { return now })
	if _, e := s.Water(context.Background(), Command{
		ID: "second", WaterID: "water-1", ActorID: "member-b",
		CounterpartID: "member-a", ReasonCode: "member_watered", ExpectedRevision: 1,
	}); !errors.Is(e, ErrNotAvailable) {
		t.Fatalf("e = %v, want ErrNotAvailable", e)
	}
}

func TestACounterpartWhoIsNotInThisWaterIsRefused(t *testing.T) {
	// The counterpart is supplied by the caller, so it has to be checked
	// against the water rather than trusted. Naming somebody else would ask
	// consent about the wrong pair and water anyway.
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockKeyer(ctrl)
	consent := &recordingConsent{}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	w, _ := domain.Start("water-1", []string{key(1), key(2)},
		domain.Command{ID: "first", ActorKey: key(1), ReasonCode: "member_watered", At: now})

	for _, c := range []struct {
		name        string
		counterpart string
		keyed       string
	}{
		{"a stranger", "member-c", key(3)},
		{"the actor themselves", "member-b", key(2)},
	} {
		t.Run(c.name, func(t *testing.T) {
			r.EXPECT().Find(gomock.Any(), "water-1").Return(w, nil)
			a.EXPECT().Require(gomock.Any(), "member-b", "seed.water.mutual", "water-1")
			// Naming yourself keys the same id twice, so the keyer answers
			// by lookup rather than by a fixed number of calls.
			keys := map[string]string{"member-b": key(2), c.counterpart: c.keyed}
			k.EXPECT().Key("seed-water:member", gomock.Any()).DoAndReturn(
				func(_, value string) (string, error) { return keys[value], nil }).AnyTimes()
			s := NewService(r, a, consent, k, fixedID("water-x"), fixedID("room"), func() time.Time { return now })
			if _, e := s.Water(context.Background(), Command{
				ID: "second", WaterID: "water-1", ActorID: "member-b",
				CounterpartID: c.counterpart, ReasonCode: "member_watered", ExpectedRevision: 1,
			}); !errors.Is(e, ErrNotAvailable) {
				t.Fatalf("e = %v, want ErrNotAvailable", e)
			}
		})
	}
	if consent.calls != 0 {
		t.Fatal("consent was asked about a pair this water is not between")
	}
}
