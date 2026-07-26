package privacy

import "testing"

func TestKeyerSeparatesNamespaceAndNeverReturnsRawValue(t *testing.T) {
	k, err := NewKeyer([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := k.Key("a", "member")
	b, _ := k.Key("b", "member")
	if a == b || a == "member" || len(a) != 64 {
		t.Fatal("invalid privacy key")
	}
}
