package purchase

import (
	"context"
	"errors"
	"strings"
)

// Order is what a collection was opened to buy.
//
// It exists because settlement has to know what was bought and the payment
// context deliberately does not: a collection intent is product-agnostic by
// design and stores no catalogue state at all. Without this, granting a pass
// on a webhook was guessing — and it guessed the payment's own id, which is
// neither the product nor a shape the membership context accepts
// (agent_plan.md §75).
type Order struct {
	// IntentID is the collection this order belongs to, which is also the
	// reference the processor was given.
	IntentID string
	// SKUKey and SKUVersion name the product. They are what a membership pass
	// is granted for: the key is a slug like membership.monthly, and the
	// version pins which priced revision of it was actually sold.
	SKUKey     string
	SKUVersion uint64
	// MemberID is raw, because a refund or a support question starts from a
	// person and the digests elsewhere cannot be reversed. It is the one
	// place a purchase is legible, and it is the operator's record rather
	// than anything a member surface reads.
	MemberID string
	// AmountPesewas is what was actually asked for, after any discount. The
	// intent holds it too; keeping it here means a statement can be read
	// without joining to the payment context.
	AmountPesewas int64
	// Code is the discount code that applied, empty when none did.
	Code string
}

var ErrOrderNotFound = errors.New("no order for that collection")

// Orders remembers what each collection was opened to buy.
type Orders interface {
	Record(context.Context, Order) error
	Find(ctx context.Context, intentID string) (Order, error)
}

func (order Order) valid() bool {
	return strings.TrimSpace(order.IntentID) != "" &&
		strings.TrimSpace(order.SKUKey) != "" &&
		order.SKUVersion > 0 &&
		strings.TrimSpace(order.MemberID) != "" &&
		order.AmountPesewas > 0
}
