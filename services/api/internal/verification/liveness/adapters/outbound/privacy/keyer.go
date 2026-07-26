package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type HMACKeyer struct {
	secret []byte
}

func NewHMACKeyer(secret []byte) (HMACKeyer, error) {
	if len(secret) < 32 {
		return HMACKeyer{}, errors.New("liveness HMAC secret must be at least 32 bytes")
	}
	return HMACKeyer{secret: append([]byte(nil), secret...)}, nil
}

func (keyer HMACKeyer) Key(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", errors.New("liveness key input is required")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
