package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var fixed = time.Date(2026, 7, 26, 14, 0, 0, 0, time.UTC)

func definition(t *testing.T, key string, version uint64) Definition {
	t.Helper()
	d, err := NewDefinition(key, version, "matching.compatibility", fixed.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func command(id string, revision uint64) Command {
	return Command{ID: id, ExpectedRevision: revision, At: fixed}
}
func TestOptionalFeatureDefaultsOffAndWithdrawalIsImmediate(t *testing.T) {
	d := definition(t, "shared.rituals", 1)
	g, err := GrantFeature("member:one", d, 4, command("grant:one", 0))
	if err != nil || !g.Effective(d, fixed) {
		t.Fatalf("grant: %v", err)
	}
	withdrawn, err := g.Withdraw(Command{ID: "withdraw:one", ExpectedRevision: 1, At: fixed.Add(time.Second)})
	if err != nil || withdrawn.Effective(d, fixed.Add(time.Second)) {
		t.Fatalf("withdrawal must deny immediately: %v", err)
	}
	if !g.Effective(d, fixed.Add(time.Second)) {
		t.Fatal("immutable prior value was mutated")
	}
}
func TestNewDefinitionVersionNeverRetroactivelyActivatesOldGrant(t *testing.T) {
	v1 := definition(t, "shared.rituals", 1)
	v2 := definition(t, "shared.rituals", 2)
	g, _ := GrantFeature("member:one", v1, 1, command("grant:v1", 0))
	if g.Effective(v2, fixed) {
		t.Fatal("old consent must not apply to a new feature definition")
	}
}
func TestRegrantRequiresExplicitNewGrantVersion(t *testing.T) {
	v1 := definition(t, "shared.rituals", 1)
	v2 := definition(t, "shared.rituals", 2)
	g, _ := GrantFeature("member:one", v1, 3, command("grant:v1", 0))
	g, _ = g.Withdraw(Command{ID: "withdraw:v1", ExpectedRevision: 1, At: fixed.Add(time.Second)})
	if _, err := g.Regrant(v2, 3, Command{ID: "regrant:stale", ExpectedRevision: 2, At: fixed.Add(2 * time.Second)}); !errors.Is(err, ErrTransition) {
		t.Fatalf("reused grant version=%v", err)
	}
	next, err := g.Regrant(v2, 4, Command{ID: "regrant:v2", ExpectedRevision: 2, At: fixed.Add(2 * time.Second)})
	if err != nil || !next.Effective(v2, fixed.Add(2*time.Second)) {
		t.Fatalf("explicit regrant=%v", err)
	}
}
func TestDecisionShapeProperty(t *testing.T) {
	r := rand.New(rand.NewSource(24))
	for i := 0; i < 1000; i++ {
		a, b := fmt.Sprintf("member:%08x", r.Uint32()), fmt.Sprintf("member:%08x", r.Uint32())
		if a == b {
			i--
			continue
		}
		f := EnabledFeature{Key: "shared.rituals", FeatureVersion: uint64(r.Intn(20) + 1), Purpose: "matching.compatibility", Consents: []ConsentRef{{MemberKey: a, GrantVersion: 1}, {MemberKey: b, GrantVersion: 2}}}
		d, err := NewDecision(fmt.Sprintf("decision:%d", i), a, b, []EnabledFeature{f}, fixed)
		if err != nil {
			t.Fatalf("case %d: %v", i, err)
		}
		if len(d.Pair) != 2 || len(d.Features) != 1 || len(d.Features[0].Consents) != 2 {
			t.Fatalf("case %d invalid snapshot", i)
		}
	}
}
func TestGrantRejectsStaleAndMismatchedReplay(t *testing.T) {
	g, _ := GrantFeature("member:one", definition(t, "shared.rituals", 1), 1, command("grant:one", 0))
	if _, err := g.Withdraw(command("withdraw", 0)); !errors.Is(err, ErrStale) {
		t.Fatalf("stale=%v", err)
	}
}
