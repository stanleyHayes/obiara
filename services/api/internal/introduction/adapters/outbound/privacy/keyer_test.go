package privacy

import (
	"strings"
	"testing"
)

func TestHMACKeyerProducesStableNonRawFingerprints(t *testing.T) {
	first, err := NewHMACKeyer([]byte(strings.Repeat("a", 32)))
	if err != nil {
		t.Fatal(err)
	}
	second, _ := NewHMACKeyer([]byte(strings.Repeat("b", 32)))
	key, _ := first.Key("member:1|asset:1")
	replay, _ := first.Key("member:1|asset:1")
	other, _ := second.Key("member:1|asset:1")
	if len(key) != 64 || key != replay || key == other || strings.Contains(key, "member") {
		t.Fatal("fingerprint is not stable, keyed, and privacy safe")
	}
}
