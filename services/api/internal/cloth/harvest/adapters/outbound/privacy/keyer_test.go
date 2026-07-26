package privacy

import "testing"

func TestNamespacedDeterministicHMAC(t *testing.T) {
	k, e := NewKeyer([]byte("0123456789abcdef0123456789abcdef"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("member", "private")
	b, _ := k.Key("member", "private")
	c, _ := k.Key("delivery", "private")
	if a != b || a == c || len(a) != 64 {
		t.Fatal("HMAC contract")
	}
}
