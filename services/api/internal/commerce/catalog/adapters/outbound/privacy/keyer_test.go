package privacy

import "testing"

func TestTitleRefIsOpaqueAndPurposeSeparated(t *testing.T) {
	k, e := New([]byte("01234567890123456789012345678901"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("title", "product name")
	b, _ := k.Key("other", "product name")
	if a == b || a == "product name" || len(a) != 64 {
		t.Fatal("key failure")
	}
}
