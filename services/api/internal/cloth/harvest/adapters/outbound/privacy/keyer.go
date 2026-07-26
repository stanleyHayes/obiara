package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Keyer struct{ secret []byte }

func NewKeyer(secret []byte) (Keyer, error) {
	if len(secret) < 32 {
		return Keyer{}, errors.New("harvest hmac secret must be at least 32 bytes")
	}
	return Keyer{append([]byte(nil), secret...)}, nil
}
func (k Keyer) Key(namespace, value string) (string, error) {
	if namespace == "" || value == "" {
		return "", errors.New("harvest hmac input is empty")
	}
	m := hmac.New(sha256.New, k.secret)
	_, _ = m.Write([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(m.Sum(nil)), nil
}
