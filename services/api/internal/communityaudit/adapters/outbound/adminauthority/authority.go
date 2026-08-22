// Package adminauthority answers the community audit desk's authority and
// MFA questions from the composed admin context.
package adminauthority

import (
	"context"
	"strings"
	"time"

	admindomain "github.com/stanleyHayes/obiara/services/api/internal/admin/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/application"
)

// Admin is the surface these bridges need from the admin context.
type Admin interface {
	Authenticate(ctx context.Context, sessionID string) (admindomain.Session, admindomain.Principal, error)
}

// Authority gates the desk on the trust-and-safety roles.
//
// A community audit reviews conduct within the community, which is the same
// responsibility the authorization kernel already assigns to the trust and
// safety desk. Reading that rule rather than inventing a new capability
// keeps one answer to "who reviews members" instead of two that can drift.
type Authority struct{ admin Admin }

func NewAuthority(admin Admin) *Authority { return &Authority{admin: admin} }

func (authority *Authority) Authorize(ctx context.Context, sessionID string, _ application.Capability) error {
	_, principal, err := authority.admin.Authenticate(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return application.ErrDenied
	}
	if principal.HasRole(admindomain.RoleTSAgent) || principal.HasRole(admindomain.RoleAdmin) {
		return nil
	}
	return application.ErrDenied
}

// MFAGate reports whether the operator has completed a recent step-up.
//
// Opening evidence about a member and deciding their case are the two
// actions the admin context already protects with step-up, so the gate reads
// the session's flag rather than keeping a second notion of recency.
type MFAGate struct{ admin Admin }

func NewMFAGate(admin Admin) *MFAGate { return &MFAGate{admin: admin} }

func (gate *MFAGate) Recent(ctx context.Context, sessionID string, at time.Time) bool {
	session, _, err := gate.admin.Authenticate(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return false
	}
	// An expired session is not a stepped-up one, whatever the flag says.
	return session.SteppedUp() && session.Active(at)
}
