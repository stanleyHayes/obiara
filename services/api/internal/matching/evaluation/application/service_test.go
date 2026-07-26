package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type keyer struct{}

func (keyer) Key(_ string, v string) (string, error) { return "key:" + v, nil }

type ids struct{}

func (ids) NewID() string { return "evaluation:1" }
func TestRecordRevalidatesConsentedSnapshotAndBoundedSlices(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	snap := NewMockSnapshotVerifier(ctrl)
	policy := NewMockSlicePolicy(ctrl)
	auth := NewMockAuthority(ctrl)
	at := time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)
	e, _ := domain.Create("evaluation:1", "candidate.compatibility", 1, domain.Command{ID: "create", At: at})
	s := domain.Snapshot{ID: "snapshot:1", Version: 1, ConsentVersion: 3, EvaluatedAt: at}
	m := domain.Metrics{Cohort: 300, Evidence: 2000, Quality: .75, ErrorRate: .2, MaxDisparity: .05, Slices: []domain.SliceMetric{{PolicyKey: "approved.region", Cohort: 100, Quality: .74, ErrorRate: .2}}}
	auth.EXPECT().RequireEvaluator(gomock.Any(), "actor").Return(nil)
	snap.EXPECT().Revalidate(gomock.Any(), s).Return(nil)
	policy.EXPECT().RequireApproved(gomock.Any(), []string{"approved.region"}).Return(nil)
	repo.EXPECT().Find(gomock.Any(), "evaluation:1").Return(e, nil)
	repo.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "evaluate").Return(nil)
	service := NewService(repo, snap, policy, auth, keyer{}, ids{}, func() time.Time { return at })
	if _, err := service.Record(context.Background(), RecordCommand{Actor: "actor", ID: "evaluation:1", CommandID: "evaluate", ExpectedRevision: 1, Snapshot: s, Metrics: m}); err != nil {
		t.Fatal(err)
	}
}
func TestReadyFailsClosedWhenSnapshotBecomesStale(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	snap := NewMockSnapshotVerifier(ctrl)
	policy := NewMockSlicePolicy(ctrl)
	auth := NewMockAuthority(ctrl)
	at := time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)
	e, _ := domain.Create("evaluation:1", "candidate.compatibility", 1, domain.Command{ID: "create", At: at})
	s := domain.Snapshot{ID: "snapshot:1", Version: 1, ConsentVersion: 3, EvaluatedAt: at}
	m := domain.Metrics{Cohort: 300, Evidence: 2000, Quality: .75, ErrorRate: .2, MaxDisparity: .05, Slices: []domain.SliceMetric{{PolicyKey: "approved.region", Cohort: 100, Quality: .74, ErrorRate: .2}}}
	e, _ = e.Record(s, m, domain.Command{ID: "evaluate", ExpectedRevision: 1, At: at})
	e, _ = e.AttachCard(domain.ModelCard{Version: 1, Purpose: "matching.compatibility", EvaluationRef: "report:1", LimitationsRef: "limits:1", Owner: "matching.team"}, domain.Command{ID: "card", ExpectedRevision: 2, At: at})
	e, _ = e.Approve("reviewer:key", at.Add(time.Hour), domain.Command{ID: "approve", ExpectedRevision: 3, At: at})
	repo.EXPECT().Find(gomock.Any(), "evaluation:1").Return(e, nil)
	snap.EXPECT().Revalidate(gomock.Any(), s).Return(errors.New("withdrawn consent"))
	service := NewService(repo, snap, policy, auth, keyer{}, ids{}, func() time.Time { return at })
	ok, err := service.Ready(context.Background(), "evaluation:1")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
