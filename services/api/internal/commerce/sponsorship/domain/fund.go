// Package domain owns an organization's funded balance.
//
// A fund is money an organization has already paid, held against seats its
// members have not taken yet. It is a **liability** until a seat is taken: the
// platform is holding somebody else's money, and calling it revenue before a
// seat is drawn would book income for something not yet delivered.
//
// Deposits are recorded by an operator rather than collected by a rail. The
// same decision as the codes themselves (agent_plan.md §41): an organization
// pays by whatever means it and Obiara agreed — a bank transfer, an invoice
// settled offline, a Paystack link — and somebody records what arrived. A B2B
// collection rail is a product this does not have and does not need in order
// for a university to sponsor fifty seats.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

type Action string

const (
	ActionOpened    Action = "opened"
	ActionDeposited Action = "deposited"
	ActionDrawn     Action = "drawn"
	ActionRefunded  Action = "refunded"
	ActionClosed    Action = "closed"
)

// MaxMovementPesewas bounds one deposit or one draw.
//
// A hundred million cedis, which no real movement approaches. It is not an
// overflow guard — reaching int64 would take nine quintillion pesewas and is
// fantasy — it is a guard against the realistic failure, which is an operator
// typing an extra six zeros into a deposit and an organization's balance
// becoming a number nobody can explain.
const MaxMovementPesewas int64 = 10_000_000_000

var (
	ErrInvalidFund = errors.New("invalid sponsorship fund")
	// ErrAmountOutOfRange refuses a movement larger than any real one, which
	// in practice means a mistyped deposit.
	ErrAmountOutOfRange = errors.New("that amount is larger than any real movement")
	ErrFundClosed       = errors.New("that fund is closed")
	ErrInsufficient     = errors.New("that fund does not hold enough")
	ErrCommandMismatch  = errors.New("sponsorship command replay mismatch")
	// ErrNotDrawn refuses a refund for a seat that was never drawn.
	ErrNotDrawn = errors.New("that seat was never drawn from this fund")
)

var (
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	actorPattern  = regexp.MustCompile(`^[a-f0-9]{64}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type Command struct {
	ID, ActorKey, ReasonCode string
	At                       time.Time
}

type Event struct {
	Sequence                        uint64
	CommandID, ActorKey, ReasonCode string
	Action                          Action
	AmountPesewas                   int64
	At                              time.Time
}

type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}

// Fund is the aggregate.
//
// Deposited and drawn are both kept, and the balance is the difference. A
// stored balance would be one number that can drift from the history that
// produced it, and this history is what an organization is shown when it asks
// where its money went.
type Fund struct {
	id, organizationID string
	depositedPesewas   int64
	drawnPesewas       int64
	// drawnSeats are the collections this fund has paid for, keyed by the
	// purchase they belong to, so one seat cannot be drawn twice and a refund
	// can find what it is reversing.
	drawnSeats map[string]int64
	closed     bool
	revision   uint64
	openedAt   time.Time
	events     []Event
	commands   []AppliedCommand
}

type State struct {
	ID, OrganizationID string
	DepositedPesewas   int64
	DrawnPesewas       int64
	DrawnSeats         map[string]int64
	Closed             bool
	Revision           uint64
	OpenedAt           time.Time
	Events             []Event
	Commands           []AppliedCommand
}

// Open starts a fund for an organization, holding nothing.
func Open(id, organizationID string, command Command) (Fund, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(id)) ||
		!opaquePattern.MatchString(strings.TrimSpace(organizationID)) ||
		!command.valid() {
		return Fund{}, ErrInvalidFund
	}
	fund := Fund{
		id: strings.TrimSpace(id), organizationID: strings.TrimSpace(organizationID),
		drawnSeats: map[string]int64{}, openedAt: command.At.UTC(),
	}
	return fund.apply(command, ActionOpened, 0, ""), nil
}

// Deposit records money an organization has paid.
//
// The operator records what arrived; nothing here collects it. What makes this
// safe is the same thing that makes every other operator act safe: it is
// idempotent by command id, it names who did it, and it says why.
func (fund Fund) Deposit(amountPesewas int64, command Command) (Fund, error) {
	if applied, found := fund.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionDeposited, pesewas(amountPesewas)) {
			return Fund{}, ErrCommandMismatch
		}
		return fund, nil
	}
	if !command.valid() || amountPesewas <= 0 {
		return Fund{}, ErrInvalidFund
	}
	if amountPesewas > MaxMovementPesewas {
		return Fund{}, ErrAmountOutOfRange
	}
	if fund.closed {
		return Fund{}, ErrFundClosed
	}
	next := fund
	next.depositedPesewas = fund.depositedPesewas + amountPesewas
	return next.apply(command, ActionDeposited, amountPesewas, ""), nil
}

// Draw takes the price of one seat.
//
// seatRef is the purchase this seat belongs to. It is what stops one
// collection drawing twice, and what a refund reverses against.
func (fund Fund) Draw(seatRef string, amountPesewas int64, command Command) (Fund, error) {
	if applied, found := fund.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionDrawn, pesewas(amountPesewas)) {
			return Fund{}, ErrCommandMismatch
		}
		return fund, nil
	}
	if !command.valid() || amountPesewas <= 0 ||
		!opaquePattern.MatchString(strings.TrimSpace(seatRef)) {
		return Fund{}, ErrInvalidFund
	}
	if amountPesewas > MaxMovementPesewas {
		return Fund{}, ErrAmountOutOfRange
	}
	if fund.closed {
		return Fund{}, ErrFundClosed
	}
	if _, drawn := fund.drawnSeats[strings.TrimSpace(seatRef)]; drawn {
		// Already paid for. Drawing again would charge an organization twice
		// for one member's seat.
		return fund, nil
	}
	if amountPesewas > fund.Balance() {
		return Fund{}, ErrInsufficient
	}
	next := fund
	next.drawnPesewas = fund.drawnPesewas + amountPesewas
	next.drawnSeats = copySeats(fund.drawnSeats)
	next.drawnSeats[strings.TrimSpace(seatRef)] = amountPesewas
	return next.apply(command, ActionDrawn, amountPesewas, strings.TrimSpace(seatRef)), nil
}

// Refund returns a drawn seat to the balance.
//
// For a membership that was refunded or reversed: the organization did not get
// what it paid for, so it gets the money back rather than the platform keeping
// it. The seat stays recorded so the same one cannot be refunded twice.
func (fund Fund) Refund(seatRef string, command Command) (Fund, error) {
	if applied, found := fund.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionRefunded, strings.TrimSpace(seatRef)) {
			return Fund{}, ErrCommandMismatch
		}
		return fund, nil
	}
	if !command.valid() {
		return Fund{}, ErrInvalidFund
	}
	amount, drawn := fund.drawnSeats[strings.TrimSpace(seatRef)]
	if !drawn || amount <= 0 {
		return Fund{}, ErrNotDrawn
	}
	next := fund
	next.drawnPesewas = fund.drawnPesewas - amount
	next.drawnSeats = copySeats(fund.drawnSeats)
	// Zeroed rather than deleted, so the seat is still known to have been
	// drawn once and cannot be refunded a second time.
	next.drawnSeats[strings.TrimSpace(seatRef)] = 0
	return next.apply(command, ActionRefunded, amount, strings.TrimSpace(seatRef)), nil
}

// Close stops a fund being drawn from. It does not move what is left: money an
// organization has paid and not used is still theirs, and returning it is a
// decision somebody makes with an invoice in front of them.
func (fund Fund) Close(command Command) (Fund, error) {
	if applied, found := fund.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionClosed, "") {
			return Fund{}, ErrCommandMismatch
		}
		return fund, nil
	}
	if !command.valid() {
		return Fund{}, ErrInvalidFund
	}
	if fund.closed {
		return Fund{}, ErrFundClosed
	}
	next := fund
	next.closed = true
	return next.apply(command, ActionClosed, 0, ""), nil
}

func (fund Fund) apply(command Command, action Action, amountPesewas int64, seatRef string) Fund {
	next := fund
	next.revision = fund.revision + 1
	next.events = append(append([]Event(nil), fund.events...), Event{
		Sequence: next.revision, CommandID: strings.TrimSpace(command.ID),
		ActorKey: command.ActorKey, ReasonCode: strings.TrimSpace(command.ReasonCode),
		Action: action, AmountPesewas: amountPesewas, At: command.At.UTC(),
	})
	discriminator := pesewas(amountPesewas)
	switch action {
	case ActionRefunded:
		discriminator = seatRef
	case ActionClosed, ActionOpened:
		discriminator = ""
	}
	next.commands = append(append([]AppliedCommand(nil), fund.commands...), AppliedCommand{
		ID: strings.TrimSpace(command.ID), Revision: next.revision,
		Fingerprint: fingerprint(command, action, discriminator),
	})
	return next
}

func (fund Fund) applied(commandID string) (AppliedCommand, bool) {
	for _, applied := range fund.commands {
		if applied.ID == strings.TrimSpace(commandID) {
			return applied, true
		}
	}
	return AppliedCommand{}, false
}

func copySeats(seats map[string]int64) map[string]int64 {
	copied := make(map[string]int64, len(seats)+1)
	for reference, amount := range seats {
		copied[reference] = amount
	}
	return copied
}

func (command Command) valid() bool {
	return opaquePattern.MatchString(strings.TrimSpace(command.ID)) &&
		actorPattern.MatchString(command.ActorKey) &&
		reasonPattern.MatchString(strings.TrimSpace(command.ReasonCode)) &&
		!command.At.IsZero()
}

// fingerprint binds a command id to what it did, so a retry carrying the same
// id and different intent is refused rather than answered with the first one's
// outcome.
//
// The discriminator is what the caller actually supplied: the amount for a
// deposit or a draw, the seat for a refund. A refund's amount is derived from
// what was drawn rather than given, so binding it would compare a number the
// caller never sent.
func fingerprint(command Command, action Action, discriminator string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(command.ID), command.ActorKey, string(action), discriminator,
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func pesewas(value int64) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 20)
	negative := value < 0
	if negative {
		value = -value
	}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// Rehydrate rebuilds a fund from storage.
func Rehydrate(state State) (Fund, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(state.ID)) ||
		state.Revision == 0 ||
		uint64(len(state.Events)) != state.Revision ||
		uint64(len(state.Commands)) != state.Revision ||
		state.DepositedPesewas < 0 || state.DrawnPesewas < 0 {
		return Fund{}, ErrInvalidFund
	}
	seats := state.DrawnSeats
	if seats == nil {
		seats = map[string]int64{}
	}
	return Fund{
		id: strings.TrimSpace(state.ID), organizationID: state.OrganizationID,
		depositedPesewas: state.DepositedPesewas, drawnPesewas: state.DrawnPesewas,
		drawnSeats: copySeats(seats), closed: state.Closed,
		revision: state.Revision, openedAt: state.OpenedAt.UTC(),
		events:   append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}, nil
}

func (fund Fund) ID() string              { return fund.id }
func (fund Fund) OrganizationID() string  { return fund.organizationID }
func (fund Fund) DepositedPesewas() int64 { return fund.depositedPesewas }
func (fund Fund) DrawnPesewas() int64     { return fund.drawnPesewas }
func (fund Fund) Closed() bool            { return fund.closed }
func (fund Fund) Revision() uint64        { return fund.revision }
func (fund Fund) OpenedAt() time.Time     { return fund.openedAt }

// Balance is what is left to spend: deposited less drawn. Derived rather than
// stored, so it cannot drift from the history an organization is shown.
func (fund Fund) Balance() int64 { return fund.depositedPesewas - fund.drawnPesewas }

// Seats is how many have been taken. A count, which is the whole of what an
// organization learns about who used them.
func (fund Fund) Seats() int {
	taken := 0
	for _, amount := range fund.drawnSeats {
		if amount > 0 {
			taken++
		}
	}
	return taken
}

func (fund Fund) DrawnSeats() map[string]int64 { return copySeats(fund.drawnSeats) }
func (fund Fund) Events() []Event {
	return append([]Event(nil), fund.events...)
}
func (fund Fund) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), fund.commands...)
}
