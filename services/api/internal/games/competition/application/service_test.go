package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/games/competition/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestCreateRequiresEveryOptIn(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, opt, results, fair, k := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockOptIn(ctrl), NewMockResultVerifier(ctrl), NewMockFairPlayVerifier(ctrl), NewMockKeyer(ctrl)
	s := NewService(r, a, opt, results, fair, k, fixedID("competition-1"), fixedID("review-1"), time.Now)
	ids := []string{"a", "b", "c", "d"}
	a.EXPECT().RevalidateCohort(gomock.Any(), "cohort-private", ids)
	for i, id := range ids {
		opt.EXPECT().Revalidate(gomock.Any(), "cohort-private", id)
		k.EXPECT().Key("game-competition:entrant", id).Return(ak(i+1), nil)
	}
	k.EXPECT().Key("game-competition:cohort", "cohort-private").Return(ak(9), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, x domain.Competition) error {
		if len(x.Entrants()) != 4 {
			t.Fatal("cohort")
		}
		return nil
	})
	p, e := s.Create(context.Background(), Command{ID: "create", CohortID: "cohort-private"}, ids)
	if e != nil || len(p.Matches) != 2 {
		t.Fatal(e)
	}
}
