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
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

var (
	// ErrNotArrived says the recording is the member's own and its bytes are
	// not in storage.
	//
	// It is the sow's own error rather than a new one with the same words:
	// two errors reading alike but comparing unequal is how a caller's
	// errors.Is silently stops matching, which cost a whole session once
	// already (agent_plan.md §49).
	ErrNotArrived = sowapplication.ErrMediaNotArrived
	// ErrArrivalUnknown reports an ownership check composed without the
	// arrival check. It refuses, rather than treating an absent check as a
	// recording that landed.
	ErrArrivalUnknown = errors.New("whether that recording arrived cannot be established")
)

// Assets reads back what storage recorded about an asset.
type Assets interface {
	FindByID(context.Context, string) (mediadomain.Asset, error)
}

// Arrival reports how many bytes are actually in the bucket for an object.
type Arrival interface {
	Stat(context.Context, string) (int64, error)
}

// Ownership answers whether recordings belong to the member sowing them, and
// whether they are actually there.
type Ownership struct {
	assets  Assets
	arrival Arrival
}

func NewOwnership(assets Assets) Ownership { return Ownership{assets: assets} }

// WithArrival adds the check that the bytes reached storage.
//
// A sow is delivered as a pod resting at somebody's house front, and a pod
// whose recording never arrived plays nothing. The member would have spent a
// seed on silence and the recipient would meet a broken player, so the sow is
// refused at the door instead.
func (ownership Ownership) WithArrival(arrival Arrival) Ownership {
	ownership.arrival = arrival
	return ownership
}

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
		if ownership.arrival == nil {
			// A missing check is not permission. Without it a sow could carry
			// a recording nobody ever uploaded.
			return false, ErrArrivalUnknown
		}
		stored, statErr := ownership.arrival.Stat(ctx, asset.ObjectKey())
		if statErr != nil || stored != asset.Size() {
			return false, ErrNotArrived
		}
	}
	return true, nil
}
