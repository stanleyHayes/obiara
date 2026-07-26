package application

import (
	"context"
	"time"
)

// EnforcementService applies trust-and-safety status effects to accounts
// (E12-S04): timed suspensions, blocks, and automatic reactivation of
// expired suspensions. Sessions are revoked separately (SessionService).
type EnforcementService struct {
	accounts AccountRepository
	now      func() time.Time
}

func NewEnforcementService(accounts AccountRepository, now func() time.Time) EnforcementService {
	return EnforcementService{accounts: accounts, now: now}
}

// Suspend places an account under a timed suspension.
func (service EnforcementService) Suspend(ctx context.Context, accountID string, until time.Time) error {
	account, err := service.accounts.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if err := account.Suspend(until); err != nil {
		return err
	}
	return service.accounts.Update(ctx, account)
}

// Block ends an account's product access.
func (service EnforcementService) Block(ctx context.Context, accountID string) error {
	account, err := service.accounts.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	account.Block()
	return service.accounts.Update(ctx, account)
}

// ReactivateExpired lifts one account's expired suspension.
func (service EnforcementService) ReactivateExpired(ctx context.Context, accountID string) error {
	account, err := service.accounts.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if err := account.Reactivate(service.now()); err != nil {
		return err
	}
	return service.accounts.Update(ctx, account)
}

// ReactivateAllExpired lifts every expired suspension and returns the
// count. The worker scheduler drives it; expiry uses server time only.
func (service EnforcementService) ReactivateAllExpired(ctx context.Context) (int, error) {
	expired, err := service.accounts.ListSuspendedExpired(ctx, service.now(), 500)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, account := range expired {
		if err := account.Reactivate(service.now()); err != nil {
			continue
		}
		if err := service.accounts.Update(ctx, account); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
