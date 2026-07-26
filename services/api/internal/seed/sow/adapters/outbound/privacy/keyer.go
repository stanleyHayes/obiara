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
		return nil, errors.New("sow privacy secret must be at least 32 bytes")
	}
	return &Keyer{append([]byte(nil), secret...)}, nil
}
func (k *Keyer) Key(namespace, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("value required")
	}
	prefix := "seed-sow:v1:" + namespace + ":"
	if namespace == "allowance-subject" {
		prefix = "seed-allowance:v1:"
	}
	m := hmac.New(sha256.New, k.secret)
	_, _ = m.Write([]byte(prefix + value))
	return hex.EncodeToString(m.Sum(nil)), nil
}
