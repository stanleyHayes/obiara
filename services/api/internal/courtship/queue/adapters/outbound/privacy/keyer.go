package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Keyer struct{ secret []byte }

func New(secret []byte) (*Keyer, error) {
	if len(secret) < 32 {
		return nil, errors.New("queue privacy secret must be at least 32 bytes")
	}
	return &Keyer{append([]byte(nil), secret...)}, nil
}
func (k *Keyer) Key(namespace, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("value required")
	}
	mac := hmac.New(sha256.New, k.secret)
	_, _ = mac.Write([]byte("courtship-queue:v1:" + namespace + ":" + value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
