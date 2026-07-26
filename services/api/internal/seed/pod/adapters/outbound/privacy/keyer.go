package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Keyer struct{ secret []byte }

func NewKeyer(s []byte) (*Keyer, error) {
	if len(s) < 32 {
		return nil, errors.New("seed pod secret too short")
	}
	return &Keyer{append([]byte(nil), s...)}, nil
}
func (k *Keyer) Key(n, v string) (string, error) {
	if k == nil || len(k.secret) < 32 || strings.TrimSpace(n) == "" || strings.TrimSpace(v) == "" {
		return "", errors.New("invalid seed pod privacy input")
	}
	m := hmac.New(sha256.New, k.secret)
	m.Write([]byte(n))
	m.Write([]byte{0})
	m.Write([]byte(strings.TrimSpace(v)))
	return hex.EncodeToString(m.Sum(nil)), nil
}
