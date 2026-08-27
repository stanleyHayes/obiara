// Package privacy derives the non-reversible key a Ghana Card is recognised
// by.
//
// The verification context's contract is that raw card artifacts never
// persist, but "one active account per verified identity" still needs to
// recognise the same card twice. An HMAC satisfies both: it is stable enough
// to compare and index, and it cannot be turned back into a national ID by
// anyone who reads the database.
package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// ErrSecretTooShort reports a key that would not carry enough entropy to
// resist a dictionary attack over a national ID's small, structured space.
var ErrSecretTooShort = errors.New("verification HMAC secret must be at least 32 bytes")

// HMACKeyer keys card numbers.
type HMACKeyer struct {
	secret []byte
}

// NewHMACKeyer copies the secret so a caller cannot mutate it afterwards.
func NewHMACKeyer(secret []byte) (HMACKeyer, error) {
	if len(secret) < 32 {
		return HMACKeyer{}, ErrSecretTooShort
	}
	return HMACKeyer{secret: append([]byte(nil), secret...)}, nil
}

// Key derives the stable identifier for a card number.
//
// Case and surrounding whitespace are normalised first: a card typed with a
// stray space would otherwise key differently and let one identity claim two
// accounts, which is the whole thing this prevents.
func (keyer HMACKeyer) Key(value string) (string, error) {
	normalised := strings.ToUpper(strings.TrimSpace(value))
	if normalised == "" {
		return "", errors.New("ghana card number is required")
	}
	mac := hmac.New(sha256.New, keyer.secret)
	_, _ = mac.Write([]byte(normalised))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
