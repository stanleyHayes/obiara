package privacy

import "testing"

func TestKeyerIsPurposeSeparated(t *testing.T) {
	k, e := New([]byte("01234567890123456789012345678901"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("member", "raw")
	b, _ := k.Key("reference", "raw")
	if a == b || a == "raw" || len(a) != 64 {
		t.Fatal("privacy key failure")
	}
}
