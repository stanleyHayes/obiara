package domain

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

func key(value string) string { return strings.Repeat(value, 64) }
func cmd(id, actor, reason string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorKey: key(actor), ExpectedRevision: revision, ReasonCode: reason, At: at}
}
func fixture(t *testing.T, units uint8) (Attestation, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	a, err := Propose("attestation-1", key("a"), key("b"), SubjectScope{Kind: "circle", Key: key("c")},
		units, "policy-v1", now.Add(time.Hour), cmd("propose-1", "b", "voucher_proposed", 0, now))
	if err != nil {
		t.Fatal(err)
	}
	return a, now
}
func TestStakeIsBoundedAndConsentCreatesImmutableProvenance(t *testing.T) {
	now := time.Now()
	if _, err := Propose("a", key("a"), key("b"), SubjectScope{Kind: "circle", Key: key("c")},
		101, "policy-v1", now.Add(time.Hour), cmd("p", "b", "voucher_proposed", 0, now)); !errors.Is(err, ErrInvalidAttestation) {
		t.Fatalf("unbounded stake = %v", err)
	}
	a, base := fixture(t, 25)
	active, err := a.Consent(cmd("consent-1", "b", "voucher_consented", 1, base.Add(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	provenance := active.Provenance()
	if provenance == nil || provenance.PolicyVersion != "policy-v1" || provenance.VoucherKey != key("b") {
		t.Fatalf("provenance=%+v", provenance)
	}
	before := active.Events()
	revoked, err := active.Revoke(cmd("revoke-1", "d", "policy_revoked", 2, base.Add(2*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	if revoked.Status() != StatusRevoked || len(revoked.Events()) != 3 ||
		before[0] != revoked.Events()[0] || before[1] != revoked.Events()[1] {
		t.Fatal("revocation mutated historical provenance")
	}
}
func TestExpiryAppendsTerminalState(t *testing.T) {
	a, now := fixture(t, 10)
	if _, err := a.Consent(cmd("late", "b", "voucher_consented", 1, now.Add(time.Hour))); !errors.Is(err, ErrAttestationExpired) {
		t.Fatalf("late consent=%v", err)
	}
	expired, err := a.Expire(cmd("expire", "d", "attestation_expired", 1, now.Add(time.Hour)))
	if err != nil || expired.Status() != StatusExpired || len(expired.Events()) != 2 {
		t.Fatalf("expired=%s events=%d err=%v", expired.Status(), len(expired.Events()), err)
	}
}
func TestPropertyAcceptedHistoriesAlwaysRehydrate(t *testing.T) {
	property := func(actions []byte) bool {
		if len(actions) > 50 {
			actions = actions[:50]
		}
		a, now := fixture(t, 20)
		for index, value := range actions {
			change := cmd(fmt.Sprintf("property-%d", index), "d", "property_action", a.Revision(), now.Add(time.Duration(index+1)*time.Second))
			var next Attestation
			var err error
			switch value % 3 {
			case 0:
				change.ActorKey = key("b")
				next, err = a.Consent(change)
			case 1:
				next, err = a.Revoke(change)
			default:
				change.At = now.Add(time.Hour)
				next, err = a.Expire(change)
			}
			if err == nil {
				a = next
			}
			rehydrated, hydrateErr := Rehydrate(State{
				ID: a.ID(), SubjectKey: a.SubjectKey(), VoucherKey: a.VoucherKey(), Scope: a.Scope(),
				StakeUnits: a.StakeUnits(), PolicyVersion: a.PolicyVersion(), Status: a.Status(),
				ExpiresAt: a.ExpiresAt(), Provenance: a.Provenance(), EndedAt: a.EndedAt(),
				Revision: a.Revision(), Events: a.Events(), Commands: a.Commands(),
			})
			if hydrateErr != nil || rehydrated.Status() != a.Status() || rehydrated.Revision() != a.Revision() {
				return false
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatal(err)
	}
}
