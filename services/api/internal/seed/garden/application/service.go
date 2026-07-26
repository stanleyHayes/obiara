package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
)

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
}

func NewService(repository Repository, keys Keyer, now func() time.Time) Service {
	return Service{repository: repository, keys: keys, now: now}
}

func (service Service) Summary(ctx context.Context, memberID string) (domain.Summary, error) {
	ownerKey, err := service.keys.Key("garden_owner", memberID)
	if err != nil {
		return domain.Summary{}, err
	}
	at := service.now().UTC()
	if _, err = service.repository.ExpireDue(ctx, ownerKey, at, 100); err != nil {
		return domain.Summary{}, err
	}
	items, err := service.repository.ListOwner(ctx, ownerKey)
	if err != nil {
		return domain.Summary{}, err
	}
	return domain.Summarize(items, at)
}

func (service Service) Project(ctx context.Context, memberID, seedID string, next domain.State, expected uint64) (domain.Item, error) {
	ownerKey, err := service.keys.Key("garden_owner", memberID)
	if err != nil {
		return domain.Item{}, err
	}
	seedKey, err := service.keys.Key("garden_seed", seedID)
	if err != nil {
		return domain.Item{}, err
	}
	item, err := service.repository.Find(ctx, ownerKey, seedKey)
	if err != nil {
		return domain.Item{}, err
	}
	changed, err := item.Transition(next, expected, service.now())
	if err != nil {
		return domain.Item{}, err
	}
	if err = service.repository.Save(ctx, changed, expected); err != nil {
		return domain.Item{}, err
	}
	return changed, nil
}

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
