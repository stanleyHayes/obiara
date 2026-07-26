package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	ExclusionPeriod  = 90 * 24 * time.Hour
	NotificationKind = "seed_not_available"
)

var (
	ErrInvalid         = errors.New("invalid seed decline")
	ErrCommandMismatch = errors.New("decline command replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	opaquePattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
)

type Decline struct {
	id           string
	declinerKey  string
	seedKey      string
	recipientKey string
	commandID    string
	fingerprint  string
	declinedAt   time.Time
	excludedTill time.Time
}

func New(id, declinerKey, seedKey, recipientKey, commandID string, declinedAt time.Time) (Decline, error) {
	id, commandID = strings.TrimSpace(id), strings.TrimSpace(commandID)
	if !opaquePattern.MatchString(id) || !keyPattern.MatchString(declinerKey) ||
		!keyPattern.MatchString(seedKey) || !keyPattern.MatchString(recipientKey) ||
		!opaquePattern.MatchString(commandID) || declinedAt.IsZero() {
		return Decline{}, ErrInvalid
	}
	declinedAt = declinedAt.UTC()
	return Decline{
		id: id, declinerKey: declinerKey, seedKey: seedKey, recipientKey: recipientKey,
		commandID: commandID, fingerprint: Fingerprint(declinerKey, seedKey, recipientKey),
		declinedAt: declinedAt, excludedTill: declinedAt.Add(ExclusionPeriod),
	}, nil
}

func Rehydrate(id, declinerKey, seedKey, recipientKey, commandID, fingerprint string, declinedAt, excludedTill time.Time) (Decline, error) {
	decline, err := New(id, declinerKey, seedKey, recipientKey, commandID, declinedAt)
	if err != nil || fingerprint != decline.fingerprint || !excludedTill.Equal(decline.excludedTill) {
		return Decline{}, ErrInvalid
	}
	return decline, nil
}

func Fingerprint(declinerKey, seedKey, recipientKey string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{declinerKey, seedKey, recipientKey}, "\x00")))
	return hex.EncodeToString(sum[:])
}

// Excludes is deliberately half-open: the pair becomes eligible at the exact
// instant the 90-day window ends.
func (decline Decline) Excludes(at time.Time) bool {
	at = at.UTC()
	return !at.Before(decline.declinedAt) && at.Before(decline.excludedTill)
}

func (decline Decline) ID() string            { return decline.id }
func (decline Decline) DeclinerKey() string   { return decline.declinerKey }
func (decline Decline) SeedKey() string       { return decline.seedKey }
func (decline Decline) RecipientKey() string  { return decline.recipientKey }
func (decline Decline) CommandID() string     { return decline.commandID }
func (decline Decline) Fingerprint() string   { return decline.fingerprint }
func (decline Decline) DeclinedAt() time.Time { return decline.declinedAt }
func (decline Decline) ExcludedUntil() time.Time {
	return decline.excludedTill
}

type Notification struct {
	eventKey     string
	recipientKey string
	kind         string
	occurredAt   time.Time
}

func NewNotification(eventKey, recipientKey string, occurredAt time.Time) (Notification, error) {
	if !keyPattern.MatchString(eventKey) || !keyPattern.MatchString(recipientKey) || occurredAt.IsZero() {
		return Notification{}, ErrInvalid
	}
	return Notification{eventKey: eventKey, recipientKey: recipientKey, kind: NotificationKind, occurredAt: occurredAt.UTC()}, nil
}

func (notification Notification) EventKey() string      { return notification.eventKey }
func (notification Notification) RecipientKey() string  { return notification.recipientKey }
func (notification Notification) Kind() string          { return notification.kind }
func (notification Notification) OccurredAt() time.Time { return notification.occurredAt }
