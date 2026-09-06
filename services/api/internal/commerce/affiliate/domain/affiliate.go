// Package domain owns the affiliate and what it has earned.
//
// An affiliate is an outside party — an organization, a creator, a campus rep
// — with a code and a relationship to Obiara. **Members are never affiliates**
// (agent_plan.md §41): paying somebody inside the community to recruit changes
// what "why is this person talking to me" means, in a product whose whole
// premise is that the answer is not money.
//
// Commission accrues on a qualified conversion and never on a signup. A scheme
// that paid per signup would reward exactly the bulk recruitment the tier
// ladder, age assurance and Sentinel exist to slow down, and the people best
// placed to exploit it are the ones the safety model can see least.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

type Status string
type Action string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"

	ActionRegistered Action = "registered"
	ActionSuspended  Action = "suspended"
	ActionRestored   Action = "restored"
	ActionAccrued    Action = "accrued"
	ActionClawedBack Action = "clawed_back"
	ActionPaidOut    Action = "paid_out"
)

var (
	ErrInvalidAffiliate  = errors.New("invalid affiliate")
	ErrInvalidTransition = errors.New("invalid affiliate transition")
	ErrCommandMismatch   = errors.New("affiliate command replay mismatch")
	// ErrAlreadyCounted refuses a second accrual for one referral. A referral
	// converts once; paying twice for it is the simplest way to farm a
	// scheme.
	ErrAlreadyCounted = errors.New("that referral has already been counted")
	// ErrNotCounted refuses a clawback for a referral that never accrued.
	ErrNotCounted = errors.New("that referral was never counted")
	// ErrInsufficientBalance refuses a payout larger than what is owed.
	ErrInsufficientBalance = errors.New("that is more than this affiliate has earned")
)

var (
	// codePattern is what a referred member types at signup. A code rather
	// than a tracking link: a code is honest, auditable, and does not require
	// following anybody around the internet to attribute them.
	codePattern   = regexp.MustCompile(`^[A-Z0-9]{4,24}$`)
	opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	keyPattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
	reasonPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{2,63}$`)
)

type Command struct {
	ID, ReasonCode string
	At             time.Time
}

type Event struct {
	Sequence      uint64
	CommandID     string
	ReasonCode    string
	Action        Action
	AmountPesewas int64
	At            time.Time
}

type AppliedCommand struct {
	ID, Fingerprint string
	Revision        uint64
}

// Affiliate is the aggregate.
//
// Referrals are keyed. An affiliate is told how many converted and never who:
// a statement that named the people somebody recruited would hand an outside
// party a list of members, which is the thing this whole codebase is built not
// to produce.
type Affiliate struct {
	id, name, code string
	contactEmail   string
	status         Status
	// countedKeys are the referrals that have accrued, so one cannot accrue
	// twice.
	countedKeys []string
	// accruedPesewas is everything ever earned; paidPesewas is everything
	// ever sent. What is owed is the difference, which is a liability rather
	// than a stored number that can drift from its own history.
	accruedPesewas int64
	paidPesewas    int64
	revision       uint64
	registeredAt   time.Time
	events         []Event
	commands       []AppliedCommand
}

type State struct {
	ID, Name, Code, ContactEmail string
	Status                       Status
	CountedKeys                  []string
	AccruedPesewas, PaidPesewas  int64
	Revision                     uint64
	RegisteredAt                 time.Time
	Events                       []Event
	Commands                     []AppliedCommand
}

// Register records an affiliate.
func Register(id, name, code, contactEmail string, command Command) (Affiliate, error) {
	name = strings.TrimSpace(name)
	code = strings.ToUpper(strings.TrimSpace(code))
	contactEmail = strings.ToLower(strings.TrimSpace(contactEmail))
	if !opaquePattern.MatchString(strings.TrimSpace(id)) || name == "" || len([]rune(name)) > 200 ||
		!codePattern.MatchString(code) || contactEmail == "" || len(contactEmail) > 254 ||
		!command.valid() {
		return Affiliate{}, ErrInvalidAffiliate
	}
	affiliate := Affiliate{
		id: strings.TrimSpace(id), name: name, code: code, contactEmail: contactEmail,
		status: StatusActive, registeredAt: command.At.UTC(),
	}
	return affiliate.apply(command, ActionRegistered, 0), nil
}

// Accrue records a qualified conversion and what it earned.
//
// referralKey is a digest of the referred member. It is kept so the same
// referral cannot be counted twice and for no other purpose — nothing reads it
// back out, and an affiliate statement is a count.
func (affiliate Affiliate) Accrue(
	referralKey string, amountPesewas int64, command Command,
) (Affiliate, error) {
	if applied, found := affiliate.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionAccrued, amountPesewas) {
			return Affiliate{}, ErrCommandMismatch
		}
		return affiliate, nil
	}
	if !command.valid() || !keyPattern.MatchString(referralKey) || amountPesewas <= 0 {
		return Affiliate{}, ErrInvalidAffiliate
	}
	if affiliate.status != StatusActive {
		return Affiliate{}, ErrInvalidTransition
	}
	if affiliate.counted(referralKey) {
		return Affiliate{}, ErrAlreadyCounted
	}
	next := affiliate
	next.countedKeys = append(append([]string(nil), affiliate.countedKeys...), referralKey)
	next.accruedPesewas = affiliate.accruedPesewas + amountPesewas
	return next.apply(command, ActionAccrued, amountPesewas), nil
}

// ClawBack reverses an accrual.
//
// A referral that was refunded, or a member removed for conduct, un-earns what
// it earned. This is what makes the qualification rule mean something after
// the fact rather than only at the moment it was checked.
//
// The referral stays counted. It converted once and was reversed once; letting
// it accrue again would make a clawback a way to double-count.
func (affiliate Affiliate) ClawBack(
	referralKey string, amountPesewas int64, command Command,
) (Affiliate, error) {
	if applied, found := affiliate.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionClawedBack, amountPesewas) {
			return Affiliate{}, ErrCommandMismatch
		}
		return affiliate, nil
	}
	if !command.valid() || !keyPattern.MatchString(referralKey) || amountPesewas <= 0 {
		return Affiliate{}, ErrInvalidAffiliate
	}
	if !affiliate.counted(referralKey) {
		return Affiliate{}, ErrNotCounted
	}
	next := affiliate
	// Accrued can go below what was paid: an affiliate paid for a conversion
	// that was later reversed owes it back. The balance is allowed to be
	// negative because pretending otherwise would silently forgive it.
	next.accruedPesewas = affiliate.accruedPesewas - amountPesewas
	return next.apply(command, ActionClawedBack, amountPesewas), nil
}

// RecordPayout records money actually sent.
func (affiliate Affiliate) RecordPayout(
	amountPesewas int64, command Command,
) (Affiliate, error) {
	if applied, found := affiliate.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, ActionPaidOut, amountPesewas) {
			return Affiliate{}, ErrCommandMismatch
		}
		return affiliate, nil
	}
	if !command.valid() || amountPesewas <= 0 {
		return Affiliate{}, ErrInvalidAffiliate
	}
	if amountPesewas > affiliate.Balance() {
		return Affiliate{}, ErrInsufficientBalance
	}
	next := affiliate
	next.paidPesewas = affiliate.paidPesewas + amountPesewas
	return next.apply(command, ActionPaidOut, amountPesewas), nil
}

// Suspend stops an affiliate earning. It does not touch what is already owed:
// work done is owed for, and whether to pay it is a decision somebody makes,
// not something a status change should silently settle.
func (affiliate Affiliate) Suspend(command Command) (Affiliate, error) {
	return affiliate.change(command, ActionSuspended, StatusActive, StatusSuspended)
}

func (affiliate Affiliate) Restore(command Command) (Affiliate, error) {
	return affiliate.change(command, ActionRestored, StatusSuspended, StatusActive)
}

func (affiliate Affiliate) change(
	command Command, action Action, from, to Status,
) (Affiliate, error) {
	if applied, found := affiliate.applied(command.ID); found {
		if applied.Fingerprint != fingerprint(command, action, 0) {
			return Affiliate{}, ErrCommandMismatch
		}
		return affiliate, nil
	}
	if !command.valid() {
		return Affiliate{}, ErrInvalidAffiliate
	}
	if affiliate.status != from {
		return Affiliate{}, ErrInvalidTransition
	}
	next := affiliate
	next.status = to
	return next.apply(command, action, 0), nil
}

func (affiliate Affiliate) apply(
	command Command, action Action, amountPesewas int64,
) Affiliate {
	next := affiliate
	next.revision = affiliate.revision + 1
	next.events = append(append([]Event(nil), affiliate.events...), Event{
		Sequence: next.revision, CommandID: strings.TrimSpace(command.ID),
		ReasonCode: strings.TrimSpace(command.ReasonCode), Action: action,
		AmountPesewas: amountPesewas, At: command.At.UTC(),
	})
	next.commands = append(append([]AppliedCommand(nil), affiliate.commands...), AppliedCommand{
		ID: strings.TrimSpace(command.ID), Revision: next.revision,
		Fingerprint: fingerprint(command, action, amountPesewas),
	})
	return next
}

func (affiliate Affiliate) applied(commandID string) (AppliedCommand, bool) {
	for _, applied := range affiliate.commands {
		if applied.ID == strings.TrimSpace(commandID) {
			return applied, true
		}
	}
	return AppliedCommand{}, false
}

func (affiliate Affiliate) counted(referralKey string) bool {
	for _, key := range affiliate.countedKeys {
		if key == referralKey {
			return true
		}
	}
	return false
}

func (command Command) valid() bool {
	return opaquePattern.MatchString(strings.TrimSpace(command.ID)) &&
		reasonPattern.MatchString(strings.TrimSpace(command.ReasonCode)) &&
		!command.At.IsZero()
}

func fingerprint(command Command, action Action, amountPesewas int64) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		strings.TrimSpace(command.ID), string(action),
		strconvInt(amountPesewas),
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func strconvInt(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := make([]byte, 0, 20)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

// Rehydrate rebuilds an affiliate from storage.
func Rehydrate(state State) (Affiliate, error) {
	if !opaquePattern.MatchString(strings.TrimSpace(state.ID)) ||
		!codePattern.MatchString(state.Code) ||
		(state.Status != StatusActive && state.Status != StatusSuspended) ||
		state.Revision == 0 ||
		uint64(len(state.Events)) != state.Revision ||
		uint64(len(state.Commands)) != state.Revision {
		return Affiliate{}, ErrInvalidAffiliate
	}
	return Affiliate{
		id: strings.TrimSpace(state.ID), name: state.Name, code: state.Code,
		contactEmail: state.ContactEmail, status: state.Status,
		countedKeys:    append([]string(nil), state.CountedKeys...),
		accruedPesewas: state.AccruedPesewas, paidPesewas: state.PaidPesewas,
		revision: state.Revision, registeredAt: state.RegisteredAt.UTC(),
		events:   append([]Event(nil), state.Events...),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}, nil
}

func (affiliate Affiliate) ID() string              { return affiliate.id }
func (affiliate Affiliate) Name() string            { return affiliate.name }
func (affiliate Affiliate) Code() string            { return affiliate.code }
func (affiliate Affiliate) ContactEmail() string    { return affiliate.contactEmail }
func (affiliate Affiliate) Status() Status          { return affiliate.status }
func (affiliate Affiliate) Revision() uint64        { return affiliate.revision }
func (affiliate Affiliate) Earning() bool           { return affiliate.status == StatusActive }
func (affiliate Affiliate) AccruedPesewas() int64   { return affiliate.accruedPesewas }
func (affiliate Affiliate) PaidPesewas() int64      { return affiliate.paidPesewas }
func (affiliate Affiliate) RegisteredAt() time.Time { return affiliate.registeredAt }

// Balance is what is owed: everything earned less everything sent. Derived
// rather than stored, so it cannot drift from the history that produced it.
func (affiliate Affiliate) Balance() int64 {
	return affiliate.accruedPesewas - affiliate.paidPesewas
}

// Conversions is how many referrals have counted. This is the whole of what an
// affiliate is told about who they brought.
func (affiliate Affiliate) Conversions() int { return len(affiliate.countedKeys) }

func (affiliate Affiliate) CountedKeys() []string {
	return append([]string(nil), affiliate.countedKeys...)
}
func (affiliate Affiliate) Events() []Event {
	return append([]Event(nil), affiliate.events...)
}
func (affiliate Affiliate) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), affiliate.commands...)
}
