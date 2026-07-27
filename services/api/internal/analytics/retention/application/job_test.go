package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/analytics/retention/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestRunRevalidatesPolicyAndAppliesBoundedActions(t *testing.T) {
	ctrl := gomock.NewController(t)
	catalog := NewMockPolicyCatalog(ctrl)
	store := NewMockStore(ctrl)
	pseudonyms := NewMockPseudonymizer(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	now := time.Date(2026, 7, 27, 2, 0, 0, 0, time.UTC)
	policy, _ := domain.NewPolicy(domain.PolicySpec{ID: "analytics.retention", ReviewID: "privacy.review", ReviewerKey: key(1), Version: 1, PseudonymKeyVersion: 2, ReviewedAt: now, BatchSize: 2})
	old := domain.Candidate{ID: "507f1f77bcf86cd799439011", Name: "epono.pod_heard", SubjectRef: key(2), OccurredAt: now.Add(-100 * 24 * time.Hour)}
	ancient := domain.Candidate{ID: "507f191e810c19729de860ea", Name: "epono.seed_sown", SubjectRef: key(3), OccurredAt: now.AddDate(0, -13, 0)}
	catalog.EXPECT().Current(gomock.Any()).Return(policy, nil)
	clock.EXPECT().Now().Return(now)
	ids.EXPECT().NewID().Return(key(4))
	store.EXPECT().ClaimDue(gomock.Any(), now, 2, key(4), now.Add(5*time.Minute)).Return([]domain.Candidate{old, ancient}, nil)
	catalog.EXPECT().Current(gomock.Any()).Return(policy, nil)
	ids.EXPECT().NewID().Return(key(5))
	pseudonyms.EXPECT().Derive(key(2), uint64(2)).Return(key(6), nil)
	store.EXPECT().Pseudonymize(gomock.Any(), old, gomock.Any(), key(6), key(5)).Return(nil)
	catalog.EXPECT().Current(gomock.Any()).Return(policy, nil)
	ids.EXPECT().NewID().Return(key(7))
	store.EXPECT().AggregateErase(gomock.Any(), ancient, gomock.Any(), key(7)).Return(nil)
	r, e := New(catalog, store, pseudonyms, ids, clock).Run(context.Background())
	if e != nil || r.Claimed != 2 || r.Pseudonymized != 1 || r.Aggregated != 1 {
		t.Fatal(r, e)
	}
}
