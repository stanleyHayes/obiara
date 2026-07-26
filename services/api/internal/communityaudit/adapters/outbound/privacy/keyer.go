package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Keyer struct{ s []byte }

func New(s []byte) (Keyer, error) {
	if len(s) < 32 {
		return Keyer{}, errors.New("secret too short")
	}
	return Keyer{append([]byte(nil), s...)}, nil
}
func (k Keyer) Key(n, v string) (string, error) {
	if strings.TrimSpace(v) == "" {
		return "", errors.New("input required")
	}
	m := hmac.New(sha256.New, k.s)
	m.Write([]byte(n + "\x00" + v))
	return hex.EncodeToString(m.Sum(nil)), nil
}
