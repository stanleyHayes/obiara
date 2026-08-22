package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/closure/domain"
	"time"
)

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
	threshold  time.Duration
}

func NewService(r Repository, k Keyer, now func() time.Time, threshold time.Duration) Service {
	return Service{r, k, now, threshold}
}
func (s Service) Close(ctx context.Context, roomID, memberID, commandID string, expected uint64) (domain.Closure, error) {
	return s.apply(ctx, roomID, memberID, commandID, expected, false)
}
func (s Service) CloseInactive(ctx context.Context, roomID, commandID string, expected uint64) (domain.Closure, error) {
	return s.apply(ctx, roomID, "", commandID, expected, true)
}
func (s Service) apply(ctx context.Context, roomID, memberID, commandID string, expected uint64, inactive bool) (domain.Closure, error) {
	roomKey, err := s.keys.Key("closure_room", roomID)
	if err != nil {
		return domain.Closure{}, err
	}
	actorKey := ""
	if !inactive {
		actorKey, err = s.keys.Key("closure_actor", memberID)
		if err != nil {
			return domain.Closure{}, err
		}
	}
	current, err := s.repository.Find(ctx, roomKey)
	if err != nil {
		return domain.Closure{}, err
	}
	cmd := domain.Command{ID: commandID, ActorKey: actorKey, ExpectedRevision: expected, At: s.now()}
	var changed domain.Closure
	if inactive {
		changed, err = current.CloseInactive(cmd, s.threshold)
	} else {
		changed, err = current.Close(cmd)
	}
	if err != nil {
		return domain.Closure{}, err
	}
	if changed.Revision() != current.Revision() {
		if err = s.repository.Save(ctx, changed, current.Revision(), commandID); err != nil {
			return domain.Closure{}, err
		}
	}
	return changed, nil
}

// Open initialises this facet for a room.
//
// Each facet of a courtship room is its own aggregate with its own store, so
// opening a room has to open all of them. Without this the room existed only
// as a pace record and every other action failed on a missing document.
func (s Service) Open(ctx context.Context, roomID string, members []string, commandID string) (domain.Closure, error) {
	room, err := s.keys.Key("closure_room", roomID)
	if err != nil {
		return domain.Closure{}, err
	}
	keyed := make([]string, len(members))
	for index, member := range members {
		keyed[index], err = s.keys.Key("closure_actor", member)
		if err != nil {
			return domain.Closure{}, err
		}
	}
	aggregate, err := domain.New(room, keyed, s.now())
	if err != nil {
		return domain.Closure{}, err
	}
	if err := s.repository.Create(ctx, aggregate); err != nil {
		return domain.Closure{}, err
	}
	return aggregate, nil
}
