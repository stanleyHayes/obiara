package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/launch/readiness/domain"
)

var (
	ErrInvalid     = errors.New("invalid launch readiness request")
	ErrUnavailable = errors.New("launch readiness unavailable")
	ErrNotFound    = errors.New("launch readiness snapshot not found")
	ErrApplied     = errors.New("launch readiness command applied")
	ErrConflict    = errors.New("launch readiness conflict")
)

type Service struct {
	authority Authority
	families  FamilyProjection
	hosts     HostProjection
	licenses  LicenseProjection
	repo      Repository
	ids       IDSource
	clock     Clock
}

func New(a Authority, f FamilyProjection, h HostProjection, l LicenseProjection, r Repository, i IDSource, c Clock) Service {
	return Service{a, f, h, l, r, i, c}
}
func (s Service) Project(ctx context.Context, actor, market, jurisdiction, command string) (domain.Snapshot, error) {
	if s.authority.RequireLaunchReviewer(ctx, actor) != nil {
		return domain.Snapshot{}, ErrUnavailable
	}
	f, e := s.families.CurrentFamilyDensity(ctx, market)
	if e != nil {
		return domain.Snapshot{}, ErrUnavailable
	}
	h, e := s.hosts.CurrentHostCoverage(ctx, market)
	if e != nil {
		return domain.Snapshot{}, ErrUnavailable
	}
	l, e := s.licenses.CurrentLicenseCoverage(ctx, market, jurisdiction)
	if e != nil {
		return domain.Snapshot{}, ErrUnavailable
	}
	now := s.clock.Now()
	snapshot, e := domain.Project(s.ids.NewID(), actor, market, jurisdiction, command, f, h, l, now)
	if e != nil {
		return domain.Snapshot{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, snapshot); e != nil {
		return domain.Snapshot{}, e
	}
	return snapshot, nil
}
