package privacy

import "testing"

func TestHMAC(t *testing.T) {
	k, e := NewKeyer([]byte("0123456789abcdef0123456789abcdef"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("fire", "raw")
	b, _ := k.Key("fire", "raw")
	c, _ := k.Key("actor", "raw")
	if a != b || a == c || len(a) != 64 {
		t.Fatal("hmac")
	}
}
