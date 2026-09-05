package apihttp

import (
	"context"
	"fmt"
	"net/http"

	authzapplication "github.com/stanleyHayes/obiara/services/api/internal/authz/application"
	authzdomain "github.com/stanleyHayes/obiara/services/api/internal/authz/domain"
)

// MemberGate asks one question at the route boundary: is this member standing
// on a high enough rung of the verification ladder to reach this surface?
//
// The rules are not here. They live in the authorization kernel's grant table
// (FR-101), which is deny-by-default, so a surface that forgets to name an
// action is refused rather than opened. This type only asks the question and
// turns a refusal into something the member can act on.
//
// What it deliberately does not gate is the road out of the gate. A Tier-0
// member must still reach verification, consent, their profile, privacy,
// settings, safety reporting and their own account — gate those and the ladder
// becomes a trap, because becoming verified requires being verified.
type MemberGate struct {
	tiers      TierReader
	authorizer authzapplication.Authorizer
}

// NewMemberGate builds the gate over the identity context's tier read.
func NewMemberGate(tiers TierReader) MemberGate {
	return MemberGate{tiers: tiers, authorizer: authzapplication.NewAuthorizer()}
}

// Allow reports whether memberID may perform action on a resource of
// resourceType, and writes the refusal itself when the answer is no, so a
// caller only has to return.
//
// A tier that cannot be read is a server fault, not a verdict. Answering 403
// there would be indistinguishable from "you are unverified" and would shut
// out members who have already earned the surface, which is the more damaging
// of the two ways this can be wrong. Only a tier that was read and found short
// answers 403.
func (gate MemberGate) Allow(w http.ResponseWriter, r *http.Request, memberID, action, resourceType string) bool {
	if gate.tiers == nil {
		// A gate wired without a tier read cannot decide anything. Refusing
		// loudly beats silently admitting everyone.
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "gate_unavailable",
			Message: "This could not be opened just now. Please try again.",
		})
		return false
	}
	tier, err := gate.tiers.Tier(r.Context(), memberID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "gate_unavailable",
			Message: "This could not be opened just now. Please try again.",
		})
		return false
	}
	subject := authzdomain.Subject{MemberID: memberID, Tier: authzdomain.Tier(tier)}
	if err := gate.authorizer.Require(subject, action, authzdomain.Resource{Type: resourceType}); err != nil {
		writeTierRefusal(w, r, action)
		return false
	}
	return true
}

// writeTierRefusal names the rung that would lift the refusal. A member who is
// told only "forbidden" has no way to know that a few minutes of verification
// would open the door, so the code carries the requirement and the message
// says what to do about it.
func writeTierRefusal(w http.ResponseWriter, r *http.Request, action string) {
	required, gated := authzdomain.RequiredTier(action)
	if !gated {
		writeError(w, r, http.StatusForbidden, APIError{
			Code:    "access_denied",
			Message: "This is not open to you.",
		})
		return
	}
	message := "This opens once your account is cleared for it."
	if required == authzdomain.TierVerified {
		message = "Verify your identity to open this. It takes a few minutes and is only done once."
	}
	writeError(w, r, http.StatusForbidden, APIError{
		Code:    fmt.Sprintf("tier_%d_required", int(required)),
		Message: message,
	})
}

// memberIDKey carries the member the gate already authenticated, so a guarded
// handler does not authenticate the same request twice.
type memberIDKey struct{}

// gatedMember returns the member the gate authenticated for this request.
func gatedMember(ctx context.Context) (string, bool) {
	memberID, ok := ctx.Value(memberIDKey{}).(string)
	return memberID, ok && memberID != ""
}

// guard wraps a handler so the route is refused before it runs unless the
// member stands high enough on the ladder for action.
//
// Gating happens at the registration table rather than inside each handler on
// purpose: for an access-control change, being able to read one list and see
// exactly which routes are gated and at which rung is worth more than saving a
// wrapper. A route absent from that list is ungated, visibly.
func (gate MemberGate) guard(sessions SessionAuthenticator, action, resourceType string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		memberID, ok := authenticatedMember(w, r, sessions)
		if !ok {
			return
		}
		if !gate.Allow(w, r, memberID, action, resourceType) {
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), memberIDKey{}, memberID)))
	})
}
