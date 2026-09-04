package application_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/adapters/outbound/simulator"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

type memoryStore struct {
	mu       sync.Mutex
	attempts map[string]domain.Attempt
	commands map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		attempts: make(map[string]domain.Attempt),
		commands: make(map[string]string),
	}
}

func (store *memoryStore) Create(_ context.Context, attempt domain.Attempt) (domain.Attempt, bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if id, exists := store.commands[attempt.CommandID()]; exists {
		return store.attempts[id], true, nil
	}
	store.attempts[attempt.ID()] = attempt
	store.commands[attempt.CommandID()] = attempt.ID()
	return attempt, false, nil
}

func (store *memoryStore) LatestBySubjectKey(_ context.Context, subjectKey string) (domain.Attempt, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	var latest domain.Attempt
	var found bool
	for _, attempt := range store.attempts {
		if attempt.SubjectKey() != subjectKey {
			continue
		}
		if !found || attempt.CreatedAt().After(latest.CreatedAt()) {
			latest, found = attempt, true
		}
	}
	if !found {
		return domain.Attempt{}, application.ErrAttemptNotFound
	}
	return latest, nil
}

func (store *memoryStore) FindByID(_ context.Context, id string) (domain.Attempt, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	attempt, exists := store.attempts[id]
	if !exists {
		return domain.Attempt{}, application.ErrAttemptNotFound
	}
	return attempt, nil
}

func (store *memoryStore) Update(_ context.Context, attempt domain.Attempt, expectedVersion uint64) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.attempts[attempt.ID()]
	if !exists {
		return application.ErrAttemptNotFound
	}
	if current.Version() != expectedVersion {
		return application.ErrOptimisticConflict
	}
	store.attempts[attempt.ID()] = attempt
	return nil
}

type memoryReviews struct {
	mu    sync.Mutex
	tasks map[string]application.ReviewTask
}

func (reviews *memoryReviews) Enqueue(_ context.Context, task application.ReviewTask) error {
	reviews.mu.Lock()
	defer reviews.mu.Unlock()
	reviews.tasks[task.AttemptID] = task
	return nil
}

func (reviews *memoryReviews) Complete(_ context.Context, attemptID string) error {
	reviews.mu.Lock()
	defer reviews.mu.Unlock()
	delete(reviews.tasks, attemptID)
	return nil
}

type sequenceIDs struct{ id string }

func (ids sequenceIDs) NewID() string { return ids.id }

func TestSimulatorUncertaintyToManualDecisionEndToEnd(t *testing.T) {
	store := newMemoryStore()
	reviews := &memoryReviews{tasks: make(map[string]application.ReviewTask)}
	provider := simulator.NewProvider()
	keyer, err := privacy.NewHMACKeyer([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 26, 18, 0, 0, 0, time.UTC)
	service := application.NewService(
		store, provider, reviews, keyer, sequenceIDs{id: "attempt:integration"},
		func() time.Time { return now },
	)
	request := application.SubmitRequest{
		CommandID: "command:integration", SubjectID: "member:integration",
		VoiceArtifactRef: "voice:uncertain", FaceArtifactRef: "face:artifact",
	}
	result, err := service.Submit(context.Background(), request)
	if !errors.Is(err, application.ErrManualReviewNeeded) ||
		result.Attempt.Status() != domain.StatusQueuedManual || result.Attempt.Passed() {
		t.Fatalf("uncertain result=%+v, %v", result, err)
	}
	task, exists := reviews.tasks[result.Attempt.ID()]
	if !exists || task.VoiceArtifactRef != request.VoiceArtifactRef {
		t.Fatal("temporary manual-review task missing")
	}
	if result.Attempt.SubjectKey() == request.SubjectID ||
		result.Attempt.InputKey() == request.VoiceArtifactRef ||
		result.Attempt.ProviderRef() != "" {
		t.Fatal("retained proof leaked raw liveness input")
	}

	// Exact replay repairs/idempotently refreshes the queue without invoking
	// the provider a second time.
	replayed, replayErr := service.Submit(context.Background(), request)
	if !replayed.Replayed || !errors.Is(replayErr, application.ErrManualReviewNeeded) ||
		len(provider.Requests()) != 1 {
		t.Fatalf("replay=%+v, %v; provider requests=%d", replayed, replayErr, len(provider.Requests()))
	}

	now = now.Add(time.Minute)
	decided, err := service.ManualDecision(
		context.Background(), result.Attempt.ID(), "reviewer:1", true,
	)
	if err != nil || !decided.Attempt.Passed() || decided.Attempt.Reason() != domain.ReasonManualPass {
		t.Fatalf("manual decision=%+v, %v", decided, err)
	}
	if _, exists := reviews.tasks[result.Attempt.ID()]; exists {
		t.Fatal("temporary biometric review task survived decision")
	}
}

func TestSimulatorOutageNeverSilentlyPasses(t *testing.T) {
	store := newMemoryStore()
	reviews := &memoryReviews{tasks: make(map[string]application.ReviewTask)}
	provider := simulator.NewProvider()
	keyer, _ := privacy.NewHMACKeyer([]byte(strings.Repeat("k", 32)))
	service := application.NewService(
		store, provider, reviews, keyer, sequenceIDs{id: "attempt:outage"}, time.Now,
	)
	result, err := service.Submit(context.Background(), application.SubmitRequest{
		CommandID: "command:outage", SubjectID: "member:outage",
		VoiceArtifactRef: "voice:outage", FaceArtifactRef: "face:artifact",
	})
	if result.Attempt.Passed() || !errors.Is(err, application.ErrManualReviewNeeded) ||
		result.Attempt.Reason() != domain.ReasonProviderUnavailable ||
		len(provider.Requests()) != 2 {
		t.Fatalf("outage result=%+v, %v requests=%d", result, err, len(provider.Requests()))
	}
}
