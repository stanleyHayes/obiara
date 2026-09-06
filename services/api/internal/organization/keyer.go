package organization

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// keyer digests an operator id under a namespace.
//
// The same shape as every other privacy keyer in the codebase: one secret per
// context, so a digest here cannot be compared with a digest anywhere else,
// and the relationships between contexts stay illegible at rest.
type keyer struct{ secret []byte }

func newKeyer(secret []byte) (keyer, error) {
	if len(secret) < 32 {
		return keyer{}, errors.New("organization keying secret must be at least 32 bytes")
	}
	return keyer{secret: append([]byte(nil), secret...)}, nil
}

func (k keyer) Key(namespace, value string) (string, error) {
	namespace, value = strings.TrimSpace(namespace), strings.TrimSpace(value)
	if namespace == "" || value == "" {
		return "", errors.New("invalid organization keying input")
	}
	mac := hmac.New(sha256.New, k.secret)
	mac.Write([]byte(namespace))
	mac.Write([]byte{0})
	mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
