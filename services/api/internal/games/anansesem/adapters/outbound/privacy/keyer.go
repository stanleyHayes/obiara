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
		return Keyer{}, errors.New("anansesem secret too short")
	}
	return Keyer{append([]byte(nil), secret...)}, nil
}
func (k Keyer) Key(ns, v string) (string, error) {
	if ns == "" || v == "" {
		return "", errors.New("anansesem hmac input empty")
	}
	m := hmac.New(sha256.New, k.secret)
	_, _ = m.Write([]byte(ns + "\x00" + v))
	return hex.EncodeToString(m.Sum(nil)), nil
}
