package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/p0gate/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func fixtures(t *testing.T) (domain.Definition, domain.Snapshot) {
	t.Helper()
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	d, e := domain.NewDefinition(domain.DefinitionSpec{ID: "p0.gates", Version: 1, ReviewID: "founder.gates", ReviewerKey: key(1), ReviewedAt: now, PodsHeardPermille: 650, SeedToSproutPermille: 250, SproutToRoomPermille: 350, WeeklyFirePermille: 400, Day30RetentionPermille: 450})
	if e != nil {
		t.Fatal(e)
	}
	s := domain.Snapshot{ID: key(2), WindowKey: key(3), SourceWatermark: key(4), Version: 1, WindowStartedAt: now.Add(-time.Hour), WindowEndedAt: now, CohortSize: 10, PodEligible: 10, PodsHeard: 10, SeedsSown: 10, SproutsOpened: 10, SproutEligible: 10, RoomsOpened: 10, WeeklyFireAttendees: 10, Day30Eligible: 10, Day30Retained: 10, PreviousRegretReports: 2, CurrentRegretReports: 1, CompleteMetrics: []domain.Metric{domain.MetricPodsHeard, domain.MetricSeedToSprout, domain.MetricSproutToRoom, domain.MetricWeeklyFire, domain.MetricDay30Retention, domain.MetricRegretTrend, domain.MetricTierAResolved}}
	return d, s
}
func TestCurrentDefinitionRevalidatedBeforeInsert(t *testing.T) {
	ctrl := gomock.NewController(t)
	d := NewMockDefinitionCatalog(ctrl)
	source := NewMockAggregateSource(ctrl)
	repo := NewMockRepository(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	definition, snapshot := fixtures(t)
	now := snapshot.WindowEndedAt
	gomock.InOrder(d.EXPECT().Current(gomock.Any()).Return(definition, nil), source.EXPECT().Current(gomock.Any(), key(3)).Return(snapshot, nil), ids.EXPECT().NewID().Return(key(5)), clock.EXPECT().Now().Return(now), d.EXPECT().Current(gomock.Any()).Return(definition, nil), repo.EXPECT().Insert(gomock.Any(), gomock.Any()).Return(nil))
	r, e := New(d, source, repo, ids, clock).Project(context.Background(), key(3), 1, 1)
	if e != nil || r.Outcome != domain.OutcomePass {
		t.Fatal(r, e)
	}
}
