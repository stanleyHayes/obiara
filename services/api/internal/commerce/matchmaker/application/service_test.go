package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestBookRequiresCurrentLicense(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := NewMockRepository(ctrl)
	l := NewMockLicenseCatalog(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	license := domain.License{ID: "license.gh", MatchmakerKey: key(2), Jurisdiction: "ghana", Version: 2, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour), MinimumFeePesewas: 1000, MaximumFeePesewas: 5000}
	terms := domain.Terms{ID: "terms.1", Version: 1, TotalFeePesewas: 1000, Milestones: []domain.Milestone{{ID: "consult", FeePesewas: 1000}}}
	l.EXPECT().Current(gomock.Any(), key(2)).Return(license, nil)
	ids.EXPECT().NewID().Return(key(1))
	clock.EXPECT().Now().Return(now)
	r.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	if _, e := New(r, l, ids, clock).Book(context.Background(), key(3), key(2), terms, "book-1"); e != nil {
		t.Fatal(e)
	}
}
