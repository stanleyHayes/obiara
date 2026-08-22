// Package adminauthority answers the ledger's authority questions from the
// composed admin context: who may post money, and who may read a balance.
package adminauthority

import (
	"context"
	"errors"
	"strings"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

var (
	// ErrNotPoster reports an operator who may not post to the ledger.
	ErrNotPoster = errors.New("operator may not post to the ledger")
	// ErrNotBalanceReader reports an operator who may not read a balance.
	ErrNotBalanceReader = errors.New("operator may not read this balance")
)

// Admin is the surface this bridge needs from the admin context.
type Admin interface {
	Authenticate(ctx context.Context, sessionID string) (admindomain.Session, admindomain.Principal, error)
}

// Authority gates the ledger on the finance and admin roles.
//
// A double-entry ledger is the record of what the platform owes and is owed,
// so both writing to it and reading balances out of it are finance-desk
// actions. Neither is widened to other desks: a verifier has no reason to
// see settlement positions, and least privilege is the point of having
// separate desks at all.
type Authority struct{ admin Admin }

func New(admin Admin) *Authority { return &Authority{admin: admin} }

func (authority *Authority) principal(ctx context.Context, sessionID string) (admindomain.Principal, bool) {
	_, principal, err := authority.admin.Authenticate(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return admindomain.Principal{}, false
	}
	return principal, principal.HasRole(admindomain.RoleFinance) || principal.HasRole(admindomain.RoleAdmin)
}

// RequirePoster gates writing a posting. The purpose is accepted but not
// narrowed further: every purpose in the ledger is a settlement movement,
// and splitting them across roles would imply a separation of duties the
// admin context does not yet model.
func (authority *Authority) RequirePoster(ctx context.Context, sessionID string, _ domain.Purpose) error {
	if _, ok := authority.principal(ctx, sessionID); !ok {
		return ErrNotPoster
	}
	return nil
}

// RequireBalanceReader gates reading an account balance.
func (authority *Authority) RequireBalanceReader(ctx context.Context, sessionID, _ string) error {
	if _, ok := authority.principal(ctx, sessionID); !ok {
		return ErrNotBalanceReader
	}
	return nil
}
