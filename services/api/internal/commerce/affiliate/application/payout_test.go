package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
	"go.uber.org/mock/gomock"
)

func approver() string { return strings.Repeat("b", 64) }

// payoutBook is the payout store.
type payoutBook struct {
	payout  domain.Payout
	saved   []domain.Payout
	findErr error
	saveErr error
	created domain.Payout
}

func (b *payoutBook) Create(_ context.Context, payout domain.Payout) error {
	b.created = payout
	b.payout = payout
	return nil
}

func (b *payoutBook) Find(context.Context, string) (domain.Payout, error) {
	return b.payout, b.findErr
}

func (b *payoutBook) Save(_ context.Context, payout domain.Payout, _ domain.PayoutStatus) error {
	if b.saveErr != nil {
		return b.saveErr
	}
	b.saved = append(b.saved, payout)
	b.payout = payout
	return nil
}

func (b *payoutBook) Pending(context.Context, int) ([]domain.Payout, error) {
	return []domain.Payout{b.payout}, nil
}

// wire is the transfer rail.
type wire struct {
	recipientErr error
	transferErr  error
	sentAmount   int64
	reference    string
	recipients   int
}

func (w *wire) CreateRecipient(context.Context, string, string, string) (string, error) {
	w.recipients++
	return "RCP_1", w.recipientErr
}

func (w *wire) Transfer(
	_ context.Context, _, reference, _ string, amountPesewas int64,
) (string, error) {
	w.sentAmount, w.reference = amountPesewas, reference
	return "TRF_1", w.transferErr
}

func earningAffiliate(t *testing.T, pesewas int64) domain.Affiliate {
	t.Helper()
	affiliate, err := affiliateFor(t).Accrue(strings.Repeat("a", 64), pesewas, domain.Command{
		ID: "accrue:1", ReasonCode: "qualified_conversion", At: now})
	if err != nil {
		t.Fatal(err)
	}
	return affiliate
}

func payoutService(
	ctrl *gomock.Controller, repository *MockRepository, book *payoutBook, rail *wire,
) PayoutService {
	return NewPayoutService(repository, book, rail, nil, fixedID("pay_1"),
		20_000, 750, func() time.Time { return now })
}

func TestAPayoutIsForTheWholeBalance(t *testing.T) {
	// A partial payout with a withholding line is two filings for one
	// payment, and there is no reason an affiliate would want one.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(earningAffiliate(t, 100_000), nil)
	book := &payoutBook{}

	payout, err := payoutService(ctrl, repository, book, &wire{}).
		Request(context.Background(), "aff_1")
	if err != nil {
		t.Fatal(err)
	}
	if payout.GrossPesewas() != 100_000 {
		t.Fatalf("gross = %d", payout.GrossPesewas())
	}
	if payout.WithheldPesewas() != 7_500 || payout.NetPesewas() != 92_500 {
		t.Fatalf("withheld %d net %d", payout.WithheldPesewas(), payout.NetPesewas())
	}
}

func TestABalanceBelowTheFloorCannotBeRequested(t *testing.T) {
	// So the platform is not sending transfer fees to move a cedi.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(earningAffiliate(t, 100), nil)

	if _, err := payoutService(ctrl, repository, &payoutBook{}, &wire{}).
		Request(context.Background(), "aff_1"); !errors.Is(err, domain.ErrBelowMinimum) {
		t.Fatalf("err = %v, want ErrBelowMinimum", err)
	}
}

func TestApprovingSendsTheNetAndSpendsTheBalanceFirst(t *testing.T) {
	// Spent before the transfer, so a transfer that succeeds while this
	// process dies cannot be requested again.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	affiliate := earningAffiliate(t, 100_000)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(affiliate, nil).Times(2)
	var spentBalance int64 = -1
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), affiliate.Revision(), gomock.Any()).
		DoAndReturn(func(_ context.Context, a domain.Affiliate, _ uint64, _ string) error {
			spentBalance = a.Balance()
			return nil
		})
	book := &payoutBook{}
	rail := &wire{}
	service := payoutService(ctrl, repository, book, rail)
	if _, err := service.Request(context.Background(), "aff_1"); err != nil {
		t.Fatal(err)
	}

	sent, err := service.Approve(context.Background(), "pay_1", approver(), "0551234987", "mtn")
	if err != nil {
		t.Fatal(err)
	}
	if sent.Status() != domain.PayoutSent {
		t.Fatalf("status = %q", sent.Status())
	}
	// The affiliate is paid the net; the withholding stays behind.
	if rail.sentAmount != 92_500 {
		t.Fatalf("sent %d, want the net", rail.sentAmount)
	}
	// The whole gross is spent against the balance, because the withholding
	// is money the platform owes the revenue authority rather than money it
	// keeps.
	if spentBalance != 0 {
		t.Fatalf("balance after payout = %d", spentBalance)
	}
	// The payout's own id is the transfer reference, so a retry cannot send
	// twice.
	if rail.reference != "pay_1" {
		t.Fatalf("reference = %q", rail.reference)
	}
}

func TestATransferTheProcessorWillNotAcceptLeavesTheBalanceSpent(t *testing.T) {
	// A balance that came back after a transfer that may or may not have gone
	// out is how somebody gets paid twice. Putting it back is an operator's
	// decision with the processor's records in front of them.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	affiliate := earningAffiliate(t, 100_000)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(affiliate, nil).Times(2)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	book := &payoutBook{}
	rail := &wire{transferErr: errors.New("paystack refused")}
	service := payoutService(ctrl, repository, book, rail)
	if _, err := service.Request(context.Background(), "aff_1"); err != nil {
		t.Fatal(err)
	}

	if _, err := service.Approve(
		context.Background(), "pay_1", approver(), "0551234987", "mtn",
	); err == nil {
		t.Fatal("a refused transfer reported success")
	}
	if book.payout.Status() != domain.PayoutFailed {
		t.Fatalf("status = %q, want failed", book.payout.Status())
	}
}

func TestARefusedPayoutSpendsNothing(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().FindByID(gomock.Any(), "aff_1").Return(earningAffiliate(t, 100_000), nil)
	// No Append expectation: nothing is spent.
	book := &payoutBook{}
	rail := &wire{}
	service := payoutService(ctrl, repository, book, rail)
	if _, err := service.Request(context.Background(), "aff_1"); err != nil {
		t.Fatal(err)
	}

	refused, err := service.Refuse(context.Background(), "pay_1", approver())
	if err != nil {
		t.Fatal(err)
	}
	if refused.Status() != domain.PayoutRefused {
		t.Fatalf("status = %q", refused.Status())
	}
	if rail.recipients != 0 {
		t.Fatal("a refused payout still reached the transfer rail")
	}
}

func TestNothingIsPaidWithoutAWithholdingRate(t *testing.T) {
	// The service refuses to run at all, rather than paying gross and
	// leaving a filing problem for somebody to find later.
	ctrl := gomock.NewController(t)
	unset := NewPayoutService(NewMockRepository(ctrl), &payoutBook{}, &wire{}, nil,
		fixedID("pay_1"), 20_000, 0, func() time.Time { return now })
	if _, err := unset.Request(context.Background(), "aff_1"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
