package privacy

import (
	"strings"
	"testing"
)

func TestOpaqueScopedKeys(t *testing.T) {
	k, _ := NewKeyer([]byte(strings.Repeat("x", 32)))
	a, _ := k.Key("safety_room", "raw")
	b, _ := k.Key("safety_actor", "raw")
	if len(a) != 64 || a == b || strings.Contains(a, "raw") {
		t.Fatal("unsafe key")
	}
}
