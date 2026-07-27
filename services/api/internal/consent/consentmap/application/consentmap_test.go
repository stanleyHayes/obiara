package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
)

var consentNow = time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC)

func TestStateForFallsBackToDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	states := NewMockStateStore(ctrl)
	service := NewConsentMapService(states, nil, func() time.Time { return consentNow }, func() string { return "cons_test" })

	states.EXPECT().Get(gomock.Any(), "m-1", domain.PurposeScamArc).Return(nil, nil)
	on, err := service.StateFor(context.Background(), "m-1", domain.PurposeScamArc)
	if err != nil || !on {
		t.Fatalf("state = %v, %v, want default-on", on, err)
	}
}

func TestSetValidatesAndReceipts(t *testing.T) {
	ctrl := gomock.NewController(t)
	states := NewMockStateStore(ctrl)
	receipts := NewMockReceiptStore(ctrl)
	service := NewConsentMapService(states, receipts, func() time.Time { return consentNow }, func() string { return "cons_test" })

	states.EXPECT().Set(gomock.Any(), "m-1", domain.PurposeMatching, true).Return(nil)
	receipts.EXPECT().Append(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, receipt domain.Receipt) error {
			if receipt.Purpose != domain.PurposeMatching || !receipt.Enabled || receipt.ID != "cons_test" {
				t.Fatalf("receipt = %#v", receipt)
			}
			return nil
		})

	if _, err := service.Set(context.Background(), "m-1", domain.PurposeMatching, true); err != nil {
		t.Fatal(err)
	}
}

func TestSetRejectsLockedPurpose(t *testing.T) {
	ctrl := gomock.NewController(t)
	states := NewMockStateStore(ctrl)
	receipts := NewMockReceiptStore(ctrl)
	service := NewConsentMapService(states, receipts, func() time.Time { return consentNow }, func() string { return "cons_test" })
	// No Set or Append expectation.

	if _, err := service.Set(context.Background(), "m-1", domain.PurposeIdentitySafety, false); err != domain.ErrPurposeLocked {
		t.Fatalf("Set = %v, want locked", err)
	}
}

func TestSwitchboardMergesDefaultsAndChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	states := NewMockStateStore(ctrl)
	service := NewConsentMapService(states, nil, func() time.Time { return consentNow }, func() string { return "cons_test" })

	states.EXPECT().AllForMember(gomock.Any(), "m-1").Return(map[domain.Purpose]bool{
		domain.PurposeMatching: true, domain.PurposeScamArc: false,
	}, nil)

	board, err := service.Switchboard(context.Background(), "m-1")
	if err != nil {
		t.Fatal(err)
	}
	if !board[domain.PurposeIdentitySafety] || !board[domain.PurposeMatching] ||
		board[domain.PurposeScamArc] || board[domain.PurposePlayPortraits] || !board[domain.PurposeProductAnalytics] {
		t.Fatalf("board = %#v", board)
	}
}
