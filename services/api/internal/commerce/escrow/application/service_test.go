package application

import (
	"context"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func TestSettlementRecordsExplicitStatementInLedger(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	ledger := NewMockLedger(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	now := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	x, _ := domain.Fund(key(1), key(2), 1000, domain.Terms{ID: "terms.1", Version: 1, Milestones: []domain.MilestoneTerm{{ID: "one", GrossPesewas: 1000, FeePesewas: 100}}}, "fund-1", now)
	x, _ = x.AddEvidence("one", domain.DeliveryEvidence, key(3), "delivery-1", now)
	x, _ = x.AddEvidence("one", domain.AcceptanceEvidence, key(4), "accept-1", now)
	repo.EXPECT().Find(gomock.Any(), key(1)).Return(x, nil)
	ids.EXPECT().NewID().Return(key(5))
	clock.EXPECT().Now().Return(now)
	ledger.EXPECT().RecordSettlement(gomock.Any(), "settle-1", gomock.Any()).DoAndReturn(func(_ context.Context, _ string, s domain.Statement) error {
		if s.NetPesewas != 900 || s.FeePesewas != 100 {
			t.Fatal(s)
		}
		return nil
	})
	repo.EXPECT().Save(gomock.Any(), gomock.Any(), x.Revision(), "settle-1").Return(nil)
	if _, _, e := New(repo, ledger, ids, clock).Settle(context.Background(), key(1), "one", "settle-1"); e != nil {
		t.Fatal(e)
	}
}
