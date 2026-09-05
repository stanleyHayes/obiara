package invariants_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	listening "github.com/stanleyHayes/obiara/services/api/internal/seed/listening/domain"
	sowapp "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
	sow "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

type screeningStub struct{ approved bool }

func (s screeningStub) Screen(context.Context, string, []string) (sowapp.ScreeningDecision, error) {
	return sowapp.ScreeningDecision{Approved: s.approved, Reference: "screened"}, nil
}

type keyerStub struct{}

func (keyerStub) Key(namespace, value string) (string, error) {
	sum := sha256.Sum256([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(sum[:]), nil
}

type idsStub struct{ next atomic.Uint64 }

func (i *idsStub) NewID() string { return fmt.Sprintf("sow-%d", i.next.Add(1)) }

type acceptanceModel struct {
	mu       sync.Mutex
	balance  int64
	accepted map[string]sow.Sow
}

// The invariants exercised here are about accepting sows concurrently, not
// about settling reviews, so both review-side methods refuse rather than
// model behaviour nothing in this file drives.
func (a *acceptanceModel) FindByScreening(context.Context, string) (sow.Sow, error) {
	return sow.Sow{}, sowapp.ErrSowNotFound
}

func (a *acceptanceModel) Settle(context.Context, sow.Sow, bool) error {
	return sowapp.ErrSowNotFound
}

func (a *acceptanceModel) Accept(_ context.Context, candidate sow.Sow) (sow.Sow, bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if existing, ok := a.accepted[candidate.CommandID]; ok {
		if existing.Fingerprint != candidate.Fingerprint {
			return sow.Sow{}, false, sow.ErrCommandMismatch
		}
		return existing, true, nil
	}
	if a.balance < candidate.AllowanceUnits {
		return sow.Sow{}, false, sowapp.ErrInsufficientAllowance
	}
	a.balance -= candidate.AllowanceUnits
	a.accepted[candidate.CommandID] = candidate
	return candidate, false, nil
}

func TestNoAllowanceEffectBeforeConfirmedScreenedAcceptance(t *testing.T) {
	model := &acceptanceModel{balance: 2, accepted: map[string]sow.Sow{}}
	ids := &idsStub{}
	command := sowapp.Command{ID: "send-1", ActorID: "actor", Body: "hello", Confirmed: true}
	rejected := sowapp.New(screeningStub{false}, model, keyerStub{}, ids, time.Now, 1)
	if _, err := rejected.Send(context.Background(), command); !errors.Is(err, sow.ErrScreeningRejected) {
		t.Fatalf("rejected send=%v", err)
	}
	if model.balance != 2 || len(model.accepted) != 0 {
		t.Fatalf("screening rejection spent allowance: balance=%d", model.balance)
	}
	command.Confirmed = false
	approved := sowapp.New(screeningStub{true}, model, keyerStub{}, ids, time.Now, 1)
	if _, err := approved.Send(context.Background(), command); !errors.Is(err, sow.ErrNotConfirmed) {
		t.Fatalf("unconfirmed send=%v", err)
	}
	if model.balance != 2 {
		t.Fatal("unconfirmed send spent allowance")
	}
	command.Confirmed = true
	if _, err := approved.Send(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	if model.balance != 1 {
		t.Fatalf("accepted send balance=%d", model.balance)
	}
	if result, err := approved.Send(context.Background(), command); err != nil || !result.Replayed {
		t.Fatalf("replay=%v err=%v", result.Replayed, err)
	}
	if model.balance != 1 {
		t.Fatal("replay double-spent allowance")
	}
}

func TestConcurrentCommandReplayHasOneAcceptanceEffect(t *testing.T) {
	model := &acceptanceModel{balance: 10, accepted: map[string]sow.Sow{}}
	service := sowapp.New(screeningStub{true}, model, keyerStub{}, &idsStub{}, time.Now, 1)
	command := sowapp.Command{ID: "same-command", ActorID: "actor", Body: "hello", Confirmed: true}
	var wg sync.WaitGroup
	errs := make(chan error, 32)
	for range 32 {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := service.Send(context.Background(), command); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if model.balance != 9 || len(model.accepted) != 1 {
		t.Fatalf("effects=%d balance=%d", len(model.accepted), model.balance)
	}
}

func FuzzListeningNeverBecomesEligibleFromReplayOverlap(f *testing.F) {
	f.Add(19.9, 10.0)
	f.Fuzz(func(t *testing.T, end, duplicate float64) {
		if end <= 0 || end >= listening.RequiredSeconds || duplicate <= 0 {
			t.Skip()
		}
		playback, err := listening.NewPlayback("listener", "asset", 120)
		if err != nil {
			t.Fatal(err)
		}
		for range 20 {
			if err = playback.Record(0, end); err != nil {
				t.Fatal(err)
			}
			if err = playback.Record(0, min(end, duplicate)); err != nil {
				t.Fatal(err)
			}
		}
		if playback.Eligible() || playback.TotalSeconds() >= listening.RequiredSeconds {
			t.Fatalf("overlap armed early: %f", playback.TotalSeconds())
		}
	})
}
