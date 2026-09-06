package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/affiliate/domain"
)

// Payouts stores payout requests.
type Payouts interface {
	Create(context.Context, domain.Payout) error
	Find(ctx context.Context, id string) (domain.Payout, error)
	Save(ctx context.Context, payout domain.Payout, expected domain.PayoutStatus) error
	Pending(ctx context.Context, limit int) ([]domain.Payout, error)
}

// Transfers sends money.
//
// Two steps, because a wrong number should fail at the recipient stage rather
// than after the money has gone.
type Transfers interface {
	CreateRecipient(ctx context.Context, name, phone, network string) (string, error)
	Transfer(ctx context.Context, recipientCode, reference, reason string, amountPesewas int64) (string, error)
}

// PayoutLedger books what leaves.
type PayoutLedger interface {
	RecordPayout(ctx context.Context, reference string, grossMinor, withheldMinor, netMinor int64, at time.Time) error
}

// PayoutService is the desk that pays affiliates.
//
// Deliberately slow: an affiliate requests, an operator approves with a
// step-up, and only then is a transfer initiated. This is the first outbound
// money this product has ever sent, and a scheduled job doing it
// automatically would put real money in real hands with nobody in the loop.
type PayoutService struct {
	affiliates             Repository
	payouts                Payouts
	transfers              Transfers
	ledger                 PayoutLedger
	ids                    IDSource
	minimumPesewas         int64
	withholdingBasisPoints int64
	now                    func() time.Time
}

func NewPayoutService(
	affiliates Repository, payouts Payouts, transfers Transfers, ledger PayoutLedger,
	ids IDSource, minimumPesewas, withholdingBasisPoints int64, now func() time.Time,
) PayoutService {
	if now == nil {
		now = time.Now
	}
	return PayoutService{
		affiliates: affiliates, payouts: payouts, transfers: transfers, ledger: ledger,
		ids: ids, minimumPesewas: minimumPesewas,
		withholdingBasisPoints: withholdingBasisPoints, now: now,
	}
}

// Request opens a payout for everything an affiliate is currently owed.
//
// The whole balance rather than an amount somebody names: a partial payout
// with a withholding line is two filings for one payment, and there is no
// reason an affiliate would want one.
func (service PayoutService) Request(
	ctx context.Context, affiliateID string,
) (domain.Payout, error) {
	if !service.ready() {
		return domain.Payout{}, ErrUnavailable
	}
	affiliate, err := service.affiliates.FindByID(ctx, strings.TrimSpace(affiliateID))
	if err != nil {
		return domain.Payout{}, err
	}
	payout, err := domain.RequestPayout(
		service.ids.NewID(), affiliate.ID(), affiliate.Balance(),
		service.minimumPesewas, service.withholdingBasisPoints, service.now().UTC(),
	)
	if err != nil {
		return domain.Payout{}, err
	}
	if err := service.payouts.Create(ctx, payout); err != nil {
		return domain.Payout{}, ErrUnavailable
	}
	return payout, nil
}

// Approve records an operator's decision and sends the money.
//
// The money moves inside the approval on purpose. An approved payout that
// nothing sent is a promise sitting in a table waiting for somebody to notice,
// and the whole reason this is operator-gated is so a person is present when
// it happens.
//
// The affiliate's balance is reduced *before* the transfer. If the transfer
// then fails the payout is marked failed and the reduction is reversed by
// nothing — it stays spent — because a balance that came back after a
// possibly-sent transfer is how somebody gets paid twice.
func (service PayoutService) Approve(
	ctx context.Context, payoutID, approverKey, phone, network string,
) (domain.Payout, error) {
	if !service.ready() || service.transfers == nil {
		return domain.Payout{}, ErrUnavailable
	}
	payout, err := service.payouts.Find(ctx, strings.TrimSpace(payoutID))
	if err != nil {
		return domain.Payout{}, err
	}
	approved, err := payout.Approve(approverKey, service.now().UTC())
	if err != nil {
		return domain.Payout{}, err
	}
	if err := service.payouts.Save(ctx, approved, domain.PayoutRequested); err != nil {
		return domain.Payout{}, ErrConflict
	}
	affiliate, err := service.affiliates.FindByID(ctx, payout.AffiliateID())
	if err != nil {
		return domain.Payout{}, ErrUnavailable
	}
	// Recorded against the affiliate before anything is sent, so a transfer
	// that succeeds while this process dies cannot be requested again.
	spent, err := affiliate.RecordPayout(payout.GrossPesewas(), domain.Command{
		ID: "payout:" + payout.ID(), ReasonCode: "affiliate_payout", At: service.now().UTC(),
	})
	if err != nil {
		return domain.Payout{}, err
	}
	if spent.Revision() != affiliate.Revision() {
		if err := service.affiliates.Append(
			ctx, spent, affiliate.Revision(), "payout:"+payout.ID(),
		); err != nil {
			return domain.Payout{}, ErrConflict
		}
	}
	recipient, err := service.transfers.CreateRecipient(ctx, affiliate.Name(), phone, network)
	if err != nil {
		return service.markFailed(ctx, approved)
	}
	// The payout's own id is the transfer reference, so a retry cannot send
	// twice: the processor refuses a duplicate reference.
	code, err := service.transfers.Transfer(
		ctx, recipient, payout.ID(), "Obiara affiliate commission", payout.NetPesewas())
	if err != nil {
		return service.markFailed(ctx, approved)
	}
	sent, err := approved.Sent(payout.ID(), code)
	if err != nil {
		return domain.Payout{}, err
	}
	if err := service.payouts.Save(ctx, sent, domain.PayoutApproved); err != nil {
		// The money is gone and the record did not update. Reported so
		// somebody reconciles it rather than silently succeeding.
		return sent, ErrConflict
	}
	if service.ledger != nil {
		_ = service.ledger.RecordPayout(ctx, payout.ID(), payout.GrossPesewas(),
			payout.WithheldPesewas(), payout.NetPesewas(), service.now().UTC())
	}
	return sent, nil
}

// Refuse records an operator declining a payout. Nothing is spent and nothing
// moves.
func (service PayoutService) Refuse(
	ctx context.Context, payoutID, approverKey string,
) (domain.Payout, error) {
	if !service.ready() {
		return domain.Payout{}, ErrUnavailable
	}
	payout, err := service.payouts.Find(ctx, strings.TrimSpace(payoutID))
	if err != nil {
		return domain.Payout{}, err
	}
	refused, err := payout.Refuse(approverKey, service.now().UTC())
	if err != nil {
		return domain.Payout{}, err
	}
	if err := service.payouts.Save(ctx, refused, domain.PayoutRequested); err != nil {
		return domain.Payout{}, ErrConflict
	}
	return refused, nil
}

// Pending is the desk's queue.
func (service PayoutService) Pending(
	ctx context.Context, limit int,
) ([]domain.Payout, error) {
	if !service.ready() {
		return nil, ErrUnavailable
	}
	return service.payouts.Pending(ctx, limit)
}

func (service PayoutService) markFailed(
	ctx context.Context, approved domain.Payout,
) (domain.Payout, error) {
	failed, err := approved.Failed()
	if err != nil {
		return domain.Payout{}, err
	}
	if saveErr := service.payouts.Save(ctx, failed, domain.PayoutApproved); saveErr != nil {
		return domain.Payout{}, ErrConflict
	}
	// The balance stays spent. A balance that came back after a transfer that
	// may or may not have gone out is how somebody gets paid twice; putting
	// it back is an operator's decision with the processor's records in front
	// of them.
	return failed, errors.New("the transfer was not accepted")
}

func (service PayoutService) ready() bool {
	return service.affiliates != nil && service.payouts != nil &&
		service.ids != nil && service.now != nil &&
		service.withholdingBasisPoints > 0
}
