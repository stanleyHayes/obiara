package privacy

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

type HMAC struct {
	secret []byte
}

func NewHMAC(secret []byte) (HMAC, error) {
	if len(secret) < 32 {
		return HMAC{}, errors.New("reviewer hmac secret must be at least 32 bytes")
	}
	return HMAC{secret: append([]byte(nil), secret...)}, nil
}

func (keyer HMAC) Key(namespace, value string) (string, error) {
	if namespace == "" || value == "" {
		return "", errors.New("reviewer hmac input is empty")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte(namespace + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (keyer HMAC) Ref(reviewerKey, reviewID string) (string, error) {
	return keyer.Key("cloth-reviewer:watermark", reviewerKey+"\x00"+reviewID)
}

type TokenSource struct{}

func (TokenSource) Token() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
