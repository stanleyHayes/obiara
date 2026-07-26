package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/collision/domain"
)

var (
	ErrSignalRequired  = errors.New("collision signal is required")
	ErrSubjectRequired = errors.New("subject id is required")
	ErrCaseNotFound    = errors.New("collision review case not found")
	ErrStaleCase       = errors.New("collision review case was updated")
)

type Assessment struct {
	Kind      domain.Kind
	Signal    string
	SubjectID string
}

type Decision struct {
	Allowed        bool
	Collision      bool
	ReviewRequired bool
	Case           domain.Case
}

type Service struct {
	repository Repository
	keyer      Keyer
	now        func() time.Time
}

func NewService(repository Repository, keyer Keyer, now func() time.Time) Service {
	return Service{repository: repository, keyer: keyer, now: now}
}

func (service Service) Assess(ctx context.Context, assessment Assessment) (Decision, error) {
	if !domain.ValidKind(assessment.Kind) {
		return Decision{}, domain.ErrInvalidKind
	}
	if strings.TrimSpace(assessment.Signal) == "" {
		return Decision{}, ErrSignalRequired
	}
	if strings.TrimSpace(assessment.SubjectID) == "" {
		return Decision{}, ErrSubjectRequired
	}
	signal := strings.TrimSpace(assessment.Signal)
	if assessment.Kind == domain.KindKnownName {
		signal = strings.Join(strings.Fields(strings.ToLower(signal)), " ")
	}
	signalKey, err := service.keyer.Key("collision:"+string(assessment.Kind), signal)
	if err != nil {
		return Decision{}, err
	}
	subjectKey, err := service.keyer.Key("collision:subject", assessment.SubjectID)
	if err != nil {
		return Decision{}, err
	}
	collided, err := service.repository.RegisterSignal(ctx, assessment.Kind, signalKey, subjectKey)
	if err != nil {
		return Decision{}, err
	}
	if !collided {
		return Decision{Allowed: true}, nil
	}
	caseID, err := service.keyer.Key("collision:case", string(assessment.Kind)+":"+signalKey+":"+subjectKey)
	if err != nil {
		return Decision{}, err
	}
	reviewCase, audit, err := domain.NewCase(caseID, assessment.Kind, signalKey, subjectKey, service.now())
	if err != nil {
		return Decision{}, err
	}
	reviewCase, _, err = service.repository.Create(ctx, reviewCase, audit)
	if err != nil {
		return Decision{}, err
	}
	return Decision{
		Allowed: reviewCase.Allowed(), Collision: true,
		ReviewRequired: reviewCase.Status() == domain.StatusPending, Case: reviewCase,
	}, nil
}

func (service Service) Resolve(ctx context.Context, caseID string, resolution domain.Resolution, reasonCode, actorID string) (Decision, error) {
	reviewCase, err := service.repository.FindByID(ctx, caseID)
	if err != nil {
		return Decision{}, err
	}
	if strings.TrimSpace(actorID) == "" {
		return Decision{}, domain.ErrActorRequired
	}
	actorKey, err := service.keyer.Key("collision:actor", actorID)
	if err != nil {
		return Decision{}, err
	}
	previousVersion := reviewCase.Version()
	audit, err := reviewCase.Resolve(resolution, reasonCode, actorKey, service.now())
	if err != nil {
		return Decision{}, err
	}
	if err := service.repository.Resolve(ctx, reviewCase, audit, previousVersion); err != nil {
		return Decision{}, err
	}
	return Decision{Allowed: reviewCase.Allowed(), Collision: true, Case: reviewCase}, nil
}
