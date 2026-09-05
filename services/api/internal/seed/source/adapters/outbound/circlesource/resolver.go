// Package circlesource resolves introduction candidates from a circle the
// requester already belongs to.
//
// This is the narrowest possible answer to "who could I be introduced to":
// the settled members of one circle. It is not a search and not a ranking —
// the port it satisfies is explicit that it "must never expose a global
// member list, reverse graph, or profile payload", and the way to honour that
// is to make one circle the only scope.
//
// It does not decide who may ask. The service checks authority before calling
// this and filters each candidate through consent visibility afterwards, so
// this returns the circle's roster and nothing about the requester.
package circlesource

import (
	"context"
	"errors"

	circledomain "github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

var (
	// ErrUnreadable covers both a circle that does not exist and one this
	// process cannot read. They are the same answer on purpose: distinguishing
	// them would let a caller probe which private circles are real.
	ErrUnreadable  = errors.New("that circle could not be read")
	ErrUnsupported = errors.New("unsupported introduction source")
)

// Circles reads one circle. Only Find is needed — listing circles is a
// different question and belongs to a different surface.
type Circles interface {
	Find(context.Context, string) (circledomain.Circle, error)
}

type Resolver struct {
	circles Circles
}

func NewResolver(circles Circles) Resolver {
	return Resolver{circles: circles}
}

// CandidateIDs returns other members of the circle, bounded by limit.
//
// Only settled memberships count. Someone who has requested and not been
// approved, or who left or was expelled, is not a person this member shares a
// circle with, and introducing through them would leak that they ever asked.
func (resolver Resolver) CandidateIDs(
	ctx context.Context,
	sourceType domain.SourceType,
	sourceRef string,
	limit int,
) ([]string, error) {
	if sourceType != domain.SourceCircle {
		return nil, ErrUnsupported
	}
	if resolver.circles == nil || limit < 1 {
		return nil, ErrUnsupported
	}
	circle, err := resolver.circles.Find(ctx, sourceRef)
	if err != nil {
		return nil, ErrUnreadable
	}

	candidates := make([]string, 0, limit)
	for _, membership := range circle.Memberships() {
		if !settled(membership.State()) {
			continue
		}
		candidates = append(candidates, membership.MemberID())
		if len(candidates) == limit {
			break
		}
	}
	return candidates, nil
}

func settled(state circledomain.MembershipState) bool {
	return state == circledomain.StateMember ||
		state == circledomain.StateHost ||
		state == circledomain.StateOwner
}
