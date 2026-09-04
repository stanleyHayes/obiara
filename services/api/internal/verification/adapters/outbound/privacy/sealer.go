package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// AESGCMSealer encrypts Ghana Card photographs at the application boundary.
//
// The key is derived with a purpose-specific label, so a card image can never
// be opened with the liveness key and vice versa: one secret leaking does not
// unlock the other store.
type AESGCMSealer struct {
	aead cipher.AEAD
}

func NewAESGCMSealer(secret []byte) (AESGCMSealer, error) {
	if len(secret) < 32 {
		return AESGCMSealer{}, fmt.Errorf("verification encryption secret must be at least 32 bytes")
	}
	key := sha256.Sum256(append([]byte("obiara:identity-document:v1:"), secret...))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return AESGCMSealer{}, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return AESGCMSealer{}, err
	}
	return AESGCMSealer{aead: aead}, nil
}

func (sealer AESGCMSealer) Seal(plaintext []byte) ([]byte, []byte, error) {
	nonce := make([]byte, sealer.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	return sealer.aead.Seal(nil, nonce, plaintext, nil), nonce, nil
}

// Open reverses Seal for a reviewer who is allowed to see the image.
func (sealer AESGCMSealer) Open(ciphertext, nonce []byte) ([]byte, error) {
	if len(nonce) != sealer.aead.NonceSize() {
		return nil, fmt.Errorf("identity document nonce is the wrong size")
	}
	return sealer.aead.Open(nil, nonce, ciphertext, nil)
}
