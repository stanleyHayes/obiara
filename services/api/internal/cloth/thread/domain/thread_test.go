package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func key(n int) string { return fmt.Sprintf("%064x", n) }
func fresh(t *testing.T) Thread {
	t.Helper()
	v, e := New("thread-pair-01", []string{key(2), key(1)})
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func command(actor string) Command {
	return Command{ID: "thread-command-01", ActorKey: actor, RevealRef: "ref_revealabcdefghijklmnop", RecipeRef: "ref_recipeabcdefghijklmnop", BandVersion: 1, At: time.Unix(100, 0)}
}
func TestExactlyTwoOpaqueMembers(t *testing.T) {
	for _, m := range [][]string{{key(1)}, {key(1), key(2), key(3)}, {key(1), "raw-member"}} {
		if _, e := New("thread-pair-01", m); e == nil {
			t.Fatalf("accepted %#v", m)
		}
	}
}
func TestEitherPairMemberIssuesOnceAndViews(t *testing.T) {
	for _, actor := range []string{key(1), key(2)} {
		v, e := fresh(t).Issue(command(actor))
		if e != nil || v.Revision() != 1 {
			t.Fatalf("%d %v", v.Revision(), e)
		}
		view, e := v.View(actor)
		if e != nil || view.Provenance.BandVersion != 1 {
			t.Fatalf("%+v %v", view, e)
		}
		if _, e = v.View(key(8)); e != ErrDenied {
			t.Fatalf("%v", e)
		}
		next := command(key(2))
		next.ID = "thread-command-02"
		next.ExpectedRevision = 1
		if _, e = v.Issue(next); e != ErrAlreadyIssued {
			t.Fatalf("%v", e)
		}
	}
}
func TestReplayFingerprintAndCAS(t *testing.T) {
	cmd := command(key(1))
	v, _ := fresh(t).Issue(cmd)
	replayed, e := v.Issue(cmd)
	if e != nil || replayed.Revision() != 1 {
		t.Fatalf("%d %v", replayed.Revision(), e)
	}
	cmd.RecipeRef = "ref_changedabcdefghijklmnop"
	if _, e = v.Issue(cmd); e != ErrCommandMismatch {
		t.Fatalf("%v", e)
	}
	cmd = command(key(1))
	cmd.ExpectedRevision = 7
	if _, e = fresh(t).Issue(cmd); e != ErrStaleRevision {
		t.Fatalf("%v", e)
	}
}
func TestImmutableVersionedOpaqueProvenanceAndPrivacy(t *testing.T) {
	v, _ := fresh(t).Issue(command(key(1)))
	p := v.Provenance()
	p.RecipeRef = "ref_mutatedabcdefghijklmnop"
	if v.Provenance().RecipeRef == p.RecipeRef {
		t.Fatal("mutable provenance")
	}
	wire := strings.ToLower(fmt.Sprintf("%+v", State{ID: v.ID(), Members: v.Members(), Provenance: v.Provenance(), Revision: v.Revision(), Commands: v.Commands()}))
	for _, bad := range []string{"raw response", "relationship", "purchase", "bypass", "public", "reverse", "list"} {
		if strings.Contains(wire, bad) {
			t.Fatalf("leak %q", bad)
		}
	}
}
func TestRejectsRawOrUnversionedInputs(t *testing.T) {
	for _, mutate := range []func(*Command){func(c *Command) { c.RevealRef = "my answer" }, func(c *Command) { c.RecipeRef = "recipe" }, func(c *Command) { c.BandVersion = 0 }} {
		c := command(key(1))
		mutate(&c)
		if _, e := fresh(t).Issue(c); e != ErrDenied {
			t.Fatalf("%v", e)
		}
	}
}
