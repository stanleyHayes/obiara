package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
)

type ExceptionView struct {
	FactID, ProviderKey, ReferenceKey string
	Currency                          domain.Currency
	Minor                             int64
	Outcome                           domain.Outcome
	Exception                         domain.ExceptionCode
	OccurredAt, RecordedAt            time.Time
}

type Overview struct {
	Exceptions  []ExceptionView
	Checkpoints []domain.Checkpoint
}

type QueryService struct{ repo Repository }

func NewQueryService(repo Repository) QueryService { return QueryService{repo: repo} }

func (service QueryService) Overview(ctx context.Context, limit int) (Overview, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	audits, err := service.repo.ListRecentAudits(ctx, limit)
	if err != nil {
		return Overview{}, err
	}
	exceptions := make([]ExceptionView, 0, len(audits))
	for _, audit := range audits {
		if audit.Outcome() != domain.OutcomeException {
			continue
		}
		fact, findErr := service.repo.FindFactByID(ctx, audit.FactID())
		if findErr != nil {
			return Overview{}, findErr
		}
		exceptions = append(exceptions, ExceptionView{
			FactID: fact.ID(), ProviderKey: fact.ProviderKey(), ReferenceKey: fact.ReferenceKey(),
			Currency: fact.Currency(), Minor: fact.Minor(), Outcome: audit.Outcome(),
			Exception: audit.Exception(), OccurredAt: fact.OccurredAt(), RecordedAt: audit.RecordedAt(),
		})
	}
	checkpoints, err := service.repo.ListRecentCheckpoints(ctx, 14)
	if err != nil {
		return Overview{}, err
	}
	return Overview{Exceptions: exceptions, Checkpoints: checkpoints}, nil
}
