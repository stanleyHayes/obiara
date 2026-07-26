package domain

import (
	"errors"
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

var ErrAccountNotUsable = errors.New("account is not active")

// Account binds one verified phone identity to one member (FR-102: exactly
// one active account per verified identity).
type Account struct {
	id        string
	phone     string
	status    AccountStatus
	createdAt time.Time
}

func NewAccount(id, phone string, now time.Time) (Account, error) {
	if id == "" {
		return Account{}, ErrSessionIDRequired
	}
	if !e164Pattern.MatchString(phone) {
		return Account{}, ErrInvalidPhone
	}
	return Account{id: id, phone: phone, status: AccountActive, createdAt: now.UTC()}, nil
}

// ReconstituteAccount rebuilds a stored account without policy checks.
func ReconstituteAccount(id, phone string, status AccountStatus, createdAt time.Time) Account {
	return Account{id: id, phone: phone, status: status, createdAt: createdAt}
}

func (account Account) Usable() error {
	if account.status != AccountActive {
		return ErrAccountNotUsable
	}
	return nil
}

func (account Account) ID() string            { return account.id }
func (account Account) Phone() string         { return account.phone }
func (account Account) Status() AccountStatus { return account.status }
func (account Account) CreatedAt() time.Time  { return account.createdAt }
