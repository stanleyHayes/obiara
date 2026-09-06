package purchase

import (
	"context"
	"time"

	ledgerapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/application"
	ledgerdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

// Poster is the double-entry book.
type Poster interface {
	Post(context.Context, ledgerapplication.PostCommand) (ledgerdomain.Posting, error)
}

// SaleBook posts a membership sale.
//
// Two lines, because that is what a sale is: cash held at the provider goes
// up, and membership revenue goes up with it. Posting one side would balance
// nowhere, and posting none would make revenue something inferred from passes
// existing rather than something written down.
//
// The accounts are named rather than keyed by member. A ledger that carried a
// member key per line would be a purchase history in the accounts, and the
// reference — the intent id — is already enough to trace one sale.
type SaleBook struct {
	poster Poster
	// actor is who the ledger records as posting. Settlement is done by the
	// system on a provider's callback, not by a person.
	actor string
}

func NewSaleBook(poster Poster, actor string) SaleBook {
	return SaleBook{poster: poster, actor: actor}
}

const (
	accountProviderCash    = "momo_settlement_receivable"
	accountMembershipSales = "membership_revenue"
)

func (book SaleBook) RecordSale(
	ctx context.Context, reference string, minor int64, currency string, _ time.Time,
) error {
	if book.poster == nil || minor <= 0 {
		return ErrUnavailable
	}
	_, err := book.poster.Post(ctx, ledgerapplication.PostCommand{
		Actor: book.actor, CommandID: "membership_sale:" + reference, ReferenceID: reference,
		Purpose:  ledgerdomain.PurposeSaleSettlement,
		Currency: ledgerdomain.Currency(currency),
		Lines: []ledgerapplication.Line{
			{AccountID: accountProviderCash, Class: ledgerdomain.ClassAsset,
				Side: ledgerdomain.SideDebit, Minor: minor},
			{AccountID: accountMembershipSales, Class: ledgerdomain.ClassRevenue,
				Side: ledgerdomain.SideCredit, Minor: minor},
		},
	})
	return err
}
