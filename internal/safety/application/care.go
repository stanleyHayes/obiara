package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// CareRepository persists care cases.
type CareRepository interface {
	Create(context.Context, domain.CareCase) error
	FindByID(context.Context, string) (domain.CareCase, error)
	Update(context.Context, domain.CareCase) error
	NextOpen(context.Context, int) ([]domain.CareCase, error)
}

// QuieteningStore records notification quietening windows.
type QuieteningStore interface {
	Set(ctx context.Context, subjectID string, until time.Time) error
}

// CareService runs the care queue (E12-S05). It touches no enforcement
// path: seeking help is never punished (Doc 09 §5).
type CareService struct {
	cases      CareRepository
	quietening QuieteningStore
	now        func() time.Time
	newID      func() string
}

func NewCareService(cases CareRepository, quietening QuieteningStore, now func() time.Time, newID func() string) CareService {
	return CareService{cases: cases, quietening: quietening, now: now, newID: newID}
}

// Flag routes a distress or closure signal into the care queue
// immediately. Closure signals also start the 72-hour quietening window.
func (service CareService) Flag(ctx context.Context, subjectID string, signal domain.Signal) (domain.CareCase, error) {
	careCase, err := domain.NewCareCase(service.newID(), subjectID, signal, service.now())
	if err != nil {
		return domain.CareCase{}, err
	}
	if err := service.cases.Create(ctx, careCase); err != nil {
		return domain.CareCase{}, err
	}
	if careCase.NeedsQuietening() {
		if err := service.quietening.Set(ctx, subjectID, service.now().Add(domain.QuieteningWindow)); err != nil {
			return domain.CareCase{}, err
		}
	}
	return careCase, nil
}

// Engage marks a trained human on the case.
func (service CareService) Engage(ctx context.Context, caseID string) (domain.CareCase, error) {
	careCase, err := service.cases.FindByID(ctx, caseID)
	if err != nil {
		return domain.CareCase{}, err
	}
	if err := careCase.Engage(); err != nil {
		return domain.CareCase{}, err
	}
	if err := service.cases.Update(ctx, careCase); err != nil {
		return domain.CareCase{}, err
	}
	return careCase, nil
}

// Resolve closes the case with the resource-first scripts used.
func (service CareService) Resolve(ctx context.Context, caseID string, scripts []domain.ScriptKey) (domain.CareCase, error) {
	careCase, err := service.cases.FindByID(ctx, caseID)
	if err != nil {
		return domain.CareCase{}, err
	}
	if err := careCase.Resolve(scripts, service.now()); err != nil {
		return domain.CareCase{}, err
	}
	if err := service.cases.Update(ctx, careCase); err != nil {
		return domain.CareCase{}, err
	}
	return careCase, nil
}

// NextOpen lists open and engaged cases oldest-first for the care rota
// (Doc 09 §3: care-flag immediate routing to trained staff).
func (service CareService) NextOpen(ctx context.Context, limit int) ([]domain.CareCase, error) {
	return service.cases.NextOpen(ctx, limit)
}
