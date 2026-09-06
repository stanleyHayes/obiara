package organization

import (
	"strings"
	"testing"
)

func TestAShortSecretIsRefusedAtStartup(t *testing.T) {
	// A weak secret on an audit trail is worse than a crash: the trail would
	// look like it protects who acted while being trivially reversible.
	if _, err := newKeyer([]byte("too short")); err == nil {
		t.Fatal("a short secret was accepted")
	}
}

func TestTheSameOperatorAlwaysKeysTheSameWay(t *testing.T) {
	k, err := newKeyer([]byte(strings.Repeat("s", 32)))
	if err != nil {
		t.Fatal(err)
	}
	first, err := k.Key("organization_operator", "operator-1")
	if err != nil {
		t.Fatal(err)
	}
	second, _ := k.Key("organization_operator", "operator-1")
	if first != second || len(first) != 64 {
		t.Fatalf("unstable digest: %q then %q", first, second)
	}
	// A different namespace is a different digest, so a key from this context
	// can never be compared with one from another.
	other, _ := k.Key("something_else", "operator-1")
	if other == first {
		t.Fatal("the namespace does not separate digests")
	}
	// And a different operator is a different digest, which is what makes the
	// trail able to distinguish them at all.
	different, _ := k.Key("organization_operator", "operator-2")
	if different == first {
		t.Fatal("two operators share a digest")
	}
}

func TestNothingKeysToAnEmptyDigest(t *testing.T) {
	k, _ := newKeyer([]byte(strings.Repeat("s", 32)))
	if _, err := k.Key("organization_operator", "  "); err == nil {
		t.Fatal("an empty operator was keyed")
	}
	if _, err := k.Key("", "operator-1"); err == nil {
		t.Fatal("an empty namespace was keyed")
	}
}
