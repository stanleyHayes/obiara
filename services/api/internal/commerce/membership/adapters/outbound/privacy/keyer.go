package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

type Keyer struct{ secret []byte }

func New(secret []byte) (Keyer, error) {
	if len(secret) < 32 {
		return Keyer{}, errors.New("commerce HMAC secret must be at least 32 bytes")
	}
	return Keyer{secret: append([]byte(nil), secret...)}, nil
}

func (keyer Keyer) MemberKey(memberID string) (string, error) {
	memberID = strings.TrimSpace(memberID)
	if memberID == "" {
		return "", errors.New("member id is required")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte("membership:member:" + memberID))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
