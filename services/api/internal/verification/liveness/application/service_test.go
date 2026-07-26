package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
	"go.uber.org/mock/gomock"
)

var serviceTime = time.Date(2026, 7, 26, 18, 0, 0, 0, time.UTC)

func submitRequest() SubmitRequest {
	return SubmitRequest{
		CommandID: "command:1", SubjectID: "member:1",
		VoiceArtifactRef: "voice:1", FaceArtifactRef: "face:1",
	}
}

func pendingAttempt(t *testing.T) domain.Attempt {
	t.Helper()
	attempt, err := domain.NewAttempt(
		"attempt:1", "command:1", strings.Repeat("a", 64),
		strings.Repeat("b", 64), serviceTime,
	)
	if err != nil {
		t.Fatal(err)
	}
	return attempt
}

func expectNewAttempt(store *MockAttemptStore, keyer *MockKeyer, ids *MockIDSource, t *testing.T) domain.Attempt {
	t.Helper()
	keyer.EXPECT().Key("member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("voice:1\x00face:1").Return(strings.Repeat("b", 64), nil)
	ids.EXPECT().NewID().Return("attempt:1")
	attempt := pendingAttempt(t)
	store.EXPECT().Create(gomock.Any(), attempt).Return(attempt, false, nil)
	return attempt
}

func TestLiveProviderIsTheOnlyAutomaticPass(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	provider := NewMockProvider(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	attempt := expectNewAttempt(store, keyer, ids, t)
	provider.EXPECT().Assess(gomock.Any(), ProviderRequest{
		CommandID: "command:1", AttemptID: "attempt:1",
		VoiceArtifactRef: "voice:1", FaceArtifactRef: "face:1",
	}).Return(ProviderResult{Outcome: OutcomeLive, ProviderRef: "provider:proof"}, nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(updated domain.Attempt) bool {
		return updated.Passed() && updated.ProviderRef() == "provider:proof" &&
			len(updated.Events()) == 1
	}), attempt.Version()).Return(nil)

	result, err := NewService(
		store, provider, NewMockManualReviewQueue(controller), keyer, ids,
		func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if err != nil || !result.Attempt.Passed() {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
}

func TestNotLiveFailsWithoutManualOrPass(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	provider := NewMockProvider(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	attempt := expectNewAttempt(store, keyer, ids, t)
	provider.EXPECT().Assess(gomock.Any(), gomock.Any()).
		Return(ProviderResult{Outcome: OutcomeNotLive, ProviderRef: "provider:proof"}, nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(updated domain.Attempt) bool {
		return updated.Status() == domain.StatusFailed && !updated.Passed()
	}), attempt.Version()).Return(nil)

	result, err := NewService(
		store, provider, NewMockManualReviewQueue(controller), keyer, ids,
		func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if result.Attempt.Passed() || !errors.Is(err, ErrLivenessFailed) {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
}

func TestProviderOutageRetriesThenQueuesManual(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	provider := NewMockProvider(controller)
	reviews := NewMockManualReviewQueue(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	attempt := expectNewAttempt(store, keyer, ids, t)
	provider.EXPECT().Assess(gomock.Any(), gomock.Any()).
		Return(ProviderResult{}, context.DeadlineExceeded).Times(defaultProviderAttempts)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(updated domain.Attempt) bool {
		return updated.Status() == domain.StatusQueuedManual &&
			updated.Reason() == domain.ReasonProviderUnavailable && !updated.Passed()
	}), attempt.Version()).Return(nil)
	reviews.EXPECT().Enqueue(gomock.Any(), ReviewTask{
		AttemptID: "attempt:1", VoiceArtifactRef: "voice:1",
		FaceArtifactRef: "face:1", Reason: domain.ReasonProviderUnavailable,
	}).Return(nil)

	result, err := NewService(
		store, provider, reviews, keyer, ids, func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if result.Attempt.Passed() || !errors.Is(err, ErrManualReviewNeeded) {
		t.Fatalf("outage silently passed: %+v, %v", result, err)
	}
}

func TestUnknownProviderOutcomeQueuesUncertain(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	provider := NewMockProvider(controller)
	reviews := NewMockManualReviewQueue(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	attempt := expectNewAttempt(store, keyer, ids, t)
	provider.EXPECT().Assess(gomock.Any(), gomock.Any()).
		Return(ProviderResult{Outcome: "vendor_added_value"}, nil)
	store.EXPECT().Update(gomock.Any(), gomock.Any(), attempt.Version()).Return(nil)
	reviews.EXPECT().Enqueue(gomock.Any(), gomock.Cond(func(task ReviewTask) bool {
		return task.Reason == domain.ReasonProviderUncertain
	})).Return(nil)

	result, err := NewService(
		store, provider, reviews, keyer, ids, func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if result.Attempt.Passed() || !errors.Is(err, ErrManualReviewNeeded) {
		t.Fatalf("unknown outcome silently passed: %+v, %v", result, err)
	}
}

func TestMalformedLiveProofQueuesUncertainInsteadOfPassing(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	provider := NewMockProvider(controller)
	reviews := NewMockManualReviewQueue(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	attempt := expectNewAttempt(store, keyer, ids, t)
	provider.EXPECT().Assess(gomock.Any(), gomock.Any()).
		Return(ProviderResult{Outcome: OutcomeLive, ProviderRef: ""}, nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(updated domain.Attempt) bool {
		return updated.Status() == domain.StatusQueuedManual &&
			updated.Reason() == domain.ReasonProviderUncertain
	}), attempt.Version()).Return(nil)
	reviews.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Return(nil)

	result, err := NewService(
		store, provider, reviews, keyer, ids, func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if result.Attempt.Passed() || !errors.Is(err, ErrManualReviewNeeded) {
		t.Fatalf("malformed live proof passed: %+v, %v", result, err)
	}
}

func TestQueuedReplayRepairsReviewQueueWithoutProviderCall(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	reviews := NewMockManualReviewQueue(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	keyer.EXPECT().Key("member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("voice:1\x00face:1").Return(strings.Repeat("b", 64), nil)
	ids.EXPECT().NewID().Return("attempt:new")
	attempt := pendingAttempt(t)
	queued, _ := attempt.QueueManual(
		domain.ReasonProviderUncertain, strings.Repeat("a", 64), serviceTime, 1,
	)
	store.EXPECT().Create(gomock.Any(), gomock.Any()).Return(queued, true, nil)
	reviews.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Return(nil)

	result, err := NewService(
		store, NewMockProvider(controller), reviews, keyer, ids,
		func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if !result.Replayed || !errors.Is(err, ErrManualReviewNeeded) {
		t.Fatalf("queued replay=%+v, %v", result, err)
	}
}

func TestIdempotencyConflictRejectsChangedArtifacts(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	keyer.EXPECT().Key("member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("voice:1\x00face:1").Return(strings.Repeat("c", 64), nil)
	ids.EXPECT().NewID().Return("attempt:new")
	store.EXPECT().Create(gomock.Any(), gomock.Any()).Return(pendingAttempt(t), true, nil)

	_, err := NewService(
		store, NewMockProvider(controller), NewMockManualReviewQueue(controller),
		keyer, ids, func() time.Time { return serviceTime },
	).Submit(context.Background(), submitRequest())
	if !errors.Is(err, ErrCommandConflict) {
		t.Fatalf("expected command conflict, got %v", err)
	}
}

func TestManualReviewDecisionDeletesTemporaryTask(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockAttemptStore(controller)
	reviews := NewMockManualReviewQueue(controller)
	keyer := NewMockKeyer(controller)
	attempt := pendingAttempt(t)
	queued, _ := attempt.QueueManual(
		domain.ReasonProviderUncertain, strings.Repeat("a", 64), serviceTime, 1,
	)
	store.EXPECT().FindByID(gomock.Any(), "attempt:1").Return(queued, nil)
	keyer.EXPECT().Key("reviewer:1").Return(strings.Repeat("d", 64), nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(updated domain.Attempt) bool {
		return updated.Passed() && updated.Reason() == domain.ReasonManualPass
	}), queued.Version()).Return(nil)
	reviews.EXPECT().Complete(gomock.Any(), "attempt:1").Return(nil)

	result, err := NewService(
		store, NewMockProvider(controller), reviews, keyer,
		NewMockIDSource(controller), func() time.Time { return serviceTime.Add(time.Minute) },
	).ManualDecision(context.Background(), "attempt:1", "reviewer:1", true)
	if err != nil || !result.Attempt.Passed() {
		t.Fatalf("manual result=%+v, %v", result, err)
	}
}
