package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

type AESGCMSealer struct {
	aead cipher.AEAD
}

func NewAESGCMSealer(secret []byte) (AESGCMSealer, error) {
	if len(secret) < 32 {
		return AESGCMSealer{}, fmt.Errorf("liveness encryption secret must be at least 32 bytes")
	}
	key := sha256.Sum256(append([]byte("obiara:liveness-artifact:v1:"), secret...))
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
