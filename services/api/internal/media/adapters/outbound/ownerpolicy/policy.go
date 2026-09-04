// Package ownerpolicy decides who may act on a media asset.
//
// The rule is the narrowest one that serves the product: a member may read and
// write their own media, and nobody else's, under a purpose that is on the
// allowed list for that action. Sharing a recording with another member is a
// separate decision made by the introduction and listening contexts, which
// mint their own grants; it is deliberately not a hole in this policy.
package ownerpolicy

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
)

// Policy authorizes only the owner.
type Policy struct {
	purposes map[string]struct{}
}

// New builds a policy that admits the named purposes. An empty list admits
// none — an unrecognised purpose must be a refusal, not a wildcard, or a new
// caller silently inherits access to every member's media.
func New(purposes ...string) Policy {
	allowed := make(map[string]struct{}, len(purposes))
	for _, purpose := range purposes {
		allowed[purpose] = struct{}{}
	}
	return Policy{purposes: allowed}
}

func (policy Policy) Authorize(_ context.Context, decision application.AccessDecision) error {
	if decision.SubjectID == "" || decision.OwnerID == "" ||
		decision.SubjectID != decision.OwnerID {
		return application.ErrAccessDenied
	}
	if _, ok := policy.purposes[decision.Purpose]; !ok {
		return application.ErrAccessDenied
	}
	return nil
}
