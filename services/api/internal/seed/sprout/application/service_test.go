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
	service := New(repository, keyer, ids, time.Now).WithListenGate(listen)
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

func sproutWith(t *testing.T, gate ListenGate) (SproutResult, error) {
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
