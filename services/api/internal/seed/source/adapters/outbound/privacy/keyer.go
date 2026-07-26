package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Keyer struct{ secret []byte }

func NewKeyer(secret []byte) (*Keyer, error) {
	if len(secret) < 32 {
		return nil, errors.New("seed source privacy secret must be at least 32 bytes")
	}
	return &Keyer{secret: append([]byte(nil), secret...)}, nil
}
func (k *Keyer) Key(namespace, value string) (string, error) {
	if k == nil || len(k.secret) < 32 || strings.TrimSpace(namespace) == "" || strings.TrimSpace(value) == "" {
		return "", errors.New("invalid seed source privacy input")
	}
	mac := hmac.New(sha256.New, k.secret)
	_, _ = mac.Write([]byte(namespace))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
