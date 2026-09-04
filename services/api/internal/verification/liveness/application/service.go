package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/domain"
)

var (
	ErrLivenessFailed     = errors.New("liveness assessment failed")
	ErrManualReviewNeeded = errors.New("liveness assessment requires manual review")
	ErrServiceUnavailable = errors.New("liveness service unavailable")
)

const defaultProviderAttempts = 2
const defaultProviderTimeout = 10 * time.Second

type Service struct {
	store    AttemptStore
	provider Provider
	reviews  ManualReviewQueue
	keyer    Keyer
	ids      IDSource
	now      func() time.Time
	attempts int
	timeout  time.Duration
}

func NewService(store AttemptStore, provider Provider, reviews ManualReviewQueue, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{
		store: store, provider: provider, reviews: reviews, keyer: keyer,
		ids: ids, now: now, attempts: defaultProviderAttempts, timeout: defaultProviderTimeout,
	}
}

type SubmitRequest struct {
	CommandID        string
	SubjectID        string
	VoiceArtifactRef string
	FaceArtifactRef  string
}

type Result struct {
	Attempt  domain.Attempt
	Replayed bool
}

// LatestAttempt reports where a member's liveness check stands.
//
// The camera step is the last of four, and it is the one most likely to be
// abandoned mid-way — a denied permission, a flat battery, a tab closed while
// a reviewer looks at the capture. Reading the standing attempt back is what
// lets the member return to the answer instead of to the beginning.
func (service Service) LatestAttempt(ctx context.Context, subjectID string) (domain.Attempt, error) {
	if service.store == nil || service.keyer == nil {
		return domain.Attempt{}, ErrServiceUnavailable
	}
	subjectKey, err := service.keyer.Key(strings.TrimSpace(subjectID))
	if err != nil {
		return domain.Attempt{}, ErrServiceUnavailable
	}
	return service.store.LatestBySubjectKey(ctx, subjectKey)
}

func (service Service) Submit(ctx context.Context, request SubmitRequest) (Result, error) {
	if service.store == nil || service.provider == nil || service.reviews == nil ||
		service.keyer == nil || service.ids == nil || service.now == nil {
		return Result{}, ErrServiceUnavailable
	}
	request.CommandID = strings.TrimSpace(request.CommandID)
	request.SubjectID = strings.TrimSpace(request.SubjectID)
	request.VoiceArtifactRef = strings.TrimSpace(request.VoiceArtifactRef)
	request.FaceArtifactRef = strings.TrimSpace(request.FaceArtifactRef)
	subjectKey, err := service.keyer.Key(request.SubjectID)
	if err != nil {
		return Result{}, ErrServiceUnavailable
	}
	inputKey, err := service.keyer.Key(request.VoiceArtifactRef + "\x00" + request.FaceArtifactRef)
	if err != nil {
		return Result{}, ErrServiceUnavailable
	}
	attempt, err := domain.NewAttempt(
		service.ids.NewID(), request.CommandID, subjectKey, inputKey, service.now().UTC(),
	)
	if err != nil {
		return Result{}, err
	}
	attempt, replayed, err := service.store.Create(ctx, attempt)
	if err != nil {
		if errors.Is(err, ErrCommandConflict) {
			return Result{}, ErrCommandConflict
		}
		return Result{}, ErrServiceUnavailable
	}
	if replayed {
		if attempt.SubjectKey() != subjectKey || attempt.InputKey() != inputKey {
			return Result{}, ErrCommandConflict
		}
		if attempt.Terminal() {
			return resultFor(attempt, true)
		}
		if attempt.Status() == domain.StatusQueuedManual {
			queueErr := service.reviews.Enqueue(ctx, ReviewTask{
				AttemptID: attempt.ID(), VoiceArtifactRef: request.VoiceArtifactRef,
				FaceArtifactRef: request.FaceArtifactRef, Reason: attempt.Reason(),
			})
			if queueErr != nil {
				return Result{Attempt: attempt, Replayed: true},
					errors.Join(ErrManualReviewNeeded, ErrReviewQueueUnavailable)
			}
			return Result{Attempt: attempt, Replayed: true}, ErrManualReviewNeeded
		}
	}

	providerRequest := ProviderRequest{
		CommandID: request.CommandID, AttemptID: attempt.ID(),
		VoiceArtifactRef: request.VoiceArtifactRef, FaceArtifactRef: request.FaceArtifactRef,
	}
	providerResult, providerErr := service.assessProvider(ctx, providerRequest)
	switch {
	case providerErr != nil:
		return service.queue(ctx, attempt, request, domain.ReasonProviderUnavailable, replayed)
	case providerResult.Outcome == OutcomeLive:
		updated, transitionErr := attempt.ProviderDecision(
			true, providerResult.ProviderRef, subjectKey, service.now().UTC(), attempt.Version(),
		)
		if transitionErr != nil {
			return service.queue(ctx, attempt, request, domain.ReasonProviderUncertain, replayed)
		}
		return service.persistDecision(ctx, updated, attempt.Version(), replayed)
	case providerResult.Outcome == OutcomeNotLive:
		updated, transitionErr := attempt.ProviderDecision(
			false, providerResult.ProviderRef, subjectKey, service.now().UTC(), attempt.Version(),
		)
		if transitionErr != nil {
			return service.queue(ctx, attempt, request, domain.ReasonProviderUncertain, replayed)
		}
		result, persistErr := service.persistDecision(ctx, updated, attempt.Version(), replayed)
		if persistErr != nil {
			return result, persistErr
		}
		return result, ErrLivenessFailed
	default:
		return service.queue(ctx, attempt, request, domain.ReasonProviderUncertain, replayed)
	}
}

func (service Service) assessProvider(ctx context.Context, request ProviderRequest) (ProviderResult, error) {
	var result ProviderResult
	var err error
	for attempt := 0; attempt < service.attempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, service.timeout)
		result, err = service.provider.Assess(attemptCtx, request)
		cancel()
		if err == nil {
			return result, nil
		}
		if ctx.Err() != nil {
			break
		}
	}
	return ProviderResult{}, err
}

func (service Service) queue(ctx context.Context, attempt domain.Attempt, request SubmitRequest, reason domain.Reason, replayed bool) (Result, error) {
	updated, err := attempt.QueueManual(reason, attempt.SubjectKey(), service.now().UTC(), attempt.Version())
	if err != nil {
		return Result{}, err
	}
	if err := service.store.Update(ctx, updated, attempt.Version()); err != nil {
		return service.resolveConflict(ctx, attempt.ID(), replayed)
	}
	result := Result{Attempt: updated, Replayed: replayed}
	if err := service.reviews.Enqueue(ctx, ReviewTask{
		AttemptID: updated.ID(), VoiceArtifactRef: request.VoiceArtifactRef,
		FaceArtifactRef: request.FaceArtifactRef, Reason: reason,
	}); err != nil {
		return result, errors.Join(ErrManualReviewNeeded, ErrReviewQueueUnavailable)
	}
	return result, ErrManualReviewNeeded
}

func (service Service) ManualDecision(ctx context.Context, attemptID, reviewerID string, passed bool) (Result, error) {
	if service.store == nil || service.reviews == nil || service.keyer == nil || service.now == nil {
		return Result{}, ErrServiceUnavailable
	}
	attempt, err := service.store.FindByID(ctx, strings.TrimSpace(attemptID))
	if err != nil {
		return Result{}, ErrAttemptNotFound
	}
	if attempt.Terminal() &&
		(attempt.Reason() == domain.ReasonManualPass || attempt.Reason() == domain.ReasonManualFail) {
		if err := service.reviews.Complete(ctx, attempt.ID()); err != nil {
			return Result{Attempt: attempt, Replayed: true}, ErrReviewQueueUnavailable
		}
		return resultFor(attempt, true)
	}
	reviewerKey, err := service.keyer.Key(strings.TrimSpace(reviewerID))
	if err != nil {
		return Result{}, ErrServiceUnavailable
	}
	updated, err := attempt.ManualDecision(passed, reviewerKey, service.now().UTC(), attempt.Version())
	if err != nil {
		return Result{}, err
	}
	if err := service.store.Update(ctx, updated, attempt.Version()); err != nil {
		return service.resolveConflict(ctx, attempt.ID(), false)
	}
	result := Result{Attempt: updated}
	if err := service.reviews.Complete(ctx, attempt.ID()); err != nil {
		return result, ErrReviewQueueUnavailable
	}
	if !passed {
		return result, ErrLivenessFailed
	}
	return result, nil
}

func (service Service) persistDecision(ctx context.Context, updated domain.Attempt, expectedVersion uint64, replayed bool) (Result, error) {
	if err := service.store.Update(ctx, updated, expectedVersion); err != nil {
		return service.resolveConflict(ctx, updated.ID(), replayed)
	}
	return Result{Attempt: updated, Replayed: replayed}, nil
}

func (service Service) resolveConflict(ctx context.Context, attemptID string, replayed bool) (Result, error) {
	current, err := service.store.FindByID(ctx, attemptID)
	if err != nil {
		return Result{}, ErrServiceUnavailable
	}
	if current.Terminal() {
		return resultFor(current, true)
	}
	if current.Status() == domain.StatusQueuedManual {
		return Result{Attempt: current, Replayed: true}, ErrManualReviewNeeded
	}
	return Result{Attempt: current, Replayed: replayed}, ErrOptimisticConflict
}

func resultFor(attempt domain.Attempt, replayed bool) (Result, error) {
	result := Result{Attempt: attempt, Replayed: replayed}
	if attempt.Status() == domain.StatusFailed {
		return result, ErrLivenessFailed
	}
	return result, nil
}
