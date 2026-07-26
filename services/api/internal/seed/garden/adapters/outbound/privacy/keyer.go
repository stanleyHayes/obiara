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
		return nil, errors.New("garden privacy key must be at least 32 bytes")
	}
	return &Keyer{secret: append([]byte(nil), secret...)}, nil
}

func (keyer *Keyer) Key(scope, value string) (string, error) {
	if keyer == nil || scope == "" || value == "" {
		return "", errors.New("invalid garden key input")
	}
	hash := hmac.New(sha256.New, keyer.secret)
	hash.Write([]byte(scope))
	hash.Write([]byte{0})
	hash.Write([]byte(value))
	return hex.EncodeToString(hash.Sum(nil)), nil
}
