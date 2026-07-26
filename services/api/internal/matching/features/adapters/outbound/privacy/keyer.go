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
		return Keyer{}, errors.New("matching feature secret too short")
	}
	return Keyer{secret: append([]byte(nil), secret...)}, nil
}
func (k Keyer) Key(namespace, value string) (string, error) {
	if namespace == "" || value == "" {
		return "", errors.New("matching feature hmac input empty")
	}
	mac := hmac.New(sha256.New, k.secret)
	_, _ = mac.Write([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
