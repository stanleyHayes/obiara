// Package domain holds the authorization policy model (agent_plan.md §11:
// deny-by-default, tested at route and use-case boundaries). A request is
// allowed only when an explicit rule grants it; anything unmatched is
// denied. Tier gates encode FR-101: romantic surfaces require Tier 1,
// sowing requires Tier 2.
package domain

import "strings"

// Tier mirrors the verification ladder from the identity context.
type Tier int

const (
	TierUnverified Tier = 0
	TierVerified   Tier = 1
	TierSowing     Tier = 2
)

// Role is a staff or community capability assignment. Member-facing access
// is primarily attribute-based (tier, ownership); roles govern operational
// desks.
type Role string

const (
	RoleHost     Role = "host"
	RoleTSAgent  Role = "ts_agent"
	RoleVerifier Role = "verifier"
	RoleAdmin    Role = "admin"
)

// Subject is the authenticated actor as asserted by the identity context.
type Subject struct {
	MemberID string
	Roles    []Role
	Tier     Tier
}

func (subject Subject) HasRole(role Role) bool {
	for _, assigned := range subject.Roles {
		if assigned == role {
			return true
		}
	}
	return false
}

// Resource is the target of an action. OwnerID enables
// attribute-based owner rules; empty means the resource has no owner.
type Resource struct {
	Type    string
	ID      string
	OwnerID string
}

// Decision is the outcome of a policy evaluation. Deny is the zero value —
// the deny-by-default posture is structural, not a fallback branch.
type Decision struct {
	Allowed bool
	Reason  string
}

func allow(reason string) Decision { return Decision{Allowed: true, Reason: reason} }
func deny(reason string) Decision  { return Decision{Allowed: false, Reason: reason} }

// rule matches an action and optional resource type to a grant check.
type rule struct {
	action       string
	resourceType string
	grant        func(Subject, Resource) (bool, string)
}

// rules is the explicit grant table. Order does not matter: any single
// grant allows; absence of a grant denies. New capabilities must be added
// here deliberately — never by widening an existing grant.
var rules = append([]rule{
	// Members act on resources they own (profiles, consent records,
	// privacy requests, notification preferences).
	{"read", "", func(subject Subject, resource Resource) (bool, string) {
		if subject.MemberID != "" && resource.OwnerID == subject.MemberID {
			return true, "owner read"
		}
		return false, ""
	}},
	{"write", "", func(subject Subject, resource Resource) (bool, string) {
		if subject.MemberID != "" && resource.OwnerID == subject.MemberID {
			return true, "owner write"
		}
		return false, ""
	}},
	// Operational desks (role-gated; least privilege per desk).
	{"verification.review", "verification_case", func(subject Subject, _ Resource) (bool, string) {
		if subject.HasRole(RoleVerifier) || subject.HasRole(RoleAdmin) {
			return true, "verifier desk"
		}
		return false, ""
	}},
	{"safety.review", "safety_case", func(subject Subject, _ Resource) (bool, string) {
		if subject.HasRole(RoleTSAgent) || subject.HasRole(RoleAdmin) {
			return true, "trust and safety desk"
		}
		return false, ""
	}},
	{"circles.host", "circle", func(subject Subject, resource Resource) (bool, string) {
		if subject.HasRole(RoleHost) && resource.OwnerID == subject.MemberID {
			return true, "host of own circle"
		}
		if subject.HasRole(RoleAdmin) {
			return true, "admin circle operations"
		}
		return false, ""
	}},
}, tierRules()...)

// tierGates is FR-101 written once: the action, the resource it names, and the
// rung a member must be standing on to reach it. Both the grant table and
// RequiredTier are derived from this, so a surface can never be gated at one
// tier and explained to the member at another.
var tierGates = []struct {
	action       string
	resourceType string
	minimum      Tier
	reason       string
}{
	{"introductions.view", "introduction", TierVerified, "tier 1 romantic surface"},
	{"rooms.participate", "room", TierVerified, "tier 1 room participation"},
	{"fires.attend", "fire", TierVerified, "tier 1 fire entry"},
	{"seeds.sow", "seed", TierSowing, "tier 2 sowing"},
	// M1-AC-01: an unverified account may look, but may not take part.
	// Placed under the FR-101 default — participatory surfaces require
	// Tier 1 — and recorded in agent_plan.md §38 so they can be overruled.
	{"circles.participate", "circle", TierVerified, "tier 1 circle participation"},
	{"games.play", "game", TierVerified, "tier 1 game participation"},
}

func tierRules() []rule {
	built := make([]rule, 0, len(tierGates))
	for _, gate := range tierGates {
		minimum, reason := gate.minimum, gate.reason
		built = append(built, rule{gate.action, gate.resourceType,
			func(subject Subject, _ Resource) (bool, string) {
				if subject.Tier >= minimum {
					return true, reason
				}
				return false, ""
			}})
	}
	return built
}

// RequiredTier reports the rung an action demands, so a refusal can name what
// would lift it rather than only that it was refused. The second result is
// false for actions that are not tier-gated at all.
func RequiredTier(action string) (Tier, bool) {
	for _, gate := range tierGates {
		if gate.action == strings.TrimSpace(action) {
			return gate.minimum, true
		}
	}
	return TierUnverified, false
}

// Evaluate decides whether subject may perform action on resource.
// It never panics and never defaults to allow.
func Evaluate(subject Subject, action string, resource Resource) Decision {
	action = strings.TrimSpace(action)
	if subject.MemberID == "" || action == "" {
		return deny("anonymous or empty action")
	}
	for _, candidate := range rules {
		if candidate.action != action {
			continue
		}
		if candidate.resourceType != "" && candidate.resourceType != resource.Type {
			continue
		}
		if granted, reason := candidate.grant(subject, resource); granted {
			return allow(reason)
		}
	}
	return deny("no grant")
}
