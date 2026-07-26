package privacy

import "testing"

func TestKeyerProducesStableNamespacedOpaqueReferences(t *testing.T) {
	keyer, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	first, _ := keyer.Key("courtship-theme:member", "private")
	replay, _ := keyer.Key("courtship-theme:member", "private")
	content, _ := keyer.Key("courtship-theme:encrypted-content", "private")
	if first != replay || first == content || len(first) != 64 || first == "private" {
		t.Fatalf("first=%q replay=%q content=%q", first, replay, content)
	}
}
