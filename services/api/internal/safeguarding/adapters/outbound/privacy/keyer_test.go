package privacy

import (
	"strings"
	"testing"
)

func TestHMACKeyerIsDeterministicAndKeyed(t *testing.T) {
	first, err := NewHMACKeyer([]byte(strings.Repeat("a", 32)))
	if err != nil {
		t.Fatal(err)
	}
	second, _ := NewHMACKeyer([]byte(strings.Repeat("b", 32)))
	firstKey, _ := first.Key("member:1")
	replayKey, _ := first.Key("member:1")
	secondKey, _ := second.Key("member:1")
	if len(firstKey) != 64 || firstKey != replayKey || firstKey == secondKey ||
		strings.Contains(firstKey, "member") {
		t.Fatal("HMAC key must be stable, secret-dependent, and non-reversible")
	}
}

func TestHMACKeyerRejectsWeakSecretAndEmptyInput(t *testing.T) {
	if _, err := NewHMACKeyer([]byte("short")); err == nil {
		t.Fatal("weak HMAC secret was accepted")
	}
	keyer, _ := NewHMACKeyer([]byte(strings.Repeat("a", 32)))
	if _, err := keyer.Key(" "); err == nil {
		t.Fatal("empty key input was accepted")
	}
}
