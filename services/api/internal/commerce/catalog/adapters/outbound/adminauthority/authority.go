// Package adminauthority answers the catalog's authority question from the
// composed admin context: may this operator curate what members can buy?
package adminauthority

import (
	"context"
	"errors"
	"strings"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
)

// ErrNotCatalogEditor reports an operator who may not curate the catalog.
var ErrNotCatalogEditor = errors.New("operator may not edit the catalog")

// Admin is the surface this bridge needs from the admin context.
type Admin interface {
	Authenticate(ctx context.Context, sessionID string) (admindomain.Session, admindomain.Principal, error)
}

// Authority permits operators holding the finance or admin role.
//
// The catalog decides what members can be charged for, so it is gated on the
// two roles that already carry commercial responsibility rather than on a
// new one invented here. Roles come from the authenticated principal, never
// from the caller.
type Authority struct{ admin Admin }

func New(admin Admin) *Authority { return &Authority{admin: admin} }

// RequireCatalogEditor takes the operator's admin session id.
func (authority *Authority) RequireCatalogEditor(ctx context.Context, sessionID string) error {
	_, principal, err := authority.admin.Authenticate(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		// An unreadable session is refused exactly as an unprivileged one
		// is, so the check cannot be used to probe which sessions are live.
		return ErrNotCatalogEditor
	}
	if principal.HasRole(admindomain.RoleFinance) || principal.HasRole(admindomain.RoleAdmin) {
		return nil
	}
	return ErrNotCatalogEditor
}
