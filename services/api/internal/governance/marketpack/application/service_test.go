package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/governance/marketpack/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestProposeRequiresAuthorityAndCurrentMaster(t *testing.T) {
	ctrl := gomock.NewController(t)
	masters := NewMockMasterCatalog(ctrl)
	authority := NewMockAuthority(ctrl)
	repo := NewMockRepository(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	now := time.Date(2026, 7, 27, 3, 0, 0, 0, time.UTC)
	master, _ := domain.NewMaster(domain.MasterSpec{ID: "ghana.master", Version: 1, Entries: []domain.MasterEntry{{Key: "hello.world", Text: "Hello {name}."}}})
	authority.EXPECT().RequireAuthor(gomock.Any(), key(2)).Return(nil)
	masters.EXPECT().Current(gomock.Any()).Return(master, nil)
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	p, e := New(masters, authority, repo, ids, clock).Propose(context.Background(), key(2), "GH", "tw-GH", 1, []domain.Translation{{Key: "hello.world", Text: "Maakye {name}."}}, "propose-1")
	if e != nil || p.State().MasterVersion != 1 {
		t.Fatal(e)
	}
}
