package domain

import (
	"errors"
	"strings"
	"time"
)

// AccountStatus is the account lifecycle state. Tier transitions are the
// S2-013 state machine; this task only establishes the account aggregate.
type AccountStatus string

const (
	AccountActive  AccountStatus = "active"
	AccountBlocked AccountStatus = "blocked"
	AccountDeleted AccountStatus = "deleted"
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
	id        string
	phone     string
	status    AccountStatus
	tier      Tier
	version   int64
	createdAt time.Time
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
func ReconstituteAccount(id, phone string, status AccountStatus, tier Tier, version int64, createdAt time.Time) Account {
	return Account{id: id, phone: phone, status: status, tier: tier, version: version, createdAt: createdAt}
}

func (account Account) Usable() error {
	if account.status != AccountActive {
		return ErrAccountNotUsable
	}
	return nil
}

// ApplyTransition moves the account along the tier ladder and returns the
// audit record. Promotion is exactly one step at a time; demotion may drop
// multiple steps but always requires a reason (e.g. verification reversal,
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

func (account Account) ID() string            { return account.id }
func (account Account) Phone() string         { return account.phone }
func (account Account) Status() AccountStatus { return account.status }
func (account Account) Tier() Tier            { return account.tier }
func (account Account) Version() int64        { return account.version }
func (account Account) CreatedAt() time.Time  { return account.createdAt }
