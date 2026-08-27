package privacy

import (
	"strings"
	"testing"
)

var testSecret = []byte("test-secret-with-32-bytes-123456")

func TestSecretValidation(t *testing.T) {
	if _, err := NewHMACKeyer([]byte("short")); err != ErrSecretTooShort {
		t.Fatalf("short secret = %v, want ErrSecretTooShort", err)
	}
	if _, err := NewHMACKeyer([]byte("31-byte-secret-that-is-too-shrt")); err != ErrSecretTooShort {
		t.Fatalf("31-byte secret = %v, want ErrSecretTooShort", err)
	}
	keyer, err := NewHMACKeyer(testSecret)
	if err != nil {
		t.Fatalf("32-byte secret = %v", err)
	}
	if keyer.secret == nil {
		t.Fatal("secret not copied")
	}
}

func TestCardKeyStability(t *testing.T) {
	keyer, _ := NewHMACKeyer(testSecret)
	key1, err1 := keyer.Key("GHA-123456789-0")
	if err1 != nil {
		t.Fatalf("first key = %v", err1)
	}
	key2, err2 := keyer.Key("GHA-123456789-0")
	if err2 != nil {
		t.Fatalf("second key = %v", err2)
	}
	if key1 != key2 {
		t.Fatalf("same card produces different keys: %s vs %s", key1, key2)
	}

	keyDifferent, _ := keyer.Key("GHA-123456789-1")
	if key1 == keyDifferent {
		t.Fatalf("different cards produce same key: %s", key1)
	}
}

func TestCardKeyCaseInsensitive(t *testing.T) {
	keyer, _ := NewHMACKeyer(testSecret)
	keyLower, _ := keyer.Key("gha-123456789-0")
	keyUpper, _ := keyer.Key("GHA-123456789-0")
	keyMixed, _ := keyer.Key("GhA-123456789-0")

	if keyLower != keyUpper {
		t.Fatalf("case affects key: %s vs %s", keyLower, keyUpper)
	}
	if keyUpper != keyMixed {
		t.Fatalf("case affects key: %s vs %s", keyUpper, keyMixed)
	}
}

func TestCardKeyWhitespaceInsensitive(t *testing.T) {
	keyer, _ := NewHMACKeyer(testSecret)
	keyNoSpace, _ := keyer.Key("GHA-123456789-0")
	keyLeadingSpace, _ := keyer.Key(" GHA-123456789-0")
	keyTrailingSpace, _ := keyer.Key("GHA-123456789-0 ")
	keyBothSpaces, _ := keyer.Key("  GHA-123456789-0  ")

	if keyNoSpace != keyLeadingSpace {
		t.Fatalf("leading space affects key: %s vs %s", keyNoSpace, keyLeadingSpace)
	}
	if keyNoSpace != keyTrailingSpace {
		t.Fatalf("trailing space affects key: %s vs %s", keyNoSpace, keyTrailingSpace)
	}
	if keyNoSpace != keyBothSpaces {
		t.Fatalf("both spaces affect key: %s vs %s", keyNoSpace, keyBothSpaces)
	}
}

func TestCardKeyNotInput(t *testing.T) {
	keyer, _ := NewHMACKeyer(testSecret)
	card := "GHA-123456789-0"
	key, _ := keyer.Key(card)

	if strings.Contains(key, card) {
		t.Fatalf("key contains card number: %s in %s", card, key)
	}
	if strings.Contains(key, strings.ToLower(card)) {
		t.Fatalf("key contains card number (lowercase): %s in %s", card, key)
	}
	if strings.Contains(key, "GHA") {
		t.Fatalf("key contains card prefix: GHA in %s", key)
	}
}

func TestCardKeyEmptyInput(t *testing.T) {
	keyer, _ := NewHMACKeyer(testSecret)
	if _, err := keyer.Key(""); err == nil {
		t.Fatal("empty card must error")
	}
	if _, err := keyer.Key("   "); err == nil {
		t.Fatal("whitespace-only card must error")
	}
}
