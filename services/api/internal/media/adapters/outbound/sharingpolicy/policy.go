// Package sharingpolicy decides who may hear a recording that is not theirs.
//
// The media context's other policy, ownerpolicy, admits a member to their own
// media and nobody else's. That is right for uploading and right as a floor,
// and for a while it was the only policy composed — which meant that in a
// product where people meet through their voices, nobody could hear anybody.
// A pod resting at a house front could be listed and never opened, and the
// twenty-second listen gate that arms a sow could never be satisfied, because
// obtaining a grant to hear another member's Voice of Introduction was
// refused by the only rule there was. See agent_plan.md §63.
//
// This does not widen ownerpolicy. It adds, per purpose, a named entitlement
// that has to say yes — so admitting a recipient to a pod says nothing about
// who may hear anything else.
package sharingpolicy

import (
	"context"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
)

// Entitlement answers whether one listener may hear one recording.
//
// It takes the owner as well as the asset because most reasons to share are
// about the two people rather than about the file: a block is between
// members, and so is delivery.
//
// It answers a bool and an error, and the two are not the same: an
// entitlement that cannot be read refuses, because a check that failed is not
// a check that passed.
type Entitlement interface {
	MayHear(ctx context.Context, listenerID, ownerID, assetID string) (bool, error)
}

// Policy admits the owner under any allowed purpose, and a non-owner only
// under a purpose that has an entitlement, and only when that entitlement
// says so.
type Policy struct {
	purposes     map[string]struct{}
	entitlements map[string]Entitlement
}

// New builds the policy.
//
// purposes is the closed list of purposes this deployment authorizes at all;
// an unlisted purpose is refused for everybody, owner included. entitlements
// is keyed by purpose: a purpose with no entry stays owner-only, which is the
// safe direction for a purpose somebody adds later and forgets to think
// about.
func New(purposes []string, entitlements map[string]Entitlement) Policy {
	allowed := make(map[string]struct{}, len(purposes))
	for _, purpose := range purposes {
		allowed[purpose] = struct{}{}
	}
	shared := make(map[string]Entitlement, len(entitlements))
	for purpose, entitlement := range entitlements {
		if entitlement == nil {
			// A nil entitlement is not "allow"; it is a purpose nobody
			// finished wiring. Dropping it leaves that purpose owner-only.
			continue
		}
		shared[purpose] = entitlement
	}
	return Policy{purposes: allowed, entitlements: shared}
}

func (policy Policy) Authorize(ctx context.Context, decision application.AccessDecision) error {
	subject := strings.TrimSpace(decision.SubjectID)
	owner := strings.TrimSpace(decision.OwnerID)
	if subject == "" || owner == "" {
		return application.ErrAccessDenied
	}
	if _, ok := policy.purposes[decision.Purpose]; !ok {
		return application.ErrAccessDenied
	}
	if subject == owner {
		return nil
	}
	// Only reading is ever shared. Writing to somebody else's recording, or
	// uploading as them, has no legitimate caller — and a purpose that
	// entitles a listener must not quietly entitle an author.
	if decision.Action != application.ActionRead {
		return application.ErrAccessDenied
	}
	entitlement, ok := policy.entitlements[decision.Purpose]
	if !ok {
		return application.ErrAccessDenied
	}
	mayHear, err := entitlement.MayHear(ctx, subject, owner, decision.AssetID)
	if err != nil || !mayHear {
		return application.ErrAccessDenied
	}
	return nil
}
