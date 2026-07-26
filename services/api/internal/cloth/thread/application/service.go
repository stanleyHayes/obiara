package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/thread/domain"
	"time"
)

type Service struct {
	repository Repository
	keys       Keyer
	evidence   RevealEvidence
	now        func() time.Time
}

func NewService(r Repository, k Keyer, e RevealEvidence, n func() time.Time) Service {
	return Service{r, k, e, n}
}
func (s Service) Issue(ctx context.Context, pairID, memberID, commandID, revealRef, recipeRef string, bandVersion uint32, expected uint64) (domain.View, error) {
	pairKey, err := s.keys.Key("thread_pair", pairID)
	if err != nil {
		return domain.View{}, err
	}
	actorKey, err := s.keys.Key("thread_member", memberID)
	if err != nil {
		return domain.View{}, err
	}
	current, err := s.repository.Find(ctx, pairKey)
	if err != nil {
		return domain.View{}, err
	}
	ok, err := s.evidence.ThemeOneRevealed(ctx, pairKey, revealRef)
	if err != nil {
		return domain.View{}, err
	}
	if !ok {
		return domain.View{}, domain.ErrDenied
	}
	changed, err := current.Issue(domain.Command{ID: commandID, ActorKey: actorKey, RevealRef: revealRef, RecipeRef: recipeRef, BandVersion: bandVersion, ExpectedRevision: expected, At: s.now()})
	if err != nil {
		return domain.View{}, err
	}
	if changed.Revision() != current.Revision() {
		if err = s.repository.Save(ctx, changed, current.Revision(), commandID); err != nil {
			return domain.View{}, err
		}
	}
	return changed.View(actorKey)
}
func (s Service) View(ctx context.Context, pairID, memberID string) (domain.View, error) {
	pairKey, err := s.keys.Key("thread_pair", pairID)
	if err != nil {
		return domain.View{}, err
	}
	actorKey, err := s.keys.Key("thread_member", memberID)
	if err != nil {
		return domain.View{}, err
	}
	t, err := s.repository.Find(ctx, pairKey)
	if err != nil {
		return domain.View{}, err
	}
	return t.View(actorKey)
}
