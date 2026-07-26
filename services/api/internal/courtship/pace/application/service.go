package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
)

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
}

func NewService(repository Repository, keys Keyer, now func() time.Time) Service {
	return Service{repository: repository, keys: keys, now: now}
}

func (service Service) Start(ctx context.Context, roomID string, members []string, commandID, actorID string) (domain.Pace, error) {
	room, err := service.keys.Key("pace_room", roomID)
	if err != nil {
		return domain.Pace{}, err
	}
	keyedMembers := make([]string, len(members))
	for index, member := range members {
		keyedMembers[index], err = service.keys.Key("pace_member", member)
		if err != nil {
			return domain.Pace{}, err
		}
	}
	actor, err := service.keys.Key("pace_member", actorID)
	if err != nil {
		return domain.Pace{}, err
	}
	pace, err := domain.New(room, keyedMembers, commandID, actor, service.now())
	if err == nil {
		err = service.repository.Create(ctx, pace)
	}
	return pace, err
}

func (service Service) Advance(ctx context.Context, roomID, commandID string, expected uint64) (domain.Pace, error) {
	return service.change(ctx, roomID, commandID, "", expected, false)
}

func (service Service) Relight(ctx context.Context, roomID, memberID, commandID string, expected uint64) (domain.Pace, error) {
	return service.change(ctx, roomID, commandID, memberID, expected, true)
}

func (service Service) change(ctx context.Context, roomID, commandID, memberID string, expected uint64, relight bool) (domain.Pace, error) {
	room, err := service.keys.Key("pace_room", roomID)
	if err != nil {
		return domain.Pace{}, err
	}
	pace, err := service.repository.Find(ctx, room)
	if err != nil {
		return domain.Pace{}, err
	}
	previousRevision := pace.Revision()
	command := domain.Command{ID: commandID, ExpectedRevision: expected, At: service.now()}
	if relight {
		command.ActorKey, err = service.keys.Key("pace_member", memberID)
		if err != nil {
			return domain.Pace{}, err
		}
		pace, err = pace.Relight(command)
	} else {
		pace, err = pace.Advance(command)
	}
	if err != nil {
		return domain.Pace{}, err
	}
	if pace.Revision() == previousRevision {
		return pace, nil
	}
	if err = service.repository.Save(ctx, pace, previousRevision, commandID); err != nil {
		return domain.Pace{}, err
	}
	return pace, nil
}
