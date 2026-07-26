package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/harvest/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestCreateRevalidatesAllBoundaries(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, pair, owner, recipes, provider, keyer := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockPairPolicy(ctrl), NewMockOwnership(ctrl), NewMockRecipeValidator(ctrl), NewMockProviderAuth(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	s := NewService(r, a, pair, owner, recipes, provider, keyer, fixedID("harvest-1"), fixedID("handoff-1"), func() time.Time { return now })
	a.EXPECT().Require(gomock.Any(), "a", "cloth.harvest.create", "")
	pair.EXPECT().Revalidate(gomock.Any(), "a", "b")
	owner.EXPECT().Revalidate(gomock.Any(), "a", "b", "recipe-private")
	keyer.EXPECT().Key("cloth-harvest:recipe", "recipe-private").Return(ak(3), nil)
	keyer.EXPECT().Key("cloth-harvest:delivery", "delivery-private").Return(ak(4), nil)
	recipes.EXPECT().Revalidate(gomock.Any(), gomock.Any())
	keyer.EXPECT().Key("cloth-harvest:member", "a").Return(ak(1), nil)
	keyer.EXPECT().Key("cloth-harvest:member", "b").Return(ak(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, h domain.Harvest) error {
		if h.Payload().DeliveryRef != ak(4) || len(h.Payload().ProductionTokens) != 6 {
			t.Fatal("unsafe payload")
		}
		return nil
	})
	_, e := s.Create(context.Background(), Command{ID: "create", ActorID: "a", FirstMemberID: "a", SecondMemberID: "b"}, Draft{RecipeRef: "recipe-private", RecipeVersion: "v1", RenderSeed: ak(5), ProductionTokens: []string{"warp_even", "weft_close", "edge_soft", "tone_warm", "mark_sparse", "finish_matte"}, Format: "woven_band", DeliveryRef: "delivery-private", PolicyVersion: "p1"})
	if e != nil {
		t.Fatal(e)
	}
}
