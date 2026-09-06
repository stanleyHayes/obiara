// Package systemauthority lets the platform itself post settlement.
//
// Money that arrives on a provider's callback has no operator standing behind
// it, and the admin authority requires an admin session. Without this, a
// member's payment could not be booked at all — which would leave revenue as
// something inferred from passes existing rather than written down.
//
// It is deliberately the narrowest thing that solves that: one actor, the
// settlement purposes only, and no balance reading whatsoever. Reading a
// balance is a person's act and stays behind the finance desk.
package systemauthority

import (
	"context"
	"errors"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

// Actor is the name the ledger records for a posting nobody made by hand.
const Actor = "system:settlement"

var (
	ErrNotPoster = errors.New("not permitted to post this")
	// ErrNotReader is unconditional. The system posts; it never reads a
	// balance, and a system that could would be a way around the finance
	// desk's own gate.
	ErrNotReader = errors.New("the system does not read balances")
)

type Authority struct{}

func New() Authority { return Authority{} }

// RequirePoster admits the system actor, for settlement only.
//
// A purpose check as well as an actor check, because the actor name is a
// constant in this binary: if it ever leaked into a request, the purpose
// bound is what still stops it being used to post anything at all.
func (Authority) RequirePoster(_ context.Context, actor string, purpose domain.Purpose) error {
	if strings.TrimSpace(actor) != Actor {
		return ErrNotPoster
	}
	if purpose != domain.PurposeSaleSettlement && purpose != domain.PurposeRefundSettlement {
		return ErrNotPoster
	}
	return nil
}

func (Authority) RequireBalanceReader(context.Context, string, string) error {
	return ErrNotReader
}
