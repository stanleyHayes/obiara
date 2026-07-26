package privacy

import (
	"strings"
	"testing"
)

func TestHMACKeyerIsNamespacedAndDoesNotExposeInput(t *testing.T) {
	keyer, err := NewHMACKeyer([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	first, _ := keyer.Key("device", "raw-device-id")
	again, _ := keyer.Key("device", "raw-device-id")
	other, _ := keyer.Key("subject", "raw-device-id")
	if first != again || first == other || strings.Contains(first, "raw-device-id") {
		t.Fatalf("unsafe or unstable keys: %q %q %q", first, again, other)
	}
}
