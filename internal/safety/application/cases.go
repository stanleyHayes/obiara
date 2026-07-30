package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var (
	ErrCaseNotFound = errors.New("safety case not found")
)

// CaseRepository persists T&S cases with a unique case per report
// (replay-safe queue building).
type CaseRepository interface {
	Create(context.Context, domain.Case) error
	FindByID(context.Context, string) (domain.Case, error)
	Update(context.Context, domain.Case) error
	// NextQueued returns open cases oldest-SLA-first for the desk.
	NextQueued(context.Context, domain.Queue, int) ([]domain.Case, error)
	// CountBreached reports open cases past their SLA deadline.
	CountBreached(context.Context, time.Time) (int, error)
}

// CaseService manages the tiered queues (E12-S02).
type CaseService struct {
	cases CaseRepository
	now   func() time.Time
	newID func() string
}

func NewCaseService(cases CaseRepository, now func() time.Time, newID func() string) CaseService {
	return CaseService{cases: cases, now: now, newID: newID}
}

// Open builds a case from a filed report with its SLA deadline. The unique
// reportId index makes redelivery a no-op (ErrReportAlreadyQueued).
func (service CaseService) Open(ctx context.Context, report domain.Report) (domain.Case, error) {
	safetyCase, err := domain.NewCaseFromReport(service.newID(), report, service.now())
	if err != nil {
		return domain.Case{}, err
	}
	if err := service.cases.Create(ctx, safetyCase); err != nil {
		return domain.Case{}, err
	}
	return safetyCase, nil
}

// Assign moves a queued case into review.
func (service CaseService) Assign(ctx context.Context, caseID, agentID string) (domain.Case, error) {
	safetyCase, err := service.cases.FindByID(ctx, caseID)
	if err != nil {
		return domain.Case{}, err
	}
	if err := safetyCase.Assign(agentID, service.now()); err != nil {
		return domain.Case{}, err
	}
	if err := service.cases.Update(ctx, safetyCase); err != nil {
		return domain.Case{}, err
	}
	return safetyCase, nil
}

// Resolve closes an in-review case with an outcome.
func (service CaseService) Resolve(ctx context.Context, caseID, outcome, agentID string) (domain.Case, error) {
	safetyCase, err := service.cases.FindByID(ctx, caseID)
	if err != nil {
		return domain.Case{}, err
	}
	if err := safetyCase.Resolve(outcome, agentID, service.now()); err != nil {
		return domain.Case{}, err
	}
	if err := service.cases.Update(ctx, safetyCase); err != nil {
		return domain.Case{}, err
	}
	return safetyCase, nil
}

// NextQueued lists open cases oldest-SLA-first for a desk.
func (service CaseService) NextQueued(ctx context.Context, queue domain.Queue, limit int) ([]domain.Case, error) {
	return service.cases.NextQueued(ctx, queue, limit)
}

func (service CaseService) Find(ctx context.Context, caseID string) (domain.Case, error) {
	return service.cases.FindByID(ctx, caseID)
}

// BreachCount is the SLA observability signal (Doc 09 §3 queues and SLAs).
func (service CaseService) BreachCount(ctx context.Context) (int, error) {
	return service.cases.CountBreached(ctx, service.now())
}
