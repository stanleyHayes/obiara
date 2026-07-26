package privacy

import "testing"

func TestHMAC(t *testing.T) {
	k, e := NewKeyer([]byte("0123456789abcdef0123456789abcdef"))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := k.Key("room", "raw")
	b, _ := k.Key("room", "raw")
	c, _ := k.Key("player", "raw")
	if a != b || a == c || len(a) != 64 {
		t.Fatal("hmac")
	}
}
