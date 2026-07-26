package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Keyer struct{ secret []byte }

func NewKeyer(s []byte) (*Keyer, error) {
	if len(s) < 32 {
		return nil, errors.New("lifecycle key too short")
	}
	return &Keyer{append([]byte(nil), s...)}, nil
}
func (k *Keyer) Key(scope, value string) (string, error) {
	if k == nil || scope == "" || value == "" {
		return "", errors.New("invalid lifecycle key")
	}
	h := hmac.New(sha256.New, k.secret)
	h.Write([]byte(scope))
	h.Write([]byte{0})
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil)), nil
}
