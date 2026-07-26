package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type HMACKeyer struct{ secret []byte }

func NewHMACKeyer(secret []byte) (HMACKeyer, error) {
	if len(secret) < 32 {
		return HMACKeyer{}, errors.New("host HMAC secret must be at least 32 bytes")
	}
	return HMACKeyer{secret: append([]byte(nil), secret...)}, nil
}

func (keyer HMACKeyer) Key(namespace, value string) (string, error) {
	namespace, value = strings.TrimSpace(namespace), strings.TrimSpace(value)
	if namespace == "" || value == "" {
		return "", errors.New("host privacy key input is required")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
