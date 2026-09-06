// Package purchase is what makes a membership something a member can buy.
//
// It is a bridge and not a context: it owns no aggregate and no storage of its
// own. The catalog says what a pass costs, the payment context collects the
// money, and the membership context grants the pass. What was missing was
// anything joining them — `membership.Service.Grant` had no callers at all, so
// no pass could ever exist and nothing in the product could be bought
// (agent_plan.md §72).
package purchase

import (
	"context"
	"errors"
	"strings"
	"time"

	catalogdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
	membershipdomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/domain"
	momoapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/application"
	momodomain "github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
	promotionapplication "github.com/stanleyHayes/obiara/services/api/internal/commerce/promotion/application"
)

// Period and Grace are what one payment buys.
//
// Thirty days is short enough that a member is never far from a decision to
// keep paying and long enough not to nag. The week of grace covers a failed
// charge or a phone with no credit without cutting anybody off the same day.
const (
	Period = 30 * 24 * time.Hour
	Grace  = 7 * 24 * time.Hour
)

var (
	// ErrUnavailable means the purchase could not be started. It never means
	// a payment might have happened.
	ErrUnavailable = errors.New("membership purchase unavailable")
	// ErrNotPurchasable refuses a SKU that is not a published membership in
	// the currency this rail collects.
	ErrNotPurchasable = errors.New("that is not a membership anybody can buy")
	// ErrSponsorshipUnavailable reports a sponsored code whose organization
	// cannot cover the seat. The member can still buy their own membership,
	// so this is a refusal of the sponsorship rather than of the purchase.
	ErrSponsorshipUnavailable = errors.New("that sponsorship is not available right now")
	// ErrAlreadySettled reports a callback for an intent that is already
	// decided. It is not an error to the provider — retries are normal — so
	// the caller answers it as success.
	ErrAlreadySettled = errors.New("that payment was already settled")
)

// Catalog says what a pass costs.
type Catalog interface {
	ReadPublished(ctx context.Context, sku string, version uint64) (catalogdomain.SKU, error)
}

// Payments collects the money.
type Payments interface {
	Create(ctx context.Context, memberKey, phoneRef string, amount uint64, command string) (momodomain.Intent, error)
	Confirm(ctx context.Context, id, command string, payer momoapplication.Payer) (momodomain.Intent, error)
	Settle(ctx context.Context, callbackID, intentID, providerRef string, success bool) (momodomain.Intent, error)
	Find(ctx context.Context, id string) (momodomain.Intent, error)
}

// Members is where a receipt goes.
//
// A payment processor needs an email: it is where a receipt is sent and what a
// dispute is attached to. Handing it over is a deliberate disclosure to the
// processor the member is paying through, and to nobody else — it is not
// written into any row this product keeps about the payment.
type Members interface {
	Email(ctx context.Context, memberID string) (string, error)
}

// Passes grants and cancels membership.
type Passes interface {
	Grant(ctx context.Context, memberKey, passID, receiptRef string, passVersion uint64,
		paidThrough time.Time, grace time.Duration, commandID string) (membershipdomain.Pass, error)
	Cancel(ctx context.Context, id, commandID string) (membershipdomain.Pass, error)
}

// Keyer turns a member id and a phone number into the digests the payment and
// membership contexts store. A raw phone number in a payment row would be a
// contact directory; a raw member id would be a purchase history.
//
// Two named methods rather than one namespaced Key, because the two digests
// come from different contexts with different secrets and must not be
// interchangeable — a namespace string is a comment the compiler does not read.
type Keyer interface {
	MemberKey(memberID string) (string, error)
	PhoneRef(phone string) (string, error)
}

// Discounts applies a code to a price. It is optional: a deployment with no
// promotion context composed simply charges everybody full price, which is
// what the product did before codes existed.
type Discounts interface {
	Apply(ctx context.Context, code, memberID, skuID string, priceMinor int64, commandID string) (promotionapplication.Applied, error)
}

// Sponsors draw the price of a seat from an organization's funded balance.
//
// Optional: a deployment with no sponsorship context composed simply has no
// sponsored codes, and every purchase is paid by the member.
type Sponsors interface {
	Draw(ctx context.Context, organizationID, seatRef string, amountPesewas int64) (bool, error)
}

// Ledger records the money. Purchases post through the same double-entry book
// as everything else, so revenue is visible rather than implied by a pass
// appearing.
type Ledger interface {
	RecordSale(ctx context.Context, reference string, minor int64, currency string, at time.Time) error
}

type Service struct {
	catalog   Catalog
	payments  Payments
	passes    Passes
	keyer     Keyer
	orders    Orders
	members   Members
	ledger    Ledger
	discounts Discounts
	sponsors  Sponsors
	now       func() time.Time
}

// WithSponsors attaches organization-funded seats.
func (service Service) WithSponsors(sponsors Sponsors) Service {
	service.sponsors = sponsors
	return service
}

// WithOrders attaches the record of what each collection was opened to buy.
// Without it settlement cannot know what to grant, so a service composed
// without it refuses rather than guessing.
func (service Service) WithOrders(orders Orders) Service {
	service.orders = orders
	return service
}

// WithMembers attaches the lookup that finds where a receipt goes. Without it
// nothing can be charged, because a processor will not open a collection
// without somewhere to send one.
func (service Service) WithMembers(members Members) Service {
	service.members = members
	return service
}

// WithDiscounts attaches discount codes. Without it every purchase is at full
// price, which is a working product and not a broken one.
func (service Service) WithDiscounts(discounts Discounts) Service {
	service.discounts = discounts
	return service
}

func New(
	catalog Catalog, payments Payments, passes Passes, keyer Keyer, ledger Ledger,
	now func() time.Time,
) Service {
	if now == nil {
		now = time.Now
	}
	return Service{
		catalog: catalog, payments: payments, passes: passes,
		keyer: keyer, ledger: ledger, now: now,
	}
}

type StartCommand struct {
	CommandID  string
	MemberID   string
	SKUID      string
	SKUVersion uint64
	// Phone is the number the provider prompts. It is keyed before it reaches
	// the payment context and never stored raw.
	Phone string
	// Network is which mobile money provider the number is on. The processor
	// needs it and cannot infer it reliably from the number.
	Network string
	// Code is an optional discount code. A code that does not apply is not an
	// error: the member came to buy a membership, and a typo should not stop
	// them.
	Code string
}

type Started struct {
	IntentID string
	Status   string
	// AmountPesewas is what the member is about to be asked for, echoed back
	// so a client can show it rather than guess at it.
	AmountPesewas uint64
	// DiscountPesewas is what a code took off, zero when none applied. Echoed
	// back so a member sees the discount they were given rather than having
	// to infer it from a smaller number.
	DiscountPesewas uint64
	// Code is the code that applied, empty when none did.
	Code string
}

// Start prices the pass, opens a payment intent and asks the provider to
// prompt the member.
//
// Nothing is granted here. A member who never approves the prompt has not
// paid, and the only thing that says otherwise is the provider's signed
// callback.
func (service Service) Start(ctx context.Context, command StartCommand) (Started, error) {
	if !service.ready() {
		return Started{}, ErrUnavailable
	}
	if strings.TrimSpace(command.CommandID) == "" || strings.TrimSpace(command.MemberID) == "" ||
		strings.TrimSpace(command.Phone) == "" || strings.TrimSpace(command.Network) == "" {
		return Started{}, ErrUnavailable
	}
	sku, err := service.catalog.ReadPublished(
		ctx, strings.TrimSpace(command.SKUID), command.SKUVersion)
	if err != nil {
		return Started{}, ErrNotPurchasable
	}
	// GHS only, because that is what this rail collects. A price in another
	// currency would be charged as the right number of the wrong unit.
	if sku.Price().Currency != catalogdomain.CurrencyGHS || sku.Price().Minor <= 0 {
		return Started{}, ErrNotPurchasable
	}
	price := sku.Price().Minor
	var applied promotionapplication.Applied
	if service.discounts != nil {
		applied, err = service.discounts.Apply(
			ctx, command.Code, command.MemberID, sku.ID(), price, command.CommandID+":code")
		if err != nil {
			return Started{}, ErrUnavailable
		}
		price -= applied.DiscountMinor
	}
	// A sponsored seat is paid by the organization, not by the member. The
	// price is drawn from their funded balance and the pass is granted here:
	// there is no prompt to send and nothing to wait for, because the money
	// arrived when the organization deposited it.
	if applied.Sponsored && service.sponsors != nil {
		return service.sponsoredSeat(ctx, command, sku, applied)
	}
	// A free membership is not a payment. Nothing here can collect zero, and
	// a code that takes the whole price is a decision somebody made rather
	// than a purchase to push through a payment rail.
	//
	// A sponsored code reaching here means the sponsorship context is not
	// composed, so the code covers the price with nobody paying it. Refused
	// rather than given away.
	if price <= 0 {
		return Started{}, ErrNotPurchasable
	}
	if service.members == nil {
		return Started{}, ErrUnavailable
	}
	email, err := service.members.Email(ctx, command.MemberID)
	if err != nil || strings.TrimSpace(email) == "" {
		// No receipt address means no collection. A processor will not open
		// one without it, and inventing an address would send a member's
		// receipt into a hole.
		return Started{}, ErrUnavailable
	}
	memberKey, err := service.keyer.MemberKey(command.MemberID)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	phoneRef, err := service.keyer.PhoneRef(command.Phone)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	intent, err := service.payments.Create(
		ctx, memberKey, phoneRef, uint64(price), command.CommandID)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	// Confirmed in the same request because the member is standing there: the
	// deliberate gesture is the purchase itself, and a second round trip only
	// adds a place for it to be abandoned.
	//
	// The raw phone goes in here and nowhere else. The intent stored a digest
	// of it, which is right for a row that outlives the payment and cannot be
	// dialled — so the number is supplied now and checked against that digest.
	confirmed, err := service.payments.Confirm(
		ctx, intent.State().ID, command.CommandID+":confirm",
		momoapplication.Payer{Phone: command.Phone, Email: email, Network: command.Network},
	)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	// Written after the prompt is out, because an order for a collection that
	// was never opened is a row describing nothing. A failure here is
	// deliberately fatal to the purchase: a payment nobody can attribute to a
	// product would take a member's money and leave settlement guessing.
	if err := service.orders.Record(ctx, Order{
		IntentID: confirmed.State().ID, SKUKey: sku.SKUKey(), SKUVersion: sku.Version(),
		MemberID: strings.TrimSpace(command.MemberID), AmountPesewas: price,
		Code: applied.Code,
	}); err != nil {
		return Started{}, ErrUnavailable
	}
	return Started{
		IntentID:        confirmed.State().ID,
		Status:          string(confirmed.State().Status),
		AmountPesewas:   uint64(price),
		DiscountPesewas: uint64(applied.DiscountMinor),
		Code:            applied.Code,
	}, nil
}

// Outcome is what a verified webhook said. The caller has already
// authenticated the processor over the exact bytes it sent.
type Outcome struct {
	// CallbackID is what makes a retried webhook settle once. Processors all
	// retry, so this is not optional.
	CallbackID string
	// Reference is the reference this product gave the processor, which is
	// the intent's own id.
	Reference     string
	Success       bool
	AmountPesewas int64
	Currency      string
}

// ErrWrongAmount refuses an outcome reporting a different amount from the one
// the collection was opened for.
var ErrWrongAmount = errors.New("that payment is not for what was asked")

// Settle applies a verified outcome.
//
// On success the pass is granted; on a failure or reversal an already-granted
// pass is cancelled rather than removed, so the trail shows
// granted-then-cancelled instead of a pass that quietly vanished.
func (service Service) Settle(ctx context.Context, outcome Outcome) error {
	if !service.ready() {
		return ErrUnavailable
	}
	if strings.TrimSpace(outcome.CallbackID) == "" ||
		strings.TrimSpace(outcome.Reference) == "" {
		return ErrUnavailable
	}
	// What was asked for, read before anything is granted. A signed webhook
	// is authentic, and authentic is not the same as correct: an outcome
	// reporting a smaller amount than the collection was opened for would
	// otherwise buy a whole month for whatever the payer felt like sending.
	intent, err := service.payments.Find(ctx, outcome.Reference)
	if err != nil {
		return err
	}
	expected := intent.State()
	if outcome.Success {
		if outcome.AmountPesewas != int64(expected.AmountPesewas) ||
			!strings.EqualFold(strings.TrimSpace(outcome.Currency), "GHS") {
			return ErrWrongAmount
		}
	}
	settled, err := service.payments.Settle(
		ctx, outcome.CallbackID, outcome.Reference, outcome.Reference, outcome.Success)
	if err != nil {
		return err
	}
	state := settled.State()
	if !outcome.Success {
		// A failure on a collection that never granted anything is nothing to
		// undo. Cancel is keyed by the intent, so this is safe to retry.
		if _, cancelErr := service.passes.Cancel(
			ctx, state.ID, outcome.CallbackID+":cancel",
		); cancelErr != nil {
			return nil
		}
		return nil
	}
	// What was bought, read back rather than assumed. Granting a pass named
	// after the payment was wrong twice: a payment is not a product, and the
	// membership context requires a slug where a payment id is hex — so most
	// payments would have taken the money and granted nothing at all.
	order, err := service.orders.Find(ctx, state.ID)
	if err != nil {
		return ErrUnavailable
	}
	paidThrough := service.now().UTC().Add(Period)
	if _, err := service.passes.Grant(
		ctx, state.MemberKey, order.SKUKey, state.ID, order.SKUVersion,
		paidThrough, Grace, outcome.CallbackID+":grant",
	); err != nil {
		return ErrUnavailable
	}
	if service.ledger != nil {
		// A failed posting does not un-grant a pass the member paid for. It is
		// a bookkeeping gap for reconciliation to find, which is what
		// reconciliation is for.
		_ = service.ledger.RecordSale(
			ctx, state.ID, int64(state.AmountPesewas), "GHS", service.now().UTC())
	}
	return nil
}

// sponsoredSeat draws the price from the issuing organization and grants the
// pass.
//
// The seat reference is the purchase command, so a retried purchase draws once
// and the organization is not charged twice for one member.
//
// A fund that cannot cover it is not an error: the sponsorship simply does not
// apply and the member buys their own membership. Blocking somebody because
// their employer's balance ran out would be the wrong way round.
func (service Service) sponsoredSeat(
	ctx context.Context, command StartCommand,
	sku catalogdomain.SKU, applied promotionapplication.Applied,
) (Started, error) {
	seatRef := "seat:" + strings.TrimSpace(command.CommandID)
	drawn, err := service.sponsors.Draw(
		ctx, applied.IssuerID, seatRef, sku.Price().Minor)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	if !drawn {
		return Started{}, ErrSponsorshipUnavailable
	}
	memberKey, err := service.keyer.MemberKey(command.MemberID)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	if err := service.orders.Record(ctx, Order{
		IntentID: seatRef, SKUKey: sku.SKUKey(), SKUVersion: sku.Version(),
		MemberID: strings.TrimSpace(command.MemberID), AmountPesewas: sku.Price().Minor,
		Code: applied.Code,
	}); err != nil {
		return Started{}, ErrUnavailable
	}
	paidThrough := service.now().UTC().Add(Period)
	if _, err := service.passes.Grant(
		ctx, memberKey, sku.SKUKey(), seatRef, sku.Version(),
		paidThrough, Grace, seatRef+":grant",
	); err != nil {
		return Started{}, ErrUnavailable
	}
	if service.ledger != nil {
		// Revenue, in full. The organization paid it; the platform earned it.
		// A sponsorship is not a discount and must not be booked as one.
		_ = service.ledger.RecordSale(
			ctx, seatRef, sku.Price().Minor, "GHS", service.now().UTC())
	}
	return Started{
		IntentID: seatRef, Status: "sponsored", AmountPesewas: 0,
		DiscountPesewas: uint64(sku.Price().Minor), Code: applied.Code,
	}, nil
}

func (service Service) ready() bool {
	return service.catalog != nil && service.payments != nil &&
		service.passes != nil && service.keyer != nil &&
		service.orders != nil && service.now != nil
}
