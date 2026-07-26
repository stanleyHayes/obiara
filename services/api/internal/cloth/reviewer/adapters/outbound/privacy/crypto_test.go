package privacy

import (
	"encoding/base64"
	"testing"
)

func TestTokenIsRandom256BitsAndHMACIsDeterministic(t *testing.T) {
	source := TokenSource{}
	first, err := source.Token()
	if err != nil {
		t.Fatal(err)
	}
	second, err := source.Token()
	if err != nil || first == second {
		t.Fatalf("tokens are not independently random: %q %q", first, second)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil || len(decoded) != 32 {
		t.Fatalf("token bytes=%d error=%v", len(decoded), err)
	}
	keyer, err := NewHMAC([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := keyer.Key("invite", first)
	b, _ := keyer.Key("invite", first)
	c, _ := keyer.Key("session", first)
	if a != b || a == c || len(a) != 64 {
		t.Fatal("namespaced HMAC contract failed")
	}
}
