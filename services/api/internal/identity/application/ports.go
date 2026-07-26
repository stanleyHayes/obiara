package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

// SessionRepository is the outbound port for session persistence.
// Implementations must enforce optimistic concurrency via the session
// version field (agent_plan.md §7.4).
type SessionRepository interface {
	Create(context.Context, domain.Session) error
	FindByID(context.Context, string) (domain.Session, error)
	// Update persists the session, rejecting stale versions with
	// ErrStaleSession.
	Update(context.Context, domain.Session) error
	// RevokeAllForMember revokes every active session of a member
	// (account takeover response, device-loss step-up).
	RevokeAllForMember(context.Context, string, time.Time) error
	// RevokeAllForDevice revokes every active session bound to a device.
	RevokeAllForDevice(context.Context, string, time.Time) error
}
