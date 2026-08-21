// Package domain models admin principals and their roles (E16-S01;
// FR-801: admin actions are least-privilege, MFA-gated and immutably
// audited). Member auth never mixes with admin auth.
package domain

import (
	"errors"
	"net/mail"
	"strings"
	"time"
)

// Role is an admin capability assignment, mirroring the authz kernel's
// role vocabulary (services/api/internal/authz).
type Role string

const (
	RoleVerifier Role = "verifier"
	RoleTSAgent  Role = "ts_agent"
	RoleHost     Role = "host"
	RoleFinance  Role = "finance"
	RoleAdmin    Role = "admin"
)

var (
	ErrInvalidEmail      = errors.New("admin principal email is invalid")
	ErrInvalidRole       = errors.New("unknown admin role")
	ErrNoRoles           = errors.New("admin principal needs at least one role")
	ErrPrincipalIDNeeded = errors.New("admin principal id is required")
	ErrInvalidStatus     = errors.New("unknown admin principal status")
)

// Principal is one admin user.
type Principal struct {
	id           string
	email        string
	roles        []Role
	status       Status
	passwordHash string
	version      int64
	createdAt    time.Time
}

// Status is the principal lifecycle.
type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

// NewPrincipal validates enrollment input.
func NewPrincipal(id, email string, roles []Role, now time.Time) (Principal, error) {
	if strings.TrimSpace(id) == "" {
		return Principal{}, ErrPrincipalIDNeeded
	}
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil || address.Address != strings.TrimSpace(email) {
		return Principal{}, ErrInvalidEmail
	}
	if err := validateRoles(roles); err != nil {
		return Principal{}, err
	}
	return Principal{
		id:        id,
		email:     address.Address,
		roles:     roles,
		status:    StatusActive,
		version:   1,
		createdAt: now.UTC(),
	}, nil
}

// ReconstitutePrincipal rebuilds a stored principal without checks.
func ReconstitutePrincipal(id, email string, roles []Role, status Status, version int64, createdAt time.Time) Principal {
	return Principal{id: id, email: email, roles: roles, status: status, version: version, createdAt: createdAt}
}

// ReconstitutePrincipalWithPassword rebuilds a stored principal including its
// password digest. Principals enrolled before password support have an empty
// digest and keep authenticating on the emailed code alone.
func ReconstitutePrincipalWithPassword(id, email string, roles []Role, status Status, passwordHash string, version int64, createdAt time.Time) Principal {
	principal := ReconstitutePrincipal(id, email, roles, status, version, createdAt)
	principal.passwordHash = passwordHash
	return principal
}

// SetPassword validates and stores a new password digest.
func (principal *Principal) SetPassword(plain string) error {
	hash, err := HashPassword(plain)
	if err != nil {
		return err
	}
	principal.passwordHash = hash
	principal.version++
	return nil
}

// HasPassword reports whether this principal must present a password before
// an MFA code is minted.
func (principal Principal) HasPassword() bool { return principal.passwordHash != "" }

// VerifyPassword checks a presented password against the stored digest.
// It returns false when no password is set, so callers cannot accidentally
// treat "no password configured" as "any password accepted".
func (principal Principal) VerifyPassword(plain string) bool {
	if principal.passwordHash == "" {
		return false
	}
	return VerifyPassword(principal.passwordHash, plain)
}

// PasswordHash exposes the stored digest for persistence only.
func (principal Principal) PasswordHash() string { return principal.passwordHash }

func (principal Principal) HasRole(role Role) bool {
	for _, assigned := range principal.roles {
		if assigned == role {
			return true
		}
	}
	return false
}

func (principal *Principal) Suspend() {
	principal.status = StatusSuspended
	principal.version++
}

// Reactivate restores an operator after a reviewed suspension.
func (principal *Principal) Reactivate() {
	principal.status = StatusActive
	principal.version++
}

// ReplaceRoles applies a complete, validated least-privilege role set.
func (principal *Principal) ReplaceRoles(roles []Role) error {
	if err := validateRoles(roles); err != nil {
		return err
	}
	principal.roles = append([]Role(nil), roles...)
	principal.version++
	return nil
}

func validateRoles(roles []Role) error {
	if len(roles) == 0 {
		return ErrNoRoles
	}
	seen := map[Role]bool{}
	for _, role := range roles {
		switch role {
		case RoleVerifier, RoleTSAgent, RoleHost, RoleFinance, RoleAdmin:
		default:
			return ErrInvalidRole
		}
		if seen[role] {
			return ErrInvalidRole
		}
		seen[role] = true
	}
	return nil
}

func (principal Principal) ID() string           { return principal.id }
func (principal Principal) Email() string        { return principal.email }
func (principal Principal) Roles() []Role        { return principal.roles }
func (principal Principal) Status() Status       { return principal.status }
func (principal Principal) Version() int64       { return principal.version }
func (principal Principal) CreatedAt() time.Time { return principal.createdAt }
