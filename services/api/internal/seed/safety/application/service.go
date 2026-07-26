package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/safety/domain"
)

var ErrNotFound = errors.New("seed safety bucket not found")

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
}

func NewService(repository Repository, keys Keyer, now func() time.Time) Service {
	return Service{repository: repository, keys: keys, now: now}
}

func (service Service) Check(ctx context.Context, memberID string, operation domain.Operation) (domain.Decision, error) {
	actorKey, err := service.keys.Key("seed_safety_actor", memberID)
	if err != nil {
		return domain.Decision{}, err
	}
	at := service.now().UTC()
	for attempt := 0; attempt < 3; attempt++ {
		bucket, findErr := service.repository.Find(ctx, actorKey)
		if errors.Is(findErr, ErrNotFound) {
			bucket, err = domain.New(actorKey, at)
			if err != nil {
				return domain.Decision{}, err
			}
			if err = service.repository.Create(ctx, bucket); err != nil {
				continue
			}
		} else if findErr != nil {
			return domain.Decision{}, findErr
		}
		changed, decision, changeErr := bucket.Evaluate(operation, bucket.Revision, at)
		if changeErr != nil {
			return domain.Decision{}, changeErr
		}
		if err = service.repository.Save(ctx, changed, bucket.Revision); errors.Is(err, domain.ErrStaleRevision) {
			continue
		}
		if err != nil {
			return domain.Decision{}, err
		}
		if decision.CareSignal {
			err = service.repository.AppendCareSignal(ctx, CareSignal{
				ActorKey: actorKey, Code: "repeated_seed_demand", WindowRevision: changed.Revision,
			})
			if err != nil {
				return domain.Decision{}, err
			}
		}
		if !decision.Allowed {
			return decision, domain.ErrThrottled
		}
		return decision, nil
	}
	return domain.Decision{}, domain.ErrStaleRevision
}
