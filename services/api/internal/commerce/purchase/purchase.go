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
	Confirm(ctx context.Context, id, command string) (momodomain.Intent, error)
	Callback(ctx context.Context, callback momoapplication.Callback) (momodomain.Intent, error)
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

// Ledger records the money. Purchases post through the same double-entry book
// as everything else, so revenue is visible rather than implied by a pass
// appearing.
type Ledger interface {
	RecordSale(ctx context.Context, reference string, minor int64, currency string, at time.Time) error
}

type Service struct {
	catalog  Catalog
	payments Payments
	passes   Passes
	keyer    Keyer
	ledger   Ledger
	now      func() time.Time
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
}

type Started struct {
	IntentID string
	Status   string
	// AmountPesewas is what the member is about to be asked for, echoed back
	// so a client can show it rather than guess at it.
	AmountPesewas uint64
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
		strings.TrimSpace(command.Phone) == "" {
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
	memberKey, err := service.keyer.MemberKey(command.MemberID)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	phoneRef, err := service.keyer.PhoneRef(command.Phone)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	intent, err := service.payments.Create(
		ctx, memberKey, phoneRef, uint64(sku.Price().Minor), command.CommandID)
	if err != nil {
		return Started{}, ErrUnavailable
	}
	// Confirmed in the same request because the member is standing there: the
	// deliberate gesture is the purchase itself, and a second round trip only
	// adds a place for it to be abandoned.
	confirmed, err := service.payments.Confirm(ctx, intent.State().ID, command.CommandID+":confirm")
	if err != nil {
		return Started{}, ErrUnavailable
	}
	return Started{
		IntentID:      confirmed.State().ID,
		Status:        string(confirmed.State().Status),
		AmountPesewas: uint64(sku.Price().Minor),
	}, nil
}

// Settle applies the provider's callback.
//
// On success the pass is granted; on a reversal an already-granted pass is
// cancelled rather than removed, so the trail shows granted-then-cancelled
// instead of a pass that quietly vanished.
func (service Service) Settle(
	ctx context.Context, callback momoapplication.Callback,
) error {
	if !service.ready() {
		return ErrUnavailable
	}
	intent, err := service.payments.Callback(ctx, callback)
	if err != nil {
		return err
	}
	state := intent.State()
	if !callback.Success {
		// A reversal on an intent that never granted anything is nothing to
		// undo. Cancel is keyed by the intent, so this is safe to retry.
		if _, cancelErr := service.passes.Cancel(
			ctx, state.ID, callback.CallbackID+":cancel",
		); cancelErr != nil {
			// A pass that was never granted cannot be cancelled, which is the
			// ordinary case for a payment that simply failed.
			return nil
		}
		return nil
	}
	paidThrough := service.now().UTC().Add(Period)
	if _, err := service.passes.Grant(
		ctx, state.MemberKey, state.ID, state.ID, 1, paidThrough, Grace,
		callback.CallbackID+":grant",
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

func (service Service) ready() bool {
	return service.catalog != nil && service.payments != nil &&
		service.passes != nil && service.keyer != nil && service.now != nil
}
