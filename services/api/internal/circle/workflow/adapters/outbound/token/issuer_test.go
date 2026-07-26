package token

import (
	"strings"
	"testing"
)

func TestTokenIsOpaqueAndOnlyDigestIsStable(t *testing.T) {
	issuer := NewIssuer()
	raw, stored, err := issuer.NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw == stored || strings.Contains(stored, raw) || len(raw) < 40 {
		t.Fatalf("token was not opaque: raw=%q stored=%q", raw, stored)
	}
	again, err := issuer.Digest(raw)
	if err != nil || again != stored {
		t.Fatalf("digest=%q err=%v, want %q", again, err, stored)
	}
	if _, err := issuer.Digest("guessable"); err != ErrInvalidToken {
		t.Fatalf("weak token accepted: %v", err)
	}
}
