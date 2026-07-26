package privacy

import "testing"

func TestKeyerIsStableNamespacedAndOpaque(t *testing.T) {
	keyer, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	first, _ := keyer.Key("courtship-drum:member", "private-member")
	replay, _ := keyer.Key("courtship-drum:member", "private-member")
	content, _ := keyer.Key("courtship-drum:voice", "private-member")
	if first != replay || first == content || len(first) != 64 || first == "private-member" {
		t.Fatalf("first=%q replay=%q content=%q", first, replay, content)
	}
}
