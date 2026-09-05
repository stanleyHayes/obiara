// Package circlepolicy decides who may open an introduction source, and which
// members of it may be offered.
//
// These are the two checks the candidate resolver deliberately does not make.
// The resolver returns a circle's roster; this decides whether the asker has
// any standing to ask, and then which of those people they may actually be
// introduced to.
package circlepolicy

import (
	"context"
	"errors"

	circledomain "github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

var ErrNotPermitted = errors.New("not permitted to open that introduction source")

// Circles reads one circle.
type Circles interface {
	Find(context.Context, string) (circledomain.Circle, error)
}

// Authorizer admits only a settled member of the circle being asked about.
//
// Without this a member could name any circle id and receive its roster, keyed
// but still a roster — who is in a private circle is exactly the kind of thing
// this product exists to keep quiet about.
type Authorizer struct {
	circles Circles
}

func NewAuthorizer(circles Circles) Authorizer {
	return Authorizer{circles: circles}
}

func (authorizer Authorizer) Require(
	ctx context.Context,
	requesterID, _ string,
	sourceType domain.SourceType,
	sourceRef string,
) error {
	if sourceType != domain.SourceCircle || authorizer.circles == nil {
		return ErrNotPermitted
	}
	circle, err := authorizer.circles.Find(ctx, sourceRef)
	if err != nil {
		return ErrNotPermitted
	}
	for _, membership := range circle.Memberships() {
		if membership.MemberID() == requesterID && settled(membership.State()) {
			return nil
		}
	}
	return ErrNotPermitted
}

// Policy admits the source types this deployment will resolve. Circle is the
// only one wired; the others are named in the domain and have no resolver
// behind them, and returning candidates for a source nobody can resolve would
// be a promise the next step cannot keep.
type Policy struct{}

func NewPolicy() Policy { return Policy{} }

func (Policy) Allow(_ context.Context, sourceType domain.SourceType, sourceRef string) error {
	if sourceType != domain.SourceCircle || sourceRef == "" {
		return ErrNotPermitted
	}
	return nil
}

// Visibility decides, per candidate, whether this requester may be offered
// them. It runs after the resolver, on each name it returned.
type Visibility struct {
	circles Circles
	blocks  Blocked
}

func NewVisibility(circles Circles) Visibility {
	return Visibility{circles: circles}
}

// WithBlocks attaches the block check. Without it this refuses to offer
// anybody, because introducing two people who have blocked each other is
// worse than introducing nobody.
func (visibility Visibility) WithBlocks(blocks Blocked) Visibility {
	visibility.blocks = blocks
	return visibility
}

// Blocked reports whether either member has blocked the other. Introducing
// two people who have blocked each other is the worst thing this surface
// could do, and until this port existed nothing here asked.
type Blocked interface {
	Blocked(ctx context.Context, memberID, otherID string) (bool, error)
}

func (visibility Visibility) Visible(
	ctx context.Context,
	requesterID, candidateID string,
	sourceType domain.SourceType,
	sourceRef string,
) (bool, error) {
	// Nobody is introduced to themselves. The resolver returns the whole
	// roster because it has no idea who is asking; this is where that is
	// known, so this is where it is removed.
	if requesterID == candidateID {
		return false, nil
	}
	// A block is checked before membership, because it outranks it: being in
	// the same circle as somebody you blocked is not a reason to be offered
	// them. Returning false rather than an error is right here — this
	// surface says nothing about why anybody is absent from it, which is
	// precisely how a block should behave.
	if visibility.blocks == nil {
		// No block check composed means every introduction is made blind to
		// blocks. Refusing to offer anybody is the safe failure.
		return false, nil
	}
	blocked, err := visibility.blocks.Blocked(ctx, requesterID, candidateID)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil
	}
	if sourceType != domain.SourceCircle || visibility.circles == nil {
		return false, nil
	}
	circle, err := visibility.circles.Find(ctx, sourceRef)
	if err != nil {
		return false, nil
	}
	// Re-read rather than trust the resolver's list: the roster was fetched a
	// moment ago and someone may have left in between. A member who has just
	// walked out of a circle should not be offered through it.
	for _, membership := range circle.Memberships() {
		if membership.MemberID() == candidateID {
			return settled(membership.State()), nil
		}
	}
	return false, nil
}

func settled(state circledomain.MembershipState) bool {
	return state == circledomain.StateMember ||
		state == circledomain.StateHost ||
		state == circledomain.StateOwner
}
