// Package fireauthority answers the run sheet's authority question from the
// composed fire module: may this member run this fire?
package fireauthority

import (
	"context"
	"errors"
	"strings"

	firedomain "github.com/stanleyHayes/obiara/services/api/internal/fire/domain"
)

// ErrNotHost reports a member who does not run the fire.
var ErrNotHost = errors.New("member does not host this fire")

// Fires is the surface this bridge needs from the fire context.
type Fires interface {
	FindByID(context.Context, string) (firedomain.Fire, error)
}

// Authority permits only the fire's host.
//
// The fire aggregate has a single host and no cohost concept. Rather than
// invent one, this reads the absent concept the only safe way it can: a
// cohort that does not exist grants nobody access. If cohosts are added to
// the fire aggregate later, this is the one place that changes.
type Authority struct{ fires Fires }

func New(fires Fires) *Authority { return &Authority{fires: fires} }

// RequireHostOrCohost takes raw identifiers: the run sheet keys them only
// after authority has been settled.
func (authority *Authority) RequireHostOrCohost(ctx context.Context, fireID, memberID string) error {
	fire, err := authority.fires.FindByID(ctx, strings.TrimSpace(fireID))
	if err != nil {
		// A fire nobody can read is one nobody may run. The caller reports
		// this the same way it reports a refusal, so a member cannot use it
		// to discover which fires exist.
		return ErrNotHost
	}
	if fire.HostID() != strings.TrimSpace(memberID) || fire.HostID() == "" {
		return ErrNotHost
	}
	return nil
}
