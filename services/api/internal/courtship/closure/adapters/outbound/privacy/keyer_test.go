package privacy

import (
	"strings"
	"testing"
)

func TestOpaqueScopedDeterministicKeys(t *testing.T) {
	k, _ := NewKeyer([]byte(strings.Repeat("x", 32)))
	a, _ := k.Key("closure_room", "raw")
	b, _ := k.Key("closure_room", "raw")
	c, _ := k.Key("closure_actor", "raw")
	if len(a) != 64 || a != b || a == c || strings.Contains(a, "raw") {
		t.Fatal("unsafe key")
	}
}
