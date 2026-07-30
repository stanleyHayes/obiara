package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/domain"
)

var packSvcNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func newService(t *testing.T) (MarketPackService, *MockPackRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	packs := NewMockPackRepository(ctrl)
	return NewMarketPackService(packs, func() time.Time { return packSvcNow }, func() string { return "pack_test" }), packs
}

func TestDraftCreatesAndAudits(t *testing.T) {
	service, packs := newService(t)
	packs.EXPECT().CreateWithAudit(gomock.Any(), gomock.Any(), "proposer-1", "marketpack.draft", gomock.Any()).Return(nil)

	pack, err := service.Draft(context.Background(), domain.MarketGhanaPidgin, "term:gh-pidgin:1", map[string]bool{"fires": true}, "proposer-1")
	if err != nil {
		t.Fatal(err)
	}
	if pack.Status() != domain.StatusDraft || pack.ProposedBy() != "proposer-1" {
		t.Fatalf("pack = %#v", pack)
	}
}

func TestPublishFlow(t *testing.T) {
	service, packs := newService(t)
	draft, _ := domain.NewPack("p-1", domain.MarketGhanaTwi, "term:gh-tw:1", nil, "proposer-1", packSvcNow)
	packs.EXPECT().FindByID(gomock.Any(), "p-1").Return(draft, nil)
	packs.EXPECT().UpdateWithAudit(gomock.Any(), gomock.Any(), "approver-1", "marketpack.publish", gomock.Any()).DoAndReturn(
		func(_ context.Context, pack domain.MarketPack, _, _ string, _ time.Time) error {
			if pack.Status() != domain.StatusPublished || pack.ApprovedBy() != "approver-1" {
				t.Fatalf("pack = %#v", pack)
			}
			return nil
		})

	if _, err := service.Publish(context.Background(), "p-1", "approver-1"); err != nil {
		t.Fatal(err)
	}
}

func TestPublishSelfApprovalRejected(t *testing.T) {
	service, packs := newService(t)
	draft, _ := domain.NewPack("p-1", domain.MarketGhanaTwi, "term:gh-tw:1", nil, "proposer-1", packSvcNow)
	packs.EXPECT().FindByID(gomock.Any(), "p-1").Return(draft, nil)
	// No Update/audit expectation.

	if _, err := service.Publish(context.Background(), "p-1", "proposer-1"); err != domain.ErrSelfApproval {
		t.Fatalf("Publish = %v, want self-approval rejected", err)
	}
}
