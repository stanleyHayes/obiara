package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/host/domain"
)

type Service struct {
	repository Repository
	provider   InstitutionalProvider
	reviews    ManualReviewQueue
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewService(repository Repository, provider InstitutionalProvider, reviews ManualReviewQueue, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{repository: repository, provider: provider, reviews: reviews, keyer: keyer, ids: ids, now: now}
}

type SubmitRequest struct {
	CommandID       string
	ApplicantID     string
	ProofReference  string
	InstitutionKind domain.InstitutionKind
	IssuerID        string
	IssuedAt        time.Time
	ExpiresAt       time.Time
}

type Result struct {
	Application domain.Application
	Replayed    bool
}

func (service Service) Submit(ctx context.Context, request SubmitRequest) (Result, error) {
	if !service.ready() {
		return Result{}, ErrDependencyUnavailable
	}
	request.CommandID = strings.TrimSpace(request.CommandID)
	applicantKey, err := service.keyer.Key("host_applicant", strings.TrimSpace(request.ApplicantID))
	if err != nil {
		return Result{}, ErrDependencyUnavailable
	}
	issuerKey, err := service.keyer.Key("institution_issuer", strings.TrimSpace(request.IssuerID))
	if err != nil {
		return Result{}, ErrDependencyUnavailable
	}
	proof, err := domain.NewProof(request.ProofReference, request.InstitutionKind, issuerKey, request.IssuedAt, request.ExpiresAt)
	if err != nil {
		return Result{}, err
	}
	now := service.now().UTC()
	application, err := domain.NewApplication(
		service.ids.NewID(), request.CommandID, applicantKey, proof,
		domain.Command{ID: request.CommandID + ".submitted", ActorKey: "system", At: now},
	)
	if err != nil {
		return Result{}, err
	}
	application, replayed, err := service.repository.Create(ctx, application)
	if err != nil {
		return Result{}, ErrDependencyUnavailable
	}
	if application.ApplicantKey() != applicantKey ||
		application.Proof().IssuerKey() != issuerKey ||
		application.Proof().Reference() != strings.TrimSpace(request.ProofReference) ||
		application.Proof().Kind() != request.InstitutionKind ||
		!application.Proof().IssuedAt().Equal(request.IssuedAt.UTC()) ||
		!application.Proof().ExpiresAt().Equal(request.ExpiresAt.UTC()) {
		return Result{}, domain.ErrCommandConflict
	}
	if replayed && application.Status() != domain.StatusSubmitted {
		return resultFor(application, true)
	}
	return service.verify(ctx, application, request.CommandID+".provider", replayed)
}

func (service Service) verify(ctx context.Context, application domain.Application, commandID string, replayed bool) (Result, error) {
	providerResult, providerErr := service.provider.Verify(ctx, ProviderRequest{
		CommandID: commandID, ApplicationID: application.ID(),
		ProofReference: application.Proof().Reference(), InstitutionKind: application.Proof().Kind(),
	})
	now := service.now().UTC()
	command := domain.Command{ID: commandID, ActorKey: "system", At: now}
	var updated domain.Application
	var err error
	switch {
	case providerErr != nil:
		updated, err = application.QueueManual(domain.ReasonProviderUnavailable, "", command, application.Version())
	case providerResult.Outcome == OutcomeVerified:
		updated, err = application.ProviderDecision(true, providerResult.ProviderRef, command, application.Version())
	case providerResult.Outcome == OutcomeRejected:
		updated, err = application.ProviderDecision(false, providerResult.ProviderRef, command, application.Version())
	default:
		updated, err = application.QueueManual(domain.ReasonProviderUncertain, providerResult.ProviderRef, command, application.Version())
	}
	if err != nil {
		// Malformed or expired provider proof is uncertainty, never approval.
		updated, err = application.QueueManual(domain.ReasonProviderUncertain, "", command, application.Version())
		if err != nil {
			return Result{}, err
		}
	}
	updated, err = service.update(ctx, application, updated, commandID)
	if err != nil {
		return Result{}, err
	}
	result := Result{Application: updated, Replayed: replayed}
	if updated.Status() == domain.StatusQueuedManual {
		if err := service.reviews.Enqueue(ctx, ReviewTask{
			ApplicationID: updated.ID(), ProofReference: updated.Proof().Reference(),
			Reason: updated.Reason(),
		}); err != nil {
			return result, errors.Join(ErrManualReviewRequired, ErrDependencyUnavailable)
		}
		return result, ErrManualReviewRequired
	}
	return resultFor(updated, replayed)
}

func (service Service) ManualDecision(ctx context.Context, applicationID, commandID, reviewerID string, approved bool) (Result, error) {
	application, err := service.find(ctx, applicationID)
	if err != nil {
		return Result{}, err
	}
	actorKey, err := service.keyer.Key("host_reviewer", strings.TrimSpace(reviewerID))
	if err != nil {
		return Result{}, ErrDependencyUnavailable
	}
	command := domain.Command{ID: strings.TrimSpace(commandID), ActorKey: actorKey, At: service.now().UTC()}
	target, reason := domain.StatusRejected, domain.ReasonManualRejected
	if approved {
		target, reason = domain.StatusApproved, domain.ReasonManualApproved
	}
	if application.HasCommand(command.ID) {
		if !application.MatchesCommand(command.ID, target, reason, actorKey) {
			return Result{}, domain.ErrCommandConflict
		}
		if err := service.reviews.Complete(ctx, application.ID()); err != nil {
			return Result{Application: application, Replayed: true}, ErrDependencyUnavailable
		}
		return resultFor(application, true)
	}
	updated, err := application.ManualDecision(approved, command, application.Version())
	if err != nil {
		return Result{}, err
	}
	updated, err = service.update(ctx, application, updated, command.ID)
	if err != nil {
		return Result{}, err
	}
	result := Result{Application: updated}
	if err := service.reviews.Complete(ctx, application.ID()); err != nil {
		return result, ErrDependencyUnavailable
	}
	return resultFor(updated, false)
}

func (service Service) RecheckDue(ctx context.Context, dueBefore time.Time, limit int) (int, error) {
	if !service.ready() || limit < 1 {
		return 0, ErrDependencyUnavailable
	}
	applications, err := service.repository.ListRecheckDue(ctx, dueBefore.UTC(), limit)
	if err != nil {
		return 0, ErrDependencyUnavailable
	}
	processed := 0
	for _, application := range applications {
		_, verifyErr := service.verify(
			ctx, application,
			"recheck:"+application.ID()+":"+service.now().UTC().Format("20060102"),
			false,
		)
		if verifyErr != nil && !errors.Is(verifyErr, ErrManualReviewRequired) &&
			!errors.Is(verifyErr, ErrInstitutionRejected) {
			return processed, verifyErr
		}
		processed++
	}
	return processed, nil
}

func (service Service) Expire(ctx context.Context, applicationID, commandID string) (domain.Application, error) {
	application, err := service.find(ctx, applicationID)
	if err != nil {
		return domain.Application{}, err
	}
	command := domain.Command{ID: strings.TrimSpace(commandID), ActorKey: "system", At: service.now().UTC()}
	updated, err := application.Expire(command, application.Version())
	if err != nil {
		return domain.Application{}, err
	}
	return service.update(ctx, application, updated, command.ID)
}

func (service Service) update(ctx context.Context, before, after domain.Application, commandID string) (domain.Application, error) {
	if err := service.repository.Update(ctx, after, before.Version(), commandID); err == nil {
		return after, nil
	}
	current, err := service.repository.FindByID(ctx, before.ID())
	if err == nil && current.SameCommand(after, commandID) {
		return current, nil
	}
	return domain.Application{}, ErrDependencyUnavailable
}

func (service Service) find(ctx context.Context, id string) (domain.Application, error) {
	if !service.ready() {
		return domain.Application{}, ErrDependencyUnavailable
	}
	application, err := service.repository.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Application{}, ErrNotFound
		}
		return domain.Application{}, ErrDependencyUnavailable
	}
	return application, nil
}

func (service Service) ready() bool {
	return service.repository != nil && service.provider != nil && service.reviews != nil &&
		service.keyer != nil && service.ids != nil && service.now != nil
}

func resultFor(application domain.Application, replayed bool) (Result, error) {
	result := Result{Application: application, Replayed: replayed}
	switch application.Status() {
	case domain.StatusApproved:
		return result, nil
	case domain.StatusRejected:
		return result, ErrInstitutionRejected
	case domain.StatusQueuedManual:
		return result, ErrManualReviewRequired
	default:
		return result, ErrDependencyUnavailable
	}
}
