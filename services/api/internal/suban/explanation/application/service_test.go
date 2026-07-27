package application

import (
	"context"
	"fmt"
	suban "github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/suban/explanation/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestFileAppealRequiresSelfAndAdverseOwnedEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	a := NewMockAuthority(ctrl)
	events := NewMockEventSource(ctrl)
	repo := NewMockAppealRepository(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	now := time.Date(2026, 7, 27, 1, 0, 0, 0, time.UTC)
	a.EXPECT().RequireSelf(gomock.Any(), "actor", "subject").Return(nil)
	events.EXPECT().FindForSubject(gomock.Any(), "subject", "event").Return(suban.Event{ID: "event", SubjectID: "subject", Kind: suban.KindHarassmentFinding, OccurredAt: now}, nil)
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	appeal, e := New(a, events, repo, ids, clock).File(context.Background(), "actor", "subject", "event", domain.ReasonEventInaccurate, "file-1")
	if e != nil || appeal.State().Status != domain.StatusPending {
		t.Fatal(e)
	}
}
func TestResolveRequiresAuthorizedDistinctHuman(t *testing.T) {
	ctrl := gomock.NewController(t)
	a := NewMockAuthority(ctrl)
	repo := NewMockAppealRepository(ctrl)
	clock := NewMockClock(ctrl)
	appeal, _ := domain.File(key(1), "subject", "event", domain.ReasonWrongSubject, "file-1", time.Now())
	a.EXPECT().RequireAppealReviewer(gomock.Any(), key(2)).Return(nil)
	repo.EXPECT().Find(gomock.Any(), key(1)).Return(appeal, nil)
	clock.EXPECT().Now().Return(time.Now())
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1), "resolve-1").Return(nil)
	if _, e := New(a, NewMockEventSource(ctrl), repo, NewMockIDSource(ctrl), clock).Resolve(context.Background(), key(2), key(1), key(3), "resolve-1", domain.StatusUpheld); e != nil {
		t.Fatal(e)
	}
}
