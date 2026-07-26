package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Keyer struct {
	secret []byte
}

func New(secret []byte) (*Keyer, error) {
	if len(secret) < 32 {
		return nil, errors.New("seed decline privacy secret must be at least 32 bytes")
	}
	return &Keyer{secret: append([]byte(nil), secret...)}, nil
}

func (keyer *Keyer) Key(namespace, value string) (string, error) {
	namespace, value = strings.TrimSpace(namespace), strings.TrimSpace(value)
	if namespace == "" || value == "" {
		return "", errors.New("namespace and value are required")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte("seed-decline:v1:" + namespace + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
