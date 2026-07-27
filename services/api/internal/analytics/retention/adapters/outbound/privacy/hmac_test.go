package privacy

import (
	"testing"
)

func TestVersionedKeysBreakCrossEpochLinkability(t *testing.T) {
	p := New(map[uint64][]byte{1: []byte("11111111111111111111111111111111"), 2: []byte("22222222222222222222222222222222")})
	a, e := p.Derive("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 1)
	if e != nil {
		t.Fatal(e)
	}
	b, e := p.Derive("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 2)
	if e != nil || a == b || len(a) != 64 || len(b) != 64 {
		t.Fatal(a, b, e)
	}
}
