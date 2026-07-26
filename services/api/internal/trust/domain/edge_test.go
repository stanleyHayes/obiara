package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDirectedEdgeIsImmutableRevocableAndExpiring(t *testing.T) {
	expiry := baseTime().Add(time.Hour)
	edge := testEdge(t, VisibilityConsentedPath, &expiry)
	if !edge.Active(baseTime().Add(30 * time.Minute)) {
		t.Fatal("edge expired early")
	}
	if edge.Active(baseTime().Add(time.Hour)) {
		t.Fatal("edge active at expiry boundary")
	}
	command := Command{
		ID: "revoke-1", ExpectedRevision: 1, ActorRef: "actor-1",
		Kind: "edge.revoke", Payload: "edge-1", RecordedAt: baseTime().Add(time.Minute),
	}
	revoked, err := edge.Revoke(command)
	if err != nil || revoked.Active(baseTime().Add(2*time.Minute)) {
		t.Fatalf("revoked active = %v, error = %v", revoked.Active(baseTime().Add(2*time.Minute)), err)
	}
	replay, err := revoked.Revoke(command)
	if err != nil || replay.Revision() != 2 {
		t.Fatalf("replay revision = %d, error = %v", replay.Revision(), err)
	}
	command.Payload = "different"
	if _, err := revoked.Revoke(command); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("mismatch error = %v, want %v", err, ErrCommandMismatch)
	}
	if revoked.SourceID() != edge.SourceID() || revoked.TargetID() != edge.TargetID() ||
		revoked.ProvenanceRef() != edge.ProvenanceRef() {
		t.Fatal("revocation mutated immutable provenance")
	}
}

func TestVisibilityIsExplicitAndFailClosed(t *testing.T) {
	now := baseTime().Add(time.Minute)
	ownerOnly := testEdge(t, VisibilityOwnerOnly, nil)
	if !ownerOnly.VisibleTo("member-a", true, now) || ownerOnly.VisibleTo("member-b", true, now) {
		t.Fatal("owner-only visibility escaped source owner")
	}
	participants := testEdge(t, VisibilityParticipants, nil)
	if !participants.VisibleTo("member-b", true, now) || participants.VisibleTo("stranger", true, now) {
		t.Fatal("participant visibility escaped endpoints")
	}
	consented := testEdge(t, VisibilityConsentedPath, nil)
	if consented.VisibleTo("owner", false, now) || !consented.VisibleTo("owner", true, now) {
		t.Fatal("consented path did not honor consent decision")
	}
}

func FuzzCreateRejectsInvalidIdentity(f *testing.F) {
	f.Add("", "target")
	f.Add("source", "source")
	f.Add("source with spaces", "target")
	f.Fuzz(func(t *testing.T, source, target string) {
		command := Command{
			ID: "create-fuzz", ActorRef: "actor-1", Kind: "edge.create",
			Payload: "payload", RecordedAt: baseTime(),
		}
		edge, err := Create(Params{
			ID: "edge-fuzz", SourceID: source, TargetID: target, Type: EdgeKnown,
			ProvenanceRef: "provenance-1", ConsentRef: "consent-1",
			Visibility: VisibilityConsentedPath, CreatedAt: baseTime(),
		}, command)
		if err == nil {
			if edge.SourceID() == edge.TargetID() || !opaquePattern.MatchString(edge.SourceID()) ||
				!opaquePattern.MatchString(edge.TargetID()) {
				t.Fatalf("invalid edge accepted: %#v", edge)
			}
		}
	})
}

func testEdge(t *testing.T, visibility Visibility, expiresAt *time.Time) Edge {
	t.Helper()
	edge, err := Create(Params{
		ID: "edge-1", SourceID: "member-a", TargetID: "member-b", Type: EdgeKnown,
		ProvenanceRef: "provenance-1", ConsentRef: "consent-1",
		Visibility: visibility, CreatedAt: baseTime(), ExpiresAt: expiresAt,
	}, Command{
		ID: "create-1", ActorRef: "actor-1", Kind: "edge.create",
		Payload: "payload", RecordedAt: baseTime(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return edge
}

func baseTime() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}
