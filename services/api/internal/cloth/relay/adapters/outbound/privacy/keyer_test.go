package privacy

import (
	"strings"
	"testing"
)

func TestOpaqueScoped(t *testing.T) {
	k, _ := NewKeyer([]byte(strings.Repeat("x", 32)))
	a, _ := k.Key("relay_pair", "raw")
	b, _ := k.Key("relay_member", "raw")
	if len(a) != 64 || a == b || strings.Contains(a, "raw") {
		t.Fatal("unsafe")
	}
}
