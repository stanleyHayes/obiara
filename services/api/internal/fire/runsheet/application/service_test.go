package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/fire/runsheet/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type fixedID string

func (i fixedID) NewID() string { return string(i) }
func ak(n int) string           { return fmt.Sprintf("%064x", n) }
func TestCreateRequiresAuthorityAndStoresOpaqueKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockKeyer(ctrl)
	s := NewService(r, a, k, fixedID("sheet-1"), func() time.Time { return time.Now() })
	a.EXPECT().RequireHostOrCohost(gomock.Any(), "fire-private", "host-private")
	k.EXPECT().Key("fire-runsheet:fire", "fire-private").Return(ak(1), nil)
	k.EXPECT().Key("fire-runsheet:host", "host-private").Return(ak(2), nil)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, v domain.RunSheet) error {
		if v.FireKey() != ak(1) || v.HostKey() != ak(2) {
			t.Fatal("raw keys")
		}
		return nil
	})
	_, e := s.Create(context.Background(), Command{ID: "create", FireID: "fire-private", ActorID: "host-private"}, 1, []domain.Segment{{Type: domain.TypeTalk, TitleCode: "welcome", PlannedDuration: time.Minute}})
	if e != nil {
		t.Fatal(e)
	}
}

func TestOutsiderCannotStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	r, a, k := NewMockRepository(ctrl), NewMockAuthority(ctrl), NewMockKeyer(ctrl)
	s := NewService(r, a, k, fixedID("sheet-1"), time.Now)
	a.EXPECT().RequireHostOrCohost(gomock.Any(), "fire-private", "outsider").Return(errors.New("denied"))
	if _, err := s.Start(context.Background(), Command{ID: "start", RunSheetID: "sheet-1", FireID: "fire-private", ActorID: "outsider", ExpectedRevision: 1}); !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("outsider start=%v", err)
	}
}
