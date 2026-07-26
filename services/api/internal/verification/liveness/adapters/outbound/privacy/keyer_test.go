package privacy

import (
	"strings"
	"testing"
)

func TestHMACKeyerProducesStableSecretDependentProof(t *testing.T) {
	first, _ := NewHMACKeyer([]byte(strings.Repeat("a", 32)))
	second, _ := NewHMACKeyer([]byte(strings.Repeat("b", 32)))
	key, err := first.Key("artifact:voice")
	replay, _ := first.Key("artifact:voice")
	other, _ := second.Key("artifact:voice")
	if err != nil || len(key) != 64 || key != replay || key == other ||
		strings.Contains(key, "artifact") {
		t.Fatal("invalid privacy key behavior")
	}
}
