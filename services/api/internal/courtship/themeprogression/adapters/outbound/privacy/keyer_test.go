package privacy

import "testing"

func TestKeyerIsStableNamespacedAndOpaque(t *testing.T) {
	keyer, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	member, _ := keyer.Key("courtship-themeprogression:member", "private")
	replay, _ := keyer.Key("courtship-themeprogression:member", "private")
	content, _ := keyer.Key("courtship-themeprogression:encrypted-content", "private")
	if member != replay || member == content || len(member) != 64 || member == "private" {
		t.Fatalf("member=%q replay=%q content=%q", member, replay, content)
	}
}
