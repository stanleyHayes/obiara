package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// TierService applies account tier transitions with audit (E03-S06). Every
// transition names its actor and reason; promotion is one step at a time
// and demotion always requires a reason (verification reversal, safety).
type TierService struct {
	accounts AccountRepository
	now      func() time.Time
}

func NewTierService(accounts AccountRepository, now func() time.Time) TierService {
	return TierService{accounts: accounts, now: now}
}

// Transition moves the account to the target tier and returns the recorded
// audit transition.
func (service TierService) Transition(ctx context.Context, accountID string, target domain.Tier, reason, actorID string) (domain.TierTransition, error) {
	account, err := service.accounts.FindByID(ctx, accountID)
	if err != nil {
		return domain.TierTransition{}, err
	}
	transition, err := account.ApplyTransition(target, reason, actorID, service.now())
	if err != nil {
		return domain.TierTransition{}, err
	}
	if err := service.accounts.UpdateWithAudit(ctx, account, transition); err != nil {
		return domain.TierTransition{}, err
	}
	return transition, nil
}
