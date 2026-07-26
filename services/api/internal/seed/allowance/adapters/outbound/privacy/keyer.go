package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Keyer struct{ secret []byte }

func NewKeyer(secret []byte) (*Keyer, error) {
	if len(secret) < 32 {
		return nil, errors.New("allowance subject key secret must be at least 32 bytes")
	}
	return &Keyer{secret: append([]byte(nil), secret...)}, nil
}

func (k *Keyer) Key(subjectID string) (string, error) {
	if subjectID == "" {
		return "", errors.New("subject id is required")
	}
	mac := hmac.New(sha256.New, k.secret)
	_, _ = mac.Write([]byte("seed-allowance:v1:" + subjectID))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
