package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/games/conduct/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestRecordRevalidatesAuthorityAndProvenance(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, e, k := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockEventVerifier(ctrl), NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	s := NewService(r, a, e, k, fixedID("signal-1"), func() time.Time { return now })
	a.EXPECT().RequireSubject(gomock.Any(), "game-private", "actor-private", "subject-private")
	e.EXPECT().Revalidate(gomock.Any(), "game-private", "event-private", "subject-private", domain.EventAbandoned)
	k.EXPECT().Key("game-conduct:game", "game-private").Return(ak(1), nil)
	k.EXPECT().Key("game-conduct:subject", "subject-private").Return(ak(2), nil)
	k.EXPECT().Key("game-conduct:event", "event-private").Return(ak(3), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, v domain.Signal) error {
		if v.Reason() != domain.ReasonAbandonment || v.Provenance() != domain.ProvenanceServerEvent {
			t.Fatal("mapping")
		}
		return nil
	})
	p, err := s.Record(context.Background(), Command{ID: "record", GameID: "game-private", ActorID: "actor-private", SubjectID: "subject-private", EventRef: "event-private"}, domain.EventAbandoned)
	if err != nil || p.Reference != "signal-1" {
		t.Fatal(err)
	}
}
