package domain

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// Admin password policy. A password is the "something you know" factor that
// sits in front of the emailed MFA code: without it, control of an
// operator's mailbox is by itself control of the admin console.
//
// Passwords are stored only as argon2id digests in PHC string format. The
// parameters below are the OWASP-recommended second-preset (64 MiB, t=3,
// p=4); they are encoded into every digest so cost can be raised later
// without invalidating existing hashes.
const (
	MinPasswordLength = 12
	MaxPasswordLength = 256

	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

var (
	ErrPasswordTooShort = fmt.Errorf("admin password must be at least %d characters", MinPasswordLength)
	ErrPasswordTooLong  = fmt.Errorf("admin password must be at most %d characters", MaxPasswordLength)
	ErrPasswordWeak     = errors.New("admin password must mix upper case, lower case, and a digit or symbol")
	ErrPasswordHash     = errors.New("admin password hash is malformed")
)

// HashPassword validates a plaintext password against policy and returns its
// argon2id PHC digest. The plaintext is never retained.
func HashPassword(plain string) (string, error) {
	if err := CheckPasswordPolicy(plain); err != nil {
		return "", err
	}
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	digest := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(digest),
	), nil
}

// CheckPasswordPolicy reports whether a plaintext password is acceptable.
func CheckPasswordPolicy(plain string) error {
	// Length is counted in runes so a passphrase in a non-Latin script is
	// not penalised for its UTF-8 byte width.
	length := len([]rune(plain))
	if length < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if length > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	var hasUpper, hasLower, hasOther bool
	for _, r := range plain {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasOther = true
		}
	}
	if !hasUpper || !hasLower || !hasOther {
		return ErrPasswordWeak
	}
	return nil
}

// VerifyPassword reports whether plain matches the stored PHC digest. It
// compares in constant time and re-derives with the digest's own
// parameters, so digests written under an older cost still verify.
//
// A malformed or empty digest returns false rather than an error: callers
// are authentication paths that must not branch differently on storage
// corruption than on a wrong password.
func VerifyPassword(encoded, plain string) bool {
	memory, time, threads, salt, digest, err := parsePHC(encoded)
	if err != nil {
		return false
	}
	candidate := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(len(digest)))
	return subtle.ConstantTimeCompare(candidate, digest) == 1
}

func parsePHC(encoded string) (memory, time uint32, threads uint8, salt, digest []byte, err error) {
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, digest
	if len(parts) != 6 || parts[1] != "argon2id" {
		return 0, 0, 0, nil, nil, ErrPasswordHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return 0, 0, 0, nil, nil, ErrPasswordHash
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return 0, 0, 0, nil, nil, ErrPasswordHash
	}
	if memory == 0 || time == 0 || threads == 0 {
		return 0, 0, 0, nil, nil, ErrPasswordHash
	}
	if salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil || len(salt) == 0 {
		return 0, 0, 0, nil, nil, ErrPasswordHash
	}
	if digest, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil || len(digest) == 0 {
		return 0, 0, 0, nil, nil, ErrPasswordHash
	}
	return memory, time, threads, salt, digest, nil
}
