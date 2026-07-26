package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrKeyTooShort = errors.New("vouch attestation HMAC key must be at least 32 bytes")

type HMACKeyer struct{ secret []byte }

func NewHMACKeyer(secret []byte) (*HMACKeyer, error) {
	if len(secret) < 32 {
		return nil, ErrKeyTooShort
	}
	return &HMACKeyer{secret: append([]byte(nil), secret...)}, nil
}
func (k *HMACKeyer) Key(namespace, value string) (string, error) {
	mac := hmac.New(sha256.New, k.secret)
	_, _ = mac.Write([]byte(namespace))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
