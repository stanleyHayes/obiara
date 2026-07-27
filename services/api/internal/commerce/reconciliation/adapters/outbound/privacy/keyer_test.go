package privacy

import "testing"

func TestKeyerIsStableScopedAndDoesNotExposeInput(t *testing.T) {
	k, err := NewKeyer([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := k.Key("provider", "raw-payment-reference")
	b, _ := k.Key("provider", "raw-payment-reference")
	c, _ := k.Key("event", "raw-payment-reference")
	if a != b || a == c || a == "raw-payment-reference" || len(a) != 64 {
		t.Fatalf("a=%q b=%q c=%q", a, b, c)
	}
}
