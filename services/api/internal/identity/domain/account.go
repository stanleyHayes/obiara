package domain

import (
	"errors"
	"strings"
	"time"
)

// AccountStatus is the account lifecycle state. Tier transitions are the
// S2-013 state machine; enforcement status transitions (suspend, block,
// reactivate) are the E12-S04 action-ladder effects.
type AccountStatus string

const (
	AccountActive    AccountStatus = "active"
	AccountSuspended AccountStatus = "suspended"
	AccountBlocked   AccountStatus = "blocked"
	AccountDeleted   AccountStatus = "deleted"
)

// Tier is the verification ladder (FR-101): Tier 0 unverified, Tier 1
// verified (romantic surfaces), Tier 2 sowing-eligible. The numeric values
// mirror the authorization kernel's Tier in services/api/internal/authz.
type Tier int

const (
	TierUnverified Tier = 0
	TierVerified   Tier = 1
	TierSowing     Tier = 2
)

var (
	ErrAccountNotUsable         = errors.New("account is not active")
	ErrInvalidTierTransition    = errors.New("invalid tier transition")
	ErrTransitionReasonRequired = errors.New("tier transition reason is required")
	ErrTransitionActorRequired  = errors.New("tier transition actor is required")
	ErrSuspensionUntilRequired  = errors.New("suspension end time is required")
	ErrNotSuspended             = errors.New("account is not suspended")
)

// TierTransition is the immutable audit record of one tier change
// (E03-S06: transitions are audited with actor and reason).
type TierTransition struct {
	AccountID  string
	From       Tier
	To         Tier
	Reason     string
	ActorID    string
	OccurredAt time.Time
}

// Account binds one verified phone identity to one member (FR-102: exactly
// one active account per verified identity).
type Account struct {
	id             string
	phone          string
	status         AccountStatus
	tier           Tier
	version        int64
	suspendedUntil *time.Time
	createdAt      time.Time
}

func NewAccount(id, phone string, now time.Time) (Account, error) {
	if id == "" {
		return Account{}, ErrSessionIDRequired
	}
	if !e164Pattern.MatchString(phone) {
		return Account{}, ErrInvalidPhone
	}
	return Account{id: id, phone: phone, status: AccountActive, tier: TierUnverified, version: 1, createdAt: now.UTC()}, nil
}

// ReconstituteAccount rebuilds a stored account without policy checks.
func ReconstituteAccount(id, phone string, status AccountStatus, tier Tier, version int64, suspendedUntil *time.Time, createdAt time.Time) Account {
	return Account{id: id, phone: phone, status: status, tier: tier, version: version, suspendedUntil: suspendedUntil, createdAt: createdAt}
}

// Usable reports whether the account may act right now. A suspended
// account becomes usable only through Reactivate; blocked and deleted are
// terminal for product surfaces.
func (account Account) Usable() error {
	if account.status != AccountActive {
		return ErrAccountNotUsable
	}
	return nil
}

// Suspend places the account under a timed suspension (Tier-B ladder
// effect, Doc 09 §2: 14-90 days). Sessions are revoked upstream.
func (account *Account) Suspend(until time.Time) error {
	if account.status != AccountActive {
		return ErrAccountNotUsable
	}
	if until.IsZero() {
		return ErrSuspensionUntilRequired
	}
	untilUTC := until.UTC()
	account.status = AccountSuspended
	account.suspendedUntil = &untilUTC
	account.version++
	return nil
}

// Block ends the account's product access (Tier-A ladder effect).
func (account *Account) Block() {
	account.status = AccountBlocked
	account.suspendedUntil = nil
	account.version++
}

// Reactivate restores an expired suspension. Calling it early is an error
// so operators cannot silently lift a running suspension.
func (account *Account) Reactivate(now time.Time) error {
	if account.status != AccountSuspended {
		return ErrNotSuspended
	}
	if account.suspendedUntil != nil && now.UTC().Before(*account.suspendedUntil) {
		return ErrSuspensionUntilRequired
	}
	account.status = AccountActive
	account.suspendedUntil = nil
	account.version++
	return nil
}

// ApplyTransition moves the account along the tier ladder and returns the
// audit record. Promotion is exactly one step at a time; demotion may drop
// multiple steps but always requires a reason (verification reversal,
// safety action). Every transition carries actor metadata (agent_plan.md
// §7.4) and increments the optimistic-concurrency version.
func (account *Account) ApplyTransition(target Tier, reason, actorID string, now time.Time) (TierTransition, error) {
	if strings.TrimSpace(reason) == "" {
		return TierTransition{}, ErrTransitionReasonRequired
	}
	if strings.TrimSpace(actorID) == "" {
		return TierTransition{}, ErrTransitionActorRequired
	}
	if err := account.Usable(); err != nil {
		return TierTransition{}, err
	}

	from := account.tier
	promoteOneStep := target == from+1
	demote := target < from
	switch {
	case target < TierUnverified || target > TierSowing:
		return TierTransition{}, ErrInvalidTierTransition
	case !promoteOneStep && !demote:
		return TierTransition{}, ErrInvalidTierTransition
	}

	account.tier = target
	account.version++
	return TierTransition{
		AccountID:  account.id,
		From:       from,
		To:         target,
		Reason:     reason,
		ActorID:    actorID,
		OccurredAt: now.UTC(),
	}, nil
}

func (account Account) ID() string                 { return account.id }
func (account Account) Phone() string              { return account.phone }
func (account Account) Status() AccountStatus      { return account.status }
func (account Account) Tier() Tier                 { return account.tier }
func (account Account) Version() int64             { return account.version }
func (account Account) SuspendedUntil() *time.Time { return account.suspendedUntil }
func (account Account) CreatedAt() time.Time       { return account.createdAt }
