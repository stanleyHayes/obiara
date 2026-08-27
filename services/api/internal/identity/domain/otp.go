package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// OTP policy (agent_plan.md §11; E03-S01): 6-digit codes, 10-minute expiry,
// bounded attempts, and per-contact resend throttling. Codes are stored
// hashed; plaintext codes exist only between minting and sending.
const (
	OtpLifetime    = 10 * time.Minute
	OtpMaxAttempts = 5
	OtpResendAfter = 60 * time.Second
	OtpHourlyLimit = 5
	otpCodeDigits  = 6
)

var (
	ErrInvalidPhone        = errors.New("phone number must be E.164")
	ErrChallengeIDNeeded   = errors.New("otp challenge id is required")
	ErrContactRequired     = errors.New("otp challenge contact is required")
	ErrOtpExpired          = errors.New("otp code is expired")
	ErrOtpMismatch         = errors.New("otp code does not match")
	ErrOtpAttemptsExceeded = errors.New("otp attempts exceeded")
	ErrOtpConsumed         = errors.New("otp challenge already consumed")
	ErrOtpRateLimited      = errors.New("otp resend rate limited")
)

// OtpChallenge is one OTP verification challenge against a contact.
type OtpChallenge struct {
	id         string
	contact    Contact
	codeHash   string
	expiresAt  time.Time
	attempts   int
	sentCount  int
	createdAt  time.Time
	consumedAt *time.Time
}

// NewChallenge validates and builds a challenge for the given plaintext
// code. The code is hashed immediately and never retained.
func NewChallenge(id string, contact Contact, plainCode string, sentCount int, now time.Time) (OtpChallenge, error) {
	if id == "" {
		return OtpChallenge{}, ErrChallengeIDNeeded
	}
	if contact.IsZero() {
		return OtpChallenge{}, ErrContactRequired
	}
	if len(plainCode) != otpCodeDigits {
		return OtpChallenge{}, fmt.Errorf("otp code must be %d digits", otpCodeDigits)
	}
	now = now.UTC()
	return OtpChallenge{
		id:        id,
		contact:   contact,
		codeHash:  HashToken(plainCode),
		expiresAt: now.Add(OtpLifetime),
		sentCount: sentCount,
		createdAt: now,
	}, nil
}

// ReconstituteChallenge rebuilds a stored challenge without policy checks.
func ReconstituteChallenge(id string, contact Contact, codeHash string, expiresAt time.Time, attempts, sentCount int, createdAt time.Time, consumedAt *time.Time) OtpChallenge {
	return OtpChallenge{
		id:         id,
		contact:    contact,
		codeHash:   codeHash,
		expiresAt:  expiresAt,
		attempts:   attempts,
		sentCount:  sentCount,
		createdAt:  createdAt,
		consumedAt: consumedAt,
	}
}

// CheckResend enforces the per-contact resend policy against the most recent
// challenge: one resend per OtpResendAfter and at most OtpHourlyLimit sends
// in the trailing hour (approximated by challenge age + send count).
func CheckResend(latest *OtpChallenge, now time.Time) error {
	if latest == nil {
		return nil
	}
	now = now.UTC()
	if now.Sub(latest.createdAt) < OtpResendAfter {
		return ErrOtpRateLimited
	}
	if now.Sub(latest.createdAt) < time.Hour && latest.sentCount >= OtpHourlyLimit {
		return ErrOtpRateLimited
	}
	return nil
}

// Verify checks a presented code. A failed attempt is recorded on the
// challenge and must be persisted by the caller.
func (challenge *OtpChallenge) Verify(plainCode string, now time.Time) error {
	if challenge.consumedAt != nil {
		return ErrOtpConsumed
	}
	if challenge.attempts >= OtpMaxAttempts {
		return ErrOtpAttemptsExceeded
	}
	if now.UTC().After(challenge.expiresAt) {
		return ErrOtpExpired
	}
	if HashToken(plainCode) != challenge.codeHash {
		challenge.attempts++
		return ErrOtpMismatch
	}
	consumed := now.UTC()
	challenge.consumedAt = &consumed
	return nil
}

func GenerateCode() (string, error) {
	code := make([]byte, otpCodeDigits)
	for i := range code {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate otp code: %w", err)
		}
		code[i] = byte('0' + digit.Int64())
	}
	return string(code), nil
}

func (challenge OtpChallenge) ID() string             { return challenge.id }
func (challenge OtpChallenge) Contact() Contact       { return challenge.contact }
func (challenge OtpChallenge) CodeHash() string       { return challenge.codeHash }
func (challenge OtpChallenge) ExpiresAt() time.Time   { return challenge.expiresAt }
func (challenge OtpChallenge) Attempts() int          { return challenge.attempts }
func (challenge OtpChallenge) SentCount() int         { return challenge.sentCount }
func (challenge OtpChallenge) CreatedAt() time.Time   { return challenge.createdAt }
func (challenge OtpChallenge) ConsumedAt() *time.Time { return challenge.consumedAt }
