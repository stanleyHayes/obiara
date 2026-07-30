package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
)

func TestOverviewReturnsOnlyExceptionAuditsWithTheirFacts(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	key := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	fact, err := domain.NewFact("fact-1", key, key, key, "ledger-1", domain.CurrencyGHS, domain.StatusSettled, 12000, now, now)
	if err != nil {
		t.Fatal(err)
	}
	exception := domain.Compare(fact, domain.LedgerProof{}, false)
	exceptionAudit, _ := domain.NewAudit("audit-1", fact, exception, now)
	reconciledAudit, _ := domain.NewAudit("audit-2", fact, domain.Compare(fact, domain.LedgerProof{
		CommandID: "ledger-1", ReferenceKey: key, Currency: domain.CurrencyGHS, Minor: 12000, Balanced: true,
	}, true), now)
	checkpoint, _ := domain.NewCheckpoint("checkpoint-1", "2026-07-30", 2, 1, 1, now)

	repo.EXPECT().ListRecentAudits(gomock.Any(), 50).Return([]domain.Audit{reconciledAudit, exceptionAudit}, nil)
	repo.EXPECT().FindFactByID(gomock.Any(), "fact-1").Return(fact, nil)
	repo.EXPECT().ListRecentCheckpoints(gomock.Any(), 14).Return([]domain.Checkpoint{checkpoint}, nil)

	overview, err := NewQueryService(repo).Overview(context.Background(), 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Exceptions) != 1 || overview.Exceptions[0].Exception != domain.ExceptionLedgerMissing {
		t.Fatalf("exceptions = %#v", overview.Exceptions)
	}
	if len(overview.Checkpoints) != 1 || overview.Checkpoints[0].Day() != "2026-07-30" {
		t.Fatalf("checkpoints = %#v", overview.Checkpoints)
	}
}
