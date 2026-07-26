package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Token lifetimes follow agent_plan.md §11: short-lived access tokens and
// rotated refresh sessions.
const (
	AccessTokenLifetime  = 15 * time.Minute
	RefreshTokenLifetime = 30 * 24 * time.Hour

	secretBytes = 32
)

// IssuedToken pairs the plaintext shown once to the client with the hash
// that is stored. Plaintext tokens never touch persistence, logs or
// analytics.
type IssuedToken struct {
	Plain     string
	Hash      string
	ExpiresAt time.Time
}

// IssueAccessToken mints an opaque access token bound to a session ID.
func IssueAccessToken(sessionID string, now time.Time) (IssuedToken, error) {
	return issue(sessionID, now, AccessTokenLifetime)
}

// IssueRefreshToken mints an opaque refresh token bound to a session ID.
func IssueRefreshToken(sessionID string, now time.Time) (IssuedToken, error) {
	return issue(sessionID, now, RefreshTokenLifetime)
}

func issue(sessionID string, now time.Time, lifetime time.Duration) (IssuedToken, error) {
	secret := make([]byte, secretBytes)
	if _, err := rand.Read(secret); err != nil {
		return IssuedToken{}, fmt.Errorf("mint token secret: %w", err)
	}
	plain := sessionID + "." + base64.RawURLEncoding.EncodeToString(secret)
	return IssuedToken{
		Plain:     plain,
		Hash:      HashToken(plain),
		ExpiresAt: now.Add(lifetime).UTC(),
	}, nil
}

// HashToken derives the stored form of a plaintext token.
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

var ErrTokenMalformed = errors.New("token is malformed")

// SplitToken separates the session ID from the secret. Lookup by ID keeps
// repository queries index-friendly; the secret is only ever hashed.
func SplitToken(plain string) (sessionID string, err error) {
	id, _, found := strings.Cut(plain, ".")
	if !found || id == "" {
		return "", ErrTokenMalformed
	}
	return id, nil
}
