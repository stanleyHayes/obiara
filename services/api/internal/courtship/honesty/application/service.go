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
