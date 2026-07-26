package privacy

import "testing"

func TestKeyerHidesRawMemberAndSeparatesPurpose(t *testing.T) {
	k, e := New([]byte("01234567890123456789012345678901"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("subject", "member@example.invalid")
	b, _ := k.Key("hold", "member@example.invalid")
	if a == b || a == "member@example.invalid" || len(a) != 64 {
		t.Fatal("privacy key failure")
	}
}
