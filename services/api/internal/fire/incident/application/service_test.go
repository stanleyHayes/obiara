package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/incident/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestSafetyActionPrecedesMinimalRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, p, safety, router, k := NewMockRepository(ctrl), NewMockParticipantAuthority(ctrl), NewMockSafetyAction(ctrl), NewMockTrustSafetyRouter(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	svc := NewService(r, p, safety, router, k, fixedID("case-1"), func() time.Time { return now })
	p.EXPECT().RequireParticipant(gomock.Any(), "fire-private", "actor-private")
	safety.EXPECT().Apply(gomock.Any(), "trigger", "fire-private", "actor-private", "leave")
	k.EXPECT().Key("fire-incident:fire", "fire-private").Return(ak(1), nil)
	k.EXPECT().Key("fire-incident:actor", "actor-private").Return(ak(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any())
	router.EXPECT().Revalidate(gomock.Any())
	router.EXPECT().Route(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c domain.Case) error {
		if c.FireKey != ak(1) || c.ActorKey != ak(2) {
			t.Fatal("raw route")
		}
		return nil
	})
	r.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "trigger:route")
	projection, e := svc.Trigger(context.Background(), Command{ID: "trigger", FireID: "fire-private", ActorID: "actor-private", Category: domain.CategoryThreat, Action: ActionLeave})
	if e != nil || projection.Reference != "case-1" || !projection.Routed {
		t.Fatalf("%+v %v", projection, e)
	}
}
