package sponsorship

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

// Book posts what an organization pays and what its members take.
//
// The two are deliberately different postings, because they are different
// events. A deposit is money received against nothing delivered — an asset
// arrives and a liability arrives with it. A draw is the moment a seat is
// actually given, which is when that liability becomes revenue.
//
// Booking a deposit as revenue would recognise income for something not yet
// delivered, and would leave nothing to recognise when the seat was taken.
type Book struct {
	poster Poster
	actor  string
}

func NewBook(poster Poster, actor string) Book {
	return Book{poster: poster, actor: actor}
}

const (
	accountCash             = "sponsorship_cash_received"
	accountOrganizationHeld = "organization_deposits_held"
	accountMembershipSales  = "membership_revenue"
)

// RecordDeposit books money an organization paid: cash up, and a liability up
// with it, because the platform is holding somebody else's money.
func (book Book) RecordDeposit(
	ctx context.Context, reference string, minor int64, _ time.Time,
) error {
	return book.post(ctx, "sponsorship_deposit:"+reference, reference, minor,
		accountCash, ledgerdomain.ClassAsset,
		accountOrganizationHeld, ledgerdomain.ClassLiability)
}

// RecordDraw books a seat taken: the liability goes down and revenue goes up.
// Nothing moves in or out of the bank here — the money arrived at deposit.
func (book Book) RecordDraw(
	ctx context.Context, reference string, minor int64, _ time.Time,
) error {
	return book.post(ctx, "sponsorship_draw:"+reference, reference, minor,
		accountOrganizationHeld, ledgerdomain.ClassLiability,
		accountMembershipSales, ledgerdomain.ClassRevenue)
}

// post writes one balanced pair. Debits equal credits by construction: there
// is one amount and it appears on both sides.
func (book Book) post(
	ctx context.Context, commandID, reference string, minor int64,
	debitAccount string, debitClass ledgerdomain.AccountClass,
	creditAccount string, creditClass ledgerdomain.AccountClass,
) error {
	if book.poster == nil || minor <= 0 {
		return nil
	}
	_, err := book.poster.Post(ctx, ledgerapplication.PostCommand{
		Actor: book.actor, CommandID: commandID, ReferenceID: reference,
		Purpose:  ledgerdomain.PurposeSaleSettlement,
		Currency: ledgerdomain.Currency("GHS"),
		Lines: []ledgerapplication.Line{
			{AccountID: debitAccount, Class: debitClass,
				Side: ledgerdomain.SideDebit, Minor: minor},
			{AccountID: creditAccount, Class: creditClass,
				Side: ledgerdomain.SideCredit, Minor: minor},
		},
	})
	return err
}
