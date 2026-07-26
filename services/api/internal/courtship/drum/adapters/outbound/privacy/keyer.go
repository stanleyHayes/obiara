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
		return nil, errors.New("courtship drum privacy secret too short")
	}
	return &Keyer{secret: append([]byte(nil), secret...)}, nil
}
func (keyer *Keyer) Key(namespace, value string) (string, error) {
	namespace, value = strings.TrimSpace(namespace), strings.TrimSpace(value)
	if keyer == nil || len(keyer.secret) < 32 || namespace == "" || value == "" {
		return "", errors.New("invalid courtship drum privacy input")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
