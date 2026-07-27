package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/fairness/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

const (
	ka = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	kb = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	kc = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	kd = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
)

type ids struct{}

func (ids) NewID() string { return kd }

type clock struct{}

func (clock) Now() time.Time { return time.Unix(3, 0) }

func TestProjectRevalidatesDefinitionAndAuthorityBeforeImmutableInsert(t *testing.T) {
	ctrl := gomock.NewController(t)
	defs, source, repo, authority := NewMockDefinitionCatalog(ctrl), NewMockAggregateSource(ctrl), NewMockRepository(ctrl), NewMockAuthority(ctrl)
	d, _ := domain.NewDefinition(domain.DefinitionSpec{ID: "fairness.v1", ReviewID: "review.v1", ReviewerKey: ka, Version: 2, MaxParityGapPermille: 50, ReviewedAt: time.Unix(1, 0)})
	snapshot := domain.Snapshot{ID: ka, QuarterKey: kb, SourceWatermark: kc, Version: 3, WindowStartedAt: time.Unix(1, 0), WindowEndedAt: time.Unix(2, 0), Cohorts: []domain.CohortAggregate{{CohortKey: ka, Eligible: 100, Exposed: 50}, {CohortKey: kb, Eligible: 100, Exposed: 55}}, PreviousRegretEligible: 100, PreviousRegretReports: 2, CurrentRegretEligible: 100, CurrentRegretReports: 1, ColorismAuditComplete: true, CompleteMetrics: []domain.Metric{domain.MetricExposureParity, domain.MetricColorismDrift, domain.MetricRegretTrend, domain.MetricTierASafety}}
	authority.EXPECT().RequireProjector(gomock.Any(), "operator").Times(2)
	defs.EXPECT().Current(gomock.Any()).Return(d, nil).Times(2)
	source.EXPECT().CurrentQuarter(gomock.Any(), kb).Return(snapshot, nil)
	repo.EXPECT().Insert(gomock.Any(), gomock.AssignableToTypeOf(domain.Report{}))
	service := New(defs, source, repo, authority, ids{}, clock{})
	report, err := service.Project(context.Background(), "operator", kb, 2, 3)
	if err != nil || report.Outcome != domain.OutcomePass {
		t.Fatalf("outcome=%s err=%v", report.Outcome, err)
	}
}
