package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pause/domain"
	"time"
)

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
}

func NewService(r Repository, k Keyer, n func() time.Time) Service { return Service{r, k, n} }
func (s Service) Apply(ctx context.Context, roomID, memberID, commandID string, action domain.Action, expected uint64) (domain.Stone, error) {
	roomKey, e := s.keys.Key("pause_room", roomID)
	if e != nil {
		return domain.Stone{}, e
	}
	actorKey, e := s.keys.Key("pause_actor", memberID)
	if e != nil {
		return domain.Stone{}, e
	}
	stone, e := s.repository.Find(ctx, roomKey)
	if e != nil {
		return domain.Stone{}, e
	}
	command := domain.Command{ID: commandID, ActorKey: actorKey, ExpectedRevision: expected, At: s.now()}
	var changed domain.Stone
	switch action {
	case domain.ActionPause:
		changed, e = stone.Pause(command)
	case domain.ActionAck:
		changed, e = stone.Acknowledge(command)
	case domain.ActionResume:
		changed, e = stone.Resume(command)
	default:
		return domain.Stone{}, domain.ErrInvalid
	}
	if e != nil {
		return domain.Stone{}, e
	}
	if changed.Revision() == stone.Revision() {
		return changed, nil
	}
	if e = s.repository.Save(ctx, changed, stone.Revision(), commandID); e != nil {
		return domain.Stone{}, e
	}
	return changed, nil
}
