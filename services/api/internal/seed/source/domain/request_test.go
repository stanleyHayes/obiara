package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"testing"
	"testing/quick"
	"time"
)

var testNow = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func key(value int) string { return fmt.Sprintf("%064x", value) }
func command(id string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorKey: key(1), ExpectedRevision: revision, ReasonCode: "user_requested", At: at}
}
func validRequest(t *testing.T, candidates []string) Request {
	t.Helper()
	r, err := Open("request-1", key(1), Source{Type: SourceCircle, Key: key(2)}, candidates, testNow.Add(time.Hour), command("open-1", 0, testNow))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestOpenRequiresExplicitSourceAndBoundedDeterministicCandidates(t *testing.T) {
	if _, err := Open("request-1", key(1), Source{Type: "global", Key: key(2)}, nil, testNow.Add(time.Hour), command("open-1", 0, testNow)); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("global source accepted: %v", err)
	}
	values := []string{key(3), key(2), key(3)}
	got := validRequest(t, values).CandidateIDs()
	want := []string{key(2), key(3)}
	if !slices.Equal(got, want) {
		t.Fatalf("candidates = %v, want %v", got, want)
	}
	tooMany := make([]string, MaxCandidates+1)
	for i := range tooMany {
		tooMany[i] = key(i + 10)
	}
	if _, err := Open("request-1", key(1), Source{Type: SourceCircle, Key: key(2)}, tooMany, testNow.Add(time.Hour), command("open-1", 0, testNow)); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("oversized candidates accepted: %v", err)
	}
}

func TestWithdrawAndExpiryAreReplaySafe(t *testing.T) {
	opened := validRequest(t, nil)
	withdraw := command("withdraw-1", 1, testNow.Add(time.Minute))
	first, err := opened.Withdraw(withdraw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := first.Withdraw(withdraw)
	if err != nil || replayed.Revision() != first.Revision() {
		t.Fatalf("replay: revision=%d err=%v", replayed.Revision(), err)
	}
	if _, err := opened.Expire(command("expire-early", 1, testNow.Add(59*time.Minute))); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("early expiry accepted: %v", err)
	}
	expired, err := opened.Expire(command("expire-1", 1, testNow.Add(time.Hour)))
	if err != nil || expired.Status() != StatusExpired {
		t.Fatalf("expiry: status=%s err=%v", expired.Status(), err)
	}
}

func TestTransitionHistoryRoundTripsProperty(t *testing.T) {
	property := func(seed uint64, withdraw bool) bool {
		random := rand.New(rand.NewSource(int64(seed)))
		n := random.Intn(MaxCandidates + 1)
		candidates := make([]string, n)
		for i := range candidates {
			candidates[i] = key(i + 20)
		}
		opened := validRequest(t, candidates)
		var next Request
		var err error
		if withdraw {
			next, err = opened.Withdraw(command("withdraw-1", 1, testNow.Add(time.Minute)))
		} else {
			next, err = opened.Expire(command("expire-1", 1, testNow.Add(time.Hour)))
		}
		if err != nil {
			return false
		}
		rehydrated, err := Rehydrate(State{ID: next.ID(), RequesterKey: next.RequesterKey(), Source: next.Source(), CandidateIDs: next.CandidateIDs(), Status: next.Status(), ExpiresAt: next.ExpiresAt(), EndedAt: next.EndedAt(), Revision: next.Revision(), Events: next.Events(), Commands: next.Commands()})
		return err == nil && rehydrated.Status() == next.Status() && slices.Equal(rehydrated.CandidateIDs(), next.CandidateIDs())
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatal(err)
	}
}
