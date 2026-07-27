package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var ErrInvalid = errors.New("invalid reconciliation key input")

type Keyer struct{ secret []byte }

func NewKeyer(secret []byte) (Keyer, error) {
	if len(secret) < 32 {
		return Keyer{}, ErrInvalid
	}
	return Keyer{secret: append([]byte(nil), secret...)}, nil
}
func (k Keyer) Key(namespace, value string) (string, error) {
	namespace, value = strings.TrimSpace(namespace), strings.TrimSpace(value)
	if namespace == "" || value == "" || len(namespace) > 256 || len(value) > 512 {
		return "", ErrInvalid
	}
	m := hmac.New(sha256.New, k.secret)
	m.Write([]byte(namespace))
	m.Write([]byte{0})
	m.Write([]byte(value))
	return hex.EncodeToString(m.Sum(nil)), nil
}
