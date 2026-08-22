package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
	"time"
)

type Service struct {
	repository Repository
	keys       Keyer
	now        func() time.Time
}

func NewService(r Repository, k Keyer, n func() time.Time) Service { return Service{r, k, n} }
func (s Service) Set(ctx context.Context, roomID, memberID, commandID string, grant bool, expected uint64) (domain.Ribbon, error) {
	room, e := s.keys.Key("honesty_room", roomID)
	if e != nil {
		return domain.Ribbon{}, e
	}
	actor, e := s.keys.Key("honesty_actor", memberID)
	if e != nil {
		return domain.Ribbon{}, e
	}
	r, e := s.repository.Find(ctx, room)
	if e != nil {
		return domain.Ribbon{}, e
	}
	c := domain.Command{ID: commandID, ActorKey: actor, ExpectedRevision: expected, At: s.now()}
	var changed domain.Ribbon
	if grant {
		changed, e = r.Grant(c)
	} else {
		changed, e = r.Revoke(c)
	}
	if e != nil {
		return domain.Ribbon{}, e
	}
	if changed.Revision() == r.Revision() {
		return changed, nil
	}
	if e = s.repository.Save(ctx, changed, r.Revision(), commandID); e != nil {
		return domain.Ribbon{}, e
	}
	return changed, nil
}

// Open initialises this facet for a room.
//
// Each facet of a courtship room is its own aggregate with its own store, so
// opening a room has to open all of them. Without this the room existed only
// as a pace record and every other action failed on a missing document.
func (s Service) Open(ctx context.Context, roomID string, members []string, commandID string) (domain.Ribbon, error) {
	room, err := s.keys.Key("honesty_room", roomID)
	if err != nil {
		return domain.Ribbon{}, err
	}
	keyed := make([]string, len(members))
	for index, member := range members {
		keyed[index], err = s.keys.Key("honesty_actor", member)
		if err != nil {
			return domain.Ribbon{}, err
		}
	}
	aggregate, err := domain.New(room, keyed)
	if err != nil {
		return domain.Ribbon{}, err
	}
	if err := s.repository.Create(ctx, aggregate); err != nil {
		return domain.Ribbon{}, err
	}
	return aggregate, nil
}
