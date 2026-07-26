package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

var ErrInvalidToken = errors.New("invalid opaque circle invite")

type Issuer struct{}

func NewIssuer() Issuer { return Issuer{} }

func (Issuer) NewToken() (string, string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(secret)
	return raw, digest(raw), nil
}

func (Issuer) Digest(raw string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) != 32 {
		return "", ErrInvalidToken
	}
	return digest(raw), nil
}

func digest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
