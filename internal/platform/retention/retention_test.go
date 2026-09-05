package retention

import (
	"testing"
	"time"
)

// TestEveryBindingPolicyCanActuallyDoSomething guards the failure mode this
// table is most exposed to: a policy that looks present and deletes nothing.
// A strip policy naming no fields unsets nothing, and a policy with no date
// field matches nothing — both run clean, report success, and quietly retain
// the data they were added to remove.
func TestEveryBindingPolicyCanActuallyDoSomething(t *testing.T) {
	seen := map[string]bool{}
	for _, policy := range BindingPolicies {
		if policy.Name == "" || policy.Collection == "" || policy.DateField == "" {
			t.Fatalf("policy %#v is missing a name, collection or date field", policy)
		}
		if seen[policy.Name] {
			// Names are how a proof record identifies its run.
			t.Fatalf("two policies share the name %q", policy.Name)
		}
		seen[policy.Name] = true
		if policy.MaxAge <= 0 {
			t.Fatalf("policy %s has no retention window", policy.Name)
		}
		if policy.Action == ActionStripField && len(policy.Fields) == 0 {
			t.Fatalf("policy %s strips no fields, so it would run clean and retain everything", policy.Name)
		}
		if policy.Action != ActionStripField && len(policy.Fields) > 0 {
			t.Fatalf("policy %s names fields its action ignores", policy.Name)
		}
	}
}

// TestNationalIdentityDataHasAnEndDate pins M1-02's deletion half. Before
// these policies a verified member's date of birth and both photographs of
// their Ghana Card were kept in Mongo permanently.
func TestNationalIdentityDataHasAnEndDate(t *testing.T) {
	byName := map[string]Policy{}
	for _, policy := range BindingPolicies {
		byName[policy.Name] = policy
	}

	images, ok := byName["identity_documents_delete_90d"]
	if !ok {
		t.Fatal("Ghana Card images have no retention policy")
	}
	if images.Collection != "identity_documents" || images.Action != ActionDelete {
		t.Fatalf("card images policy = %#v", images)
	}

	// Two policies, because a case that is never decided has no decidedAt and
	// the decided-case policy can never match it.
	decided, ok := byName["identity_dob_strip_decided_30d"]
	if !ok {
		t.Fatal("decided cases keep their date of birth forever")
	}
	stale, ok := byName["identity_dob_strip_stale_180d"]
	if !ok {
		t.Fatal("abandoned cases keep their date of birth forever")
	}
	for _, policy := range []Policy{decided, stale} {
		if policy.Collection != "identity_verifications" || policy.Action != ActionStripField {
			t.Fatalf("date-of-birth policy = %#v", policy)
		}
		if len(policy.Fields) != 1 || policy.Fields[0] != "dateOfBirth" {
			t.Fatalf("policy %s strips %v, want only dateOfBirth", policy.Name, policy.Fields)
		}
	}
	if decided.DateField != "decidedAt" || stale.DateField != "createdAt" {
		t.Fatalf("the two date-of-birth policies must key off different dates: %q and %q",
			decided.DateField, stale.DateField)
	}
	// The backstop must not be able to strip a case that is still being
	// reviewed before the decided-case policy would have.
	if stale.MaxAge <= decided.MaxAge {
		t.Fatalf("the abandoned-case backstop (%v) is not longer than the decided window (%v)",
			stale.MaxAge, decided.MaxAge)
	}
	if images.MaxAge < 30*24*time.Hour {
		// The document store refuses a TTL on purpose: an image that expired
		// mid-review leaves a member unverifiable with nothing explaining why.
		t.Fatalf("card images expire after %v, inside a plausible review queue", images.MaxAge)
	}
}
