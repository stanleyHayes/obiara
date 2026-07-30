package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

type Service struct {
	profiles Repository
	consent  ConsentEvaluator
	now      func() time.Time
}

func NewService(profiles Repository, consent ConsentEvaluator, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{profiles: profiles, consent: consent, now: now}
}

type FieldInput struct {
	Value      string
	Visibility domain.Visibility
	ConsentRef string
}

type UpsertCommand struct {
	CommandID        string
	MemberID         string
	ExpectedRevision uint64
	DisplayName      FieldInput
	Introduction     FieldInput
}

type UpsertResult struct {
	Profile  domain.Profile
	Replayed bool
}

func (service Service) Upsert(ctx context.Context, command UpsertCommand) (UpsertResult, error) {
	if service.profiles == nil || service.now == nil {
		return UpsertResult{}, ErrRepositoryUnavailable
	}
	displayName, err := domain.NewField(command.DisplayName.Value, command.DisplayName.Visibility, command.DisplayName.ConsentRef, 80, true)
	if err != nil {
		return UpsertResult{}, err
	}
	introduction, err := domain.NewField(command.Introduction.Value, command.Introduction.Visibility, command.Introduction.ConsentRef, 280, false)
	if err != nil {
		return UpsertResult{}, err
	}
	change := domain.Change{
		CommandID: strings.TrimSpace(command.CommandID), ExpectedRevision: command.ExpectedRevision,
		DisplayName: displayName, Introduction: introduction, RecordedAt: service.now().UTC(),
	}
	current, findErr := service.profiles.Find(ctx, strings.TrimSpace(command.MemberID))
	if errors.Is(findErr, ErrNotFound) {
		current, err = domain.Create(command.MemberID, change)
	} else if findErr != nil {
		return UpsertResult{}, ErrRepositoryUnavailable
	} else {
		wasApplied := current.HasCommand(change.CommandID)
		current, err = current.Update(change)
		if err == nil && wasApplied {
			return UpsertResult{Profile: current, Replayed: true}, nil
		}
	}
	if err != nil {
		return UpsertResult{}, err
	}
	if err = service.profiles.Save(ctx, current, command.ExpectedRevision, change.CommandID); err == nil {
		return UpsertResult{Profile: current}, nil
	}
	if errors.Is(err, ErrCommandAlreadyApplied) {
		reloaded, reloadErr := service.profiles.Find(ctx, current.MemberID())
		if reloadErr == nil && reloaded.HasCommand(change.CommandID) {
			replay, replayErr := reloaded.Update(change)
			if replayErr == nil && replay.HasCommand(change.CommandID) {
				return UpsertResult{Profile: replay, Replayed: true}, nil
			}
			if errors.Is(replayErr, domain.ErrCommandMismatch) {
				return UpsertResult{}, replayErr
			}
		}
		if reloadErr == nil {
			return UpsertResult{}, domain.ErrCommandMismatch
		}
	}
	if errors.Is(err, ErrOptimisticConflict) {
		return UpsertResult{}, domain.ErrStaleRevision
	}
	return UpsertResult{}, ErrRepositoryUnavailable
}

type View struct {
	MemberID               string
	DisplayName            *string
	Introduction           *string
	DisplayNameVisibility  domain.Visibility
	IntroductionVisibility domain.Visibility
	Revision               uint64
	UpdatedAt              time.Time
}

func (service Service) View(ctx context.Context, memberID string, audience domain.Audience) (View, error) {
	if service.profiles == nil {
		return View{}, ErrRepositoryUnavailable
	}
	profile, err := service.profiles.Find(ctx, strings.TrimSpace(memberID))
	if err != nil {
		return View{}, err
	}
	view := View{
		MemberID: profile.MemberID(), Revision: profile.Revision(), UpdatedAt: profile.UpdatedAt(),
		DisplayNameVisibility:  profile.DisplayName().Visibility(),
		IntroductionVisibility: profile.Introduction().Visibility(),
	}
	if visible, err := service.fieldVisible(ctx, profile, profile.DisplayName(), audience); err != nil {
		return View{}, err
	} else if visible {
		value := profile.DisplayName().Value()
		view.DisplayName = &value
	}
	if visible, err := service.fieldVisible(ctx, profile, profile.Introduction(), audience); err != nil {
		return View{}, err
	} else if visible && profile.Introduction().Value() != "" {
		value := profile.Introduction().Value()
		view.Introduction = &value
	}
	return view, nil
}

func (service Service) fieldVisible(ctx context.Context, profile domain.Profile, field domain.Field, audience domain.Audience) (bool, error) {
	visible, err := field.VisibleTo(audience)
	if err != nil || !visible || audience != domain.AudienceCommunity {
		return visible, err
	}
	if service.consent == nil {
		return false, ErrConsentDenied
	}
	allowed, err := service.consent.Allows(ctx, profile.MemberID(), field.ConsentRef())
	if err != nil || !allowed {
		return false, ErrConsentDenied
	}
	return true, nil
}
