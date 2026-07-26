package privacy

import "testing"

func TestKeyerIsStableNamespacedAndDoesNotExposeInput(t *testing.T) {
	keyer, err := New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	first, _ := keyer.Key("decliner", "member-private")
	replay, _ := keyer.Key("decliner", "member-private")
	other, _ := keyer.Key("seed", "member-private")
	if first != replay || first == other || len(first) != 64 || first == "member-private" {
		t.Fatalf("first=%q replay=%q other=%q", first, replay, other)
	}
}
