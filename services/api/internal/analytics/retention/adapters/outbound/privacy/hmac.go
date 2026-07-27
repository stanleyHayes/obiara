package privacy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type Pseudonymizer struct{ keys map[uint64][]byte }

func New(keys map[uint64][]byte) Pseudonymizer {
	copyKeys := map[uint64][]byte{}
	for version, key := range keys {
		copyKeys[version] = append([]byte(nil), key...)
	}
	return Pseudonymizer{copyKeys}
}
func (p Pseudonymizer) Derive(current string, version uint64) (string, error) {
	key := p.keys[version]
	if len(key) < 32 {
		return "", errors.New("missing pseudonym key")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(fmt.Sprintf("analytics-retention:v%d:%s", version, current)))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
