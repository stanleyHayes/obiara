// Package domain owns the discount code.
//
// Called a promotion and not a voucher, because `vouch/assisted` already owns
// that word for a person who vouches for another member's trustworthiness —
// which has nothing to do with money (agent_plan.md §41).
//
// A code is a bearer token by decision: whoever has it may use it until the
// cap is reached. That bounds the damage a leaked code can do rather than
// preventing it, which is the trade the owner chose over asking organizations
// to supply something verifiable.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

type Shape string
type Action string

const (
	// ShapePercentage takes a percentage off. ShapeFixed takes a fixed number
	// of minor units.
	ShapePercentage Shape = "percentage"
	ShapeFixed      Shape = "fixed"
	// ShapeSponsored is different in kind from the other two. They reduce
	// what a member is charged and the platform simply earns less. This one
	// means somebody else pays: the whole price is drawn from the issuing
	// organization's funded balance, and the revenue is earned in full.
	//
	// The distinction matters in the books. A discount is revenue forgone; a
	// sponsorship is revenue received from a different party, and the two
	// must not be posted the same way (agent_plan.md §78).
	ShapeSponsored Shape = "sponsored"

	ActionIssued    Action = "issued"
	ActionRedeemed  Action = "redeemed"
	ActionWithdrawn Action = "withdrawn"
)

var (
	ErrInvalidPromotion = errors.New("invalid promotion")
	ErrNotRedeemable    = errors.New("that code cannot be used")
	ErrAlreadyRedeemed  = errors.New("that code has already been used by this member")
	ErrExhausted        = errors.New("that code has been used up")
	ErrCommandMismatch  = errors.New("promotion command replay mismatch")
)

var (
	// codePattern is what a member types. Upper case and digits only, because
	// a code is read off a poster and typed on a phone: mixed case and
	// punctuation turn a discount into a support ticket.
	codePattern   = regexp.MustCompile(`^[A-Z0-9]{4,24}$`)
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	keyPattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// MaxPercentage is a hundred: a code may make a pass free, and nothing more.
// A percentage above it would owe a member money, which is not a discount.
const MaxPercentage = 100

type Command struct {
	ID string
	At time.Time
}

type Event struct {
	Sequence  uint64
	CommandID string
	Action    Action
	At        time.Time
}

type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}

// Promotion is the aggregate.
//
// Redemptions are keyed, so the row says how many members used a code and
// never which. Counting is the whole reporting story the owner chose: an
// organization sees "37 of your 50 are in use", never who.
type Promotion struct {
	id, code, issuerID string
	skuID              string
	shape              Shape
	amount             int64
	startsAt, endsAt   time.Time
	cap                uint32
	redeemedKeys       []string
	withdrawn          bool
	revision           uint64
	events             []Event
	commands           []AppliedCommand
}

type State struct {
	ID, Code, IssuerID, SKUID string
	Shape                     Shape
	Amount                    int64
	StartsAt, EndsAt          time.Time
	Cap                       uint32
	RedeemedKeys              []string
	Withdrawn                 bool
	Revision                  uint64
	Events                    []Event
	Commands                  []AppliedCommand
}

// Issue creates a code.
func Issue(
	id, code, issuerID, skuID string, shape Shape, amount int64,
	startsAt, endsAt time.Time, redemptionCap uint32, command Command,
) (Promotion, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !opaquePattern.MatchString(strings.TrimSpace(id)) ||
		!codePattern.MatchString(code) ||
		!opaquePattern.MatchString(strings.TrimSpace(issuerID)) ||
		!opaquePattern.MatchString(strings.TrimSpace(skuID)) ||
		!validShape(shape, amount) ||
		startsAt.IsZero() || endsAt.IsZero() || !endsAt.After(startsAt) ||
		redemptionCap == 0 ||
		!opaquePattern.MatchString(strings.TrimSpace(command.ID)) || command.At.IsZero() {
		return Promotion{}, ErrInvalidPromotion
	}
	promotion := Promotion{
		id: strings.TrimSpace(id), code: code, issuerID: strings.TrimSpace(issuerID),
		skuID: strings.TrimSpace(skuID), shape: shape, amount: amount,
		startsAt: startsAt.UTC(), endsAt: endsAt.UTC(), cap: redemptionCap,
	}
	return promotion.apply(command, ActionIssued), nil
}

func validShape(shape Shape, amount int64) bool {
	switch shape {
	case ShapePercentage:
		return amount > 0 && amount <= MaxPercentage
	case ShapeFixed:
		return amount > 0
	case ShapeSponsored:
		// A sponsorship covers the whole price, so an amount would be a
		// number nothing reads. Requiring it to be zero means a code cannot
		// be issued that looks like it means something it does not.
		return amount == 0
	default:
		return false
	}
}

// Redeemable reports whether this member may use the code now.
//
// It is separate from Redeem so a purchase can quote a discounted price
// without spending a redemption on a member who then abandons the payment.
func (promotion Promotion) Redeemable(memberKey string, at time.Time) error {
	if promotion.withdrawn {
		return ErrNotRedeemable
	}
	at = at.UTC()
	if at.Before(promotion.startsAt) || !at.Before(promotion.endsAt) {
		return ErrNotRedeemable
	}
	if uint32(len(promotion.redeemedKeys)) >= promotion.cap {
		return ErrExhausted
	}
	if !keyPattern.MatchString(memberKey) {
		return ErrNotRedeemable
	}
	for _, key := range promotion.redeemedKeys {
		if key == memberKey {
			// One per member, so a code is a welcome and not an income.
			return ErrAlreadyRedeemed
		}
	}
	return nil
}

// Redeem records that this member used the code.
func (promotion Promotion) Redeem(memberKey string, command Command) (Promotion, error) {
	if applied, found := promotion.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionRedeemed) {
			return Promotion{}, ErrCommandMismatch
		}
		// A retried redemption is the same redemption, not a second one.
		return promotion, nil
	}
	if err := promotion.Redeemable(memberKey, command.At); err != nil {
		return Promotion{}, err
	}
	if !opaquePattern.MatchString(strings.TrimSpace(command.ID)) || command.At.IsZero() {
		return Promotion{}, ErrInvalidPromotion
	}
	next := promotion
	next.redeemedKeys = append(append([]string(nil), promotion.redeemedKeys...), memberKey)
	return next.apply(command, ActionRedeemed), nil
}

// Withdraw stops a code being used, without touching what it already paid for.
func (promotion Promotion) Withdraw(command Command) (Promotion, error) {
	if applied, found := promotion.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionWithdrawn) {
			return Promotion{}, ErrCommandMismatch
		}
		return promotion, nil
	}
	if promotion.withdrawn {
		return Promotion{}, ErrNotRedeemable
	}
	next := promotion
	next.withdrawn = true
	return next.apply(command, ActionWithdrawn), nil
}

// Discount is what this code takes off a price, in minor units.
//
// Never more than the price: a discount larger than what is owed would make
// the platform owe the member, and the answer to "90% off a free pass" is
// zero rather than a refund. Rounding is toward the platform on purpose —
// a member is charged the whole pesewa, not a fraction nobody can pay.
func (promotion Promotion) Discount(priceMinor int64) int64 {
	if priceMinor <= 0 {
		return 0
	}
	var off int64
	switch promotion.shape {
	case ShapePercentage:
		off = priceMinor * promotion.amount / MaxPercentage
	case ShapeFixed:
		off = promotion.amount
	case ShapeSponsored:
		// The whole price, but it is not a discount: somebody else pays it.
		// The caller has to know which, which is what Sponsored reports.
		off = priceMinor
	}
	if off > priceMinor {
		return priceMinor
	}
	if off < 0 {
		return 0
	}
	return off
}

func (promotion Promotion) apply(command Command, action Action) Promotion {
	next := promotion
	next.revision = promotion.revision + 1
	next.events = append(append([]Event(nil), promotion.events...), Event{
		Sequence: next.revision, CommandID: strings.TrimSpace(command.ID),
		Action: action, At: command.At.UTC(),
	})
	next.commands = append(append([]AppliedCommand(nil), promotion.commands...), AppliedCommand{
		ID: strings.TrimSpace(command.ID), Fingerprint: fingerprint(command, action),
		Revision: next.revision,
	})
	return next
}

func (promotion Promotion) applied(commandID string) (AppliedCommand, bool) {
	for _, applied := range promotion.commands {
		if applied.ID == strings.TrimSpace(commandID) {
			return applied, true
		}
	}
	return AppliedCommand{}, false
}

func fingerprint(command Command, action Action) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(command.ID) + "\x00" + string(action)))
	return hex.EncodeToString(sum[:])
}

// Rehydrate rebuilds a promotion from storage.
func Rehydrate(state State) (Promotion, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(state.ID)) ||
		!codePattern.MatchString(state.Code) ||
		!validShape(state.Shape, state.Amount) ||
		state.Cap == 0 || state.Revision == 0 ||
		uint64(len(state.Events)) != state.Revision ||
		uint64(len(state.Commands)) != state.Revision {
		return Promotion{}, ErrInvalidPromotion
	}
	return Promotion{
		id: strings.TrimSpace(state.ID), code: state.Code, issuerID: state.IssuerID,
		skuID: state.SKUID, shape: state.Shape, amount: state.Amount,
		startsAt: state.StartsAt.UTC(), endsAt: state.EndsAt.UTC(), cap: state.Cap,
		redeemedKeys: append([]string(nil), state.RedeemedKeys...),
		withdrawn:    state.Withdrawn, revision: state.Revision,
		events:   append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}, nil
}

func (promotion Promotion) ID() string          { return promotion.id }
func (promotion Promotion) Code() string        { return promotion.code }
func (promotion Promotion) IssuerID() string    { return promotion.issuerID }
func (promotion Promotion) SKUID() string       { return promotion.skuID }
func (promotion Promotion) Shape() Shape        { return promotion.shape }
func (promotion Promotion) Amount() int64       { return promotion.amount }
func (promotion Promotion) Cap() uint32         { return promotion.cap }
func (promotion Promotion) Withdrawn() bool     { return promotion.withdrawn }
func (promotion Promotion) Revision() uint64    { return promotion.revision }
func (promotion Promotion) StartsAt() time.Time { return promotion.startsAt }
func (promotion Promotion) EndsAt() time.Time   { return promotion.endsAt }

// Redeemed is how many members have used it. The count is the whole reporting
// story: an organization sees how many of its codes are in use and never who
// used them.
func (promotion Promotion) Redeemed() uint32 { return uint32(len(promotion.redeemedKeys)) }

func (promotion Promotion) RedeemedKeys() []string {
	return append([]string(nil), promotion.redeemedKeys...)
}
func (promotion Promotion) Events() []Event {
	return append([]Event(nil), promotion.events...)
}
func (promotion Promotion) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), promotion.commands...)
}

// Sponsored reports whether this code means somebody else pays rather than
// that the member pays less.
//
// The two are the same number off the price and completely different money:
// one is revenue the platform never earns, the other is revenue it earns from
// an organization's funded balance. A caller that treated them alike would
// post a sponsorship as a discount and lose the deposit it drew down.
func (promotion Promotion) Sponsored() bool { return promotion.shape == ShapeSponsored }
