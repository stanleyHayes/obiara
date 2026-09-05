// Package media binds a sow's recordings to the media context.
//
// The sow context asks one question — are these recordings this member's? —
// and the media context already knows, because every asset records its owner.
// This only translates between the two vocabularies.
package media

import (
	"context"
	"errors"

	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

// Assets reads back what storage recorded about an asset.
type Assets interface {
	FindByID(context.Context, string) (mediadomain.Asset, error)
}

// Ownership answers whether recordings belong to the member sowing them.
type Ownership struct{ assets Assets }

func NewOwnership(assets Assets) Ownership { return Ownership{assets: assets} }

// OwnedBy reports whether every reference belongs to ownerID.
//
// A reference that cannot be read is not owned. The alternative — treating an
// unreadable asset as somebody's own — would let a member attach a reference
// to a recording nobody can account for, which is exactly the case the check
// exists for.
//
// A deleted asset is not owned either. Its row survives deletion so retention
// stays provable, and a member whose recording has been erased must not be
// able to keep sowing it.
func (ownership Ownership) OwnedBy(ctx context.Context, ownerID string, refs []string) (bool, error) {
	if ownership.assets == nil || ownerID == "" {
		return false, errors.New("sow media ownership is not composed")
	}
	for _, ref := range refs {
		asset, err := ownership.assets.FindByID(ctx, ref)
		if err != nil {
			// Not an error to the caller: the answer to "is this theirs" is
			// no. Reporting a fault here would refuse the sow with an outage
			// when the truthful answer is a refusal.
			return false, nil
		}
		if asset.IsDeleted() || asset.OwnerID() != ownerID {
			return false, nil
		}
	}
	return true, nil
}
