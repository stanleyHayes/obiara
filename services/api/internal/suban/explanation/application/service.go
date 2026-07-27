package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/domain"
)

var (
	ErrInvalid     = errors.New("invalid Suban explanation request")
	ErrUnavailable = errors.New("Suban explanation unavailable")
	ErrNotFound    = errors.New("appeal not found")
	ErrConflict    = errors.New("appeal conflict")
	ErrApplied     = errors.New("appeal command applied")
)

type Service struct {
	authority Authority
	events    EventSource
	appeals   AppealRepository
	ids       IDSource
	clock     Clock
}

func New(a Authority, e EventSource, r AppealRepository, ids IDSource, c Clock) Service {
	return Service{a, e, r, ids, c}
}
func (s Service) Explain(ctx context.Context, actor, subject string) (domain.Explanation, error) {
	if s.authority.RequireSelf(ctx, actor, subject) != nil {
		return domain.Explanation{}, ErrUnavailable
	}
	events, e := s.events.ListForSubject(ctx, subject)
	if e != nil {
		return domain.Explanation{}, ErrUnavailable
	}
	result, e := domain.Explain(subject, events, s.clock.Now())
	if e != nil {
		return domain.Explanation{}, ErrInvalid
	}
	return result, nil
}
func (s Service) File(ctx context.Context, actor, subject, eventID string, reason domain.Reason, command string) (domain.Appeal, error) {
	if s.authority.RequireSelf(ctx, actor, subject) != nil {
		return domain.Appeal{}, ErrUnavailable
	}
	event, e := s.events.FindForSubject(ctx, subject, eventID)
	if e != nil || event.SubjectID != subject || !domain.Appealable(event.Kind) {
		return domain.Appeal{}, ErrInvalid
	}
	appeal, e := domain.File(s.ids.NewID(), subject, eventID, reason, command, s.clock.Now())
	if e != nil {
		return domain.Appeal{}, ErrInvalid
	}
	if e = s.appeals.Create(ctx, appeal); e != nil {
		return domain.Appeal{}, e
	}
	return appeal, nil
}
func (s Service) Resolve(ctx context.Context, reviewer, appealID, reasoningRef, command string, status domain.Status) (domain.Appeal, error) {
	if s.authority.RequireAppealReviewer(ctx, reviewer) != nil {
		return domain.Appeal{}, ErrUnavailable
	}
	appeal, e := s.appeals.Find(ctx, appealID)
	if e != nil {
		return domain.Appeal{}, e
	}
	next, e := appeal.Resolve(status, reviewer, reasoningRef, command, s.clock.Now())
	if e != nil {
		return domain.Appeal{}, ErrInvalid
	}
	if e = s.appeals.Save(ctx, next, appeal.Revision(), command); e != nil {
		return domain.Appeal{}, e
	}
	return next, nil
}
