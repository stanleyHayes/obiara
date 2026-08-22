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

// Open initialises this facet for a room.
//
// Each facet of a courtship room is its own aggregate with its own store, so
// opening a room has to open all of them. Without this the room existed only
// as a pace record and every other action failed on a missing document.
func (s Service) Open(ctx context.Context, roomID string, members []string, commandID string) (domain.Stone, error) {
	room, err := s.keys.Key("pause_room", roomID)
	if err != nil {
		return domain.Stone{}, err
	}
	keyed := make([]string, len(members))
	for index, member := range members {
		keyed[index], err = s.keys.Key("pause_actor", member)
		if err != nil {
			return domain.Stone{}, err
		}
	}
	aggregate, err := domain.New(room, keyed)
	if err != nil {
		return domain.Stone{}, err
	}
	if err := s.repository.Create(ctx, aggregate); err != nil {
		return domain.Stone{}, err
	}
	return aggregate, nil
}
