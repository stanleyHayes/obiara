package privacy

import "testing"

func TestKeysHideMemberAndPaymentData(t *testing.T) {
	k, e := New([]byte("01234567890123456789012345678901"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("account", "member@example.invalid")
	b, _ := k.Key("payment", "member@example.invalid")
	if a == b || a == "member@example.invalid" || len(a) != 64 {
		t.Fatal("privacy failure")
	}
}
