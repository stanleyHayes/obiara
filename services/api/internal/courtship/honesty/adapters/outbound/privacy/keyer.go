package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Keyer struct{ secret []byte }

func NewKeyer(secret []byte) (*Keyer, error) {
	if len(secret) < 32 {
		return nil, errors.New("honesty key must be at least 32 bytes")
	}
	return &Keyer{append([]byte(nil), secret...)}, nil
}
func (k *Keyer) Key(scope, value string) (string, error) {
	if k == nil || scope == "" || value == "" {
		return "", errors.New("invalid honesty key input")
	}
	h := hmac.New(sha256.New, k.secret)
	h.Write([]byte(scope))
	h.Write([]byte{0})
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil)), nil
}
