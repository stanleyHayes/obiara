package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sprout/domain"
	"go.uber.org/mock/gomock"
)

func TestUnilateralSproutReturnsNoDoorway(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	keyer.EXPECT().Key("participant", "alice").Return("alice-key", nil)
	keyer.EXPECT().Key("participant", "bob").Return("bob-key", nil)
	keyer.EXPECT().Key("seed", "seed-raw").Return("seed-key", nil)
	ids.EXPECT().NewID().Return("intent-1")
	repository.EXPECT().RecordIntent(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, intent domain.Intent) (*domain.Doorway, bool, error) {
		if intent.ActorKey != "alice-key" || intent.TargetKey != "bob-key" {
			t.Fatalf("intent=%#v", intent)
		}
		return nil, false, nil
	})
	listen := NewMockListenGate(ctrl)
	listen.EXPECT().Heard(gomock.Any(), "alice", "bob").Return(true, nil)
	allowance := NewMockAllowance(ctrl)
	allowance.EXPECT().Spend(gomock.Any(), "alice", "command").Return(nil)
	declines := NewMockDeclineLock(ctrl)
	declines.EXPECT().Locked(gomock.Any(), "alice", "bob").Return(false, nil)
	service := New(repository, keyer, ids, time.Now).
		WithListenGate(listen).WithAllowance(allowance).WithDeclineLock(declines)
	result, err := service.Sprout(context.Background(), SproutCommand{"command", "alice", "bob", "seed-raw"})
	if err != nil || result.Doorway != nil {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestExchangeUsesOpaqueReferences(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Now()
	current, _ := domain.Open("doorway", "alice-key", "bob-key", now)
	repository.EXPECT().FindDoorway(gomock.Any(), "doorway").Return(current, nil)
	keyer.EXPECT().Key("participant", "alice").Return("alice-key", nil)
	keyer.EXPECT().Key("message", "raw-message").Return("message-key", nil)
	repository.EXPECT().AppendExchange(gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(func(_ context.Context, next domain.Doorway, _ uint64) (domain.Doorway, bool, error) {
		if next.Exchanges()[0].MessageKey != "message-key" {
			t.Fatal("raw ref persisted")
		}
		return next, false, nil
	})
	service := New(repository, keyer, ids, func() time.Time { return now })
	result, err := service.Exchange(context.Background(), ExchangeCommand{"exchange", "doorway", "alice", "raw-message"})
	if err != nil || len(result.Doorway.Exchanges()) != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

// heardGate is a listen gate with a fixed answer, for the tests that are
// about what the gate decides rather than how it decides it.
type heardGate struct {
	heard bool
	err   error
}

func (g heardGate) Heard(context.Context, string, string) (bool, error) { return g.heard, g.err }

// payingAllowance is an allowance with a fixed answer.
type payingAllowance struct{ err error }

func (a payingAllowance) Spend(context.Context, string, string) error { return a.err }

// openLock is a decline lock that shields nobody.
type openLock struct {
	locked bool
	err    error
}

func (l openLock) Locked(context.Context, string, string) (bool, error) { return l.locked, l.err }

func sproutWith(t *testing.T, gate ListenGate) (SproutResult, error) {
	return sproutPaying(t, gate, payingAllowance{})
}

func sproutPaying(t *testing.T, gate ListenGate, allowance Allowance) (SproutResult, error) {
	return sproutFull(t, gate, allowance, openLock{})
}

func sproutFull(t *testing.T, gate ListenGate, allowance Allowance, lock DeclineLock) (SproutResult, error) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	// No expectations set on purpose: if the gate refuses, the controller
	// fails the test should anything be keyed, minted or written. That is
	// the assertion — a refused sow leaves no trace.
	// Distinct per value: one key for everybody would make the actor and the
	// target the same person, which the aggregate rightly refuses, and the
	// test would pass for the wrong reason.
	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).DoAndReturn(
		func(namespace, value string) (string, error) { return namespace + "-" + value, nil },
	).AnyTimes()
	ids.EXPECT().NewID().Return("intent-1").AnyTimes()
	repository.EXPECT().RecordIntent(gomock.Any(), gomock.Any()).Return(nil, false, nil).AnyTimes()

	service := New(repository, keyer, ids, time.Now)
	if gate != nil {
		service = service.WithListenGate(gate)
	}
	if allowance != nil {
		service = service.WithAllowance(allowance)
	}
	if lock != nil {
		service = service.WithDeclineLock(lock)
	}
	return service.Sprout(context.Background(), SproutCommand{"command", "alice", "bob", "seed-raw"})
}

func TestASowNeedsTheirVoiceToHaveBeenHeard(t *testing.T) {
	// FR-202. Reaching for someone you have not listened to is the thing
	// this product exists not to be, and until now nothing stopped it.
	if _, err := sproutWith(t, heardGate{heard: false}); !errors.Is(err, ErrNotHeard) {
		t.Fatalf("an unheard sow returned %v, want ErrNotHeard", err)
	}
	if _, err := sproutWith(t, heardGate{heard: true}); err != nil {
		t.Fatalf("a heard sow was refused: %v", err)
	}
}

func TestAnUnansweredListenGateRefusesRatherThanPasses(t *testing.T) {
	// The direction that matters: an outage must not arm a sow. A nil gate
	// is the same case — a deployment with no recordings anywhere cannot
	// have anyone who has been heard.
	if _, err := sproutWith(t, heardGate{err: errors.New("listening unavailable")}); err == nil {
		t.Fatal("a sow was armed while the gate was unreachable")
	}
	if _, err := sproutWith(t, nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("a sow with no gate composed returned %v, want ErrUnavailable", err)
	}
}

func TestASowCostsASeed(t *testing.T) {
	// FR-201a. A sow that costs nothing is a swipe, and until now the
	// allowance ledger existed and nothing ever spent from it.
	if _, err := sproutPaying(t, heardGate{heard: true}, payingAllowance{err: ErrNoSeeds}); !errors.Is(err, ErrNoSeeds) {
		t.Fatalf("a sow with no seeds returned %v, want ErrNoSeeds", err)
	}
	if _, err := sproutPaying(t, heardGate{heard: true}, payingAllowance{}); err != nil {
		t.Fatalf("a paid sow was refused: %v", err)
	}
	if _, err := sproutPaying(t, heardGate{heard: true}, nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("a sow with no allowance composed returned %v, want ErrUnavailable", err)
	}
}

func TestARefusedSowIsNeverCharged(t *testing.T) {
	// Order matters. Charging before the listen gate would take a seed from
	// a member whose sow was then refused, which is the one failure here
	// that costs them something real.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	allowance := NewMockAllowance(ctrl)
	// No Spend expectation: the controller fails the test if the allowance
	// is touched at all. That is the assertion.
	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).Return("key", nil).AnyTimes()
	ids.EXPECT().NewID().Return("intent-1").AnyTimes()
	repository.EXPECT().RecordIntent(gomock.Any(), gomock.Any()).Return(nil, false, nil).AnyTimes()

	service := New(repository, keyer, ids, time.Now).
		WithListenGate(heardGate{heard: false}).WithAllowance(allowance)
	if _, err := service.Sprout(context.Background(), SproutCommand{"command", "alice", "bob", "seed-raw"}); !errors.Is(err, ErrNotHeard) {
		t.Fatalf("err = %v, want ErrNotHeard", err)
	}
}

func TestADeclineShieldsForNinetyDays(t *testing.T) {
	// M4-AC-01. Until now a member could be declined and reach again the same
	// minute, which makes the decline meaningless for the person it protects.
	if _, err := sproutFull(t, heardGate{heard: true}, payingAllowance{}, openLock{locked: true}); !errors.Is(err, ErrReachNotAvailable) {
		t.Fatalf("a shielded target returned %v, want ErrReachNotAvailable", err)
	}
	if _, err := sproutFull(t, heardGate{heard: true}, payingAllowance{}, openLock{}); err != nil {
		t.Fatalf("an unshielded sow was refused: %v", err)
	}
	if _, err := sproutFull(t, heardGate{heard: true}, payingAllowance{}, openLock{err: errors.New("store down")}); err == nil {
		t.Fatal("a sow went through while the shield could not be read")
	}
	if _, err := sproutFull(t, heardGate{heard: true}, payingAllowance{}, nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("a sow with no lock composed returned %v, want ErrUnavailable", err)
	}
}

func TestAShieldedSowIsNeverCharged(t *testing.T) {
	// Same ordering rule as the listen gate: a member must not spend a seed
	// on a sow that was never going to land.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	allowance := NewMockAllowance(ctrl)
	// No Spend expectation: touching the allowance fails the test.
	keyer.EXPECT().Key(gomock.Any(), gomock.Any()).Return("key", nil).AnyTimes()
	ids.EXPECT().NewID().Return("intent-1").AnyTimes()
	repository.EXPECT().RecordIntent(gomock.Any(), gomock.Any()).Return(nil, false, nil).AnyTimes()

	service := New(repository, keyer, ids, time.Now).
		WithListenGate(heardGate{heard: true}).
		WithAllowance(allowance).
		WithDeclineLock(openLock{locked: true})
	if _, err := service.Sprout(context.Background(), SproutCommand{"command", "alice", "bob", "seed-raw"}); !errors.Is(err, ErrReachNotAvailable) {
		t.Fatalf("err = %v, want ErrReachNotAvailable", err)
	}
}
