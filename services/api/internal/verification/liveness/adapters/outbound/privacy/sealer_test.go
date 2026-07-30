package privacy

import (
	"bytes"
	"strings"
	"testing"
)

func TestAESGCMSealerDoesNotRetainPlaintextAndRandomizesNonce(t *testing.T) {
	sealer, err := NewAESGCMSealer([]byte(strings.Repeat("s", 32)))
	if err != nil {
		t.Fatal(err)
	}
	first, firstNonce, err := sealer.Seal([]byte("sensitive-face"))
	if err != nil {
		t.Fatal(err)
	}
	second, secondNonce, err := sealer.Seal([]byte("sensitive-face"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(first, []byte("sensitive-face")) || bytes.Equal(first, second) ||
		bytes.Equal(firstNonce, secondNonce) {
		t.Fatal("ciphertext retained plaintext or reused a nonce")
	}
}
