package domain

import "testing"

func member(id string, tier Tier, roles ...Role) Subject {
	return Subject{MemberID: id, Tier: tier, Roles: roles}
}

func TestDenyByDefault(t *testing.T) {
	cases := map[string]struct {
		subject  Subject
		action   string
		resource Resource
	}{
		"anonymous":            {member("", TierSowing), "read", Resource{Type: "profile", OwnerID: "m-1"}},
		"empty action":         {member("m-1", TierSowing), "", Resource{Type: "profile", OwnerID: "m-1"}},
		"unknown action":       {member("m-1", TierSowing), "teleport", Resource{Type: "profile", OwnerID: "m-1"}},
		"unknown resource act": {member("m-1", TierSowing), "delete", Resource{Type: "profile", OwnerID: "m-1"}},
		"zero value subject":   {Subject{}, "read", Resource{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if decision := Evaluate(tc.subject, tc.action, tc.resource); decision.Allowed {
				t.Fatalf("Evaluate allowed %+v, want deny", tc)
			}
		})
	}
}

func TestOwnerRules(t *testing.T) {
	own := Resource{Type: "profile", ID: "p-1", OwnerID: "m-1"}
	other := Resource{Type: "profile", ID: "p-2", OwnerID: "m-2"}

	if d := Evaluate(member("m-1", TierUnverified), "read", own); !d.Allowed {
		t.Fatal("owner read must be allowed even at tier 0")
	}
	if d := Evaluate(member("m-1", TierUnverified), "write", own); !d.Allowed {
		t.Fatal("owner write must be allowed even at tier 0")
	}
	// A member must not read or write another member's resource.
	if d := Evaluate(member("m-1", TierSowing), "read", other); d.Allowed {
		t.Fatal("m-1 reading m-2's resource must be denied")
	}
	if d := Evaluate(member("m-1", TierSowing), "write", other); d.Allowed {
		t.Fatal("m-1 writing m-2's resource must be denied")
	}
	// Ownership must be claimed, not defaulted: an ownerless resource is
	// not readable by members.
	if d := Evaluate(member("m-1", TierSowing), "read", Resource{Type: "profile", ID: "p-3"}); d.Allowed {
		t.Fatal("ownerless resource must not fall to member read")
	}
}

func TestTierGatesFR101(t *testing.T) {
	cases := []struct {
		name   string
		tier   Tier
		action string
		want   bool
	}{
		{"tier 0 introductions denied", TierUnverified, "introductions.view", false},
		{"tier 1 introductions allowed", TierVerified, "introductions.view", true},
		{"tier 0 room denied", TierUnverified, "rooms.participate", false},
		{"tier 1 room allowed", TierVerified, "rooms.participate", true},
		{"tier 0 fire denied", TierUnverified, "fires.attend", false},
		{"tier 1 fire allowed", TierVerified, "fires.attend", true},
		{"tier 1 sow denied", TierVerified, "seeds.sow", false},
		{"tier 2 sow allowed", TierSowing, "seeds.sow", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource := Resource{Type: resourceTypeForAction(tc.action)}
			if d := Evaluate(member("m-1", tc.tier), tc.action, resource); d.Allowed != tc.want {
				t.Fatalf("Evaluate(%s at tier %d) = %v, want %v", tc.action, tc.tier, d.Allowed, tc.want)
			}
		})
	}
}

func resourceTypeForAction(action string) string {
	switch action {
	case "introductions.view":
		return "introduction"
	case "rooms.participate":
		return "room"
	case "fires.attend":
		return "fire"
	case "seeds.sow":
		return "seed"
	}
	return ""
}

func TestRoleDesks(t *testing.T) {
	verificationCase := Resource{Type: "verification_case", ID: "vc-1"}
	safetyCase := Resource{Type: "safety_case", ID: "sc-1"}

	if d := Evaluate(member("m-1", TierSowing), "verification.review", verificationCase); d.Allowed {
		t.Fatal("plain member must not reach the verifier desk")
	}
	if d := Evaluate(member("m-1", TierSowing, RoleVerifier), "verification.review", verificationCase); !d.Allowed {
		t.Fatal("verifier must reach the verifier desk")
	}
	if d := Evaluate(member("m-1", TierSowing, RoleVerifier), "safety.review", safetyCase); d.Allowed {
		t.Fatal("verifier role must not leak into the T&S desk (least privilege)")
	}
	if d := Evaluate(member("m-1", TierSowing, RoleTSAgent), "safety.review", safetyCase); !d.Allowed {
		t.Fatal("T&S agent must reach the T&S desk")
	}
}

func TestHostCircleScope(t *testing.T) {
	ownCircle := Resource{Type: "circle", ID: "c-1", OwnerID: "m-1"}
	otherCircle := Resource{Type: "circle", ID: "c-2", OwnerID: "m-2"}

	if d := Evaluate(member("m-1", TierSowing, RoleHost), "circles.host", ownCircle); !d.Allowed {
		t.Fatal("host must manage their own circle")
	}
	if d := Evaluate(member("m-1", TierSowing, RoleHost), "circles.host", otherCircle); d.Allowed {
		t.Fatal("host must not manage another host's circle")
	}
	if d := Evaluate(member("m-9", TierSowing, RoleAdmin), "circles.host", otherCircle); !d.Allowed {
		t.Fatal("admin must reach circle operations")
	}
}
