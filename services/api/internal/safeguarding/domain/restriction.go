// Package domain contains the fail-closed age gate and privacy-minimal purge
// proof model. Dates of birth and raw subject identifiers are never retained
// in a Restriction.
package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	MinimumAge = 18
	PurgeSLA   = 24 * time.Hour
)

var (
	ErrInvalidAssessment  = errors.New("invalid age assessment")
	ErrInvalidRestriction = errors.New("invalid safeguarding restriction")
	ErrStaleVersion       = errors.New("stale safeguarding restriction")
)

var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

// AgeAt calculates completed years using server time. It intentionally does
// not approximate years with a duration, which is wrong around birthdays.
func AgeAt(dateOfBirth, serverTime time.Time) (int, error) {
	if dateOfBirth.IsZero() || serverTime.IsZero() {
		return 0, ErrInvalidAssessment
	}
	birth := dateOfBirth.UTC()
	now := serverTime.UTC()
	if birth.After(now) {
		return 0, ErrInvalidAssessment
	}
	age := now.Year() - birth.Year()
	birthday := time.Date(now.Year(), birth.Month(), birth.Day(), 0, 0, 0, 0, time.UTC)
	if now.Before(birthday) {
		age--
	}
	return age, nil
}

func Eligible(dateOfBirth, serverTime time.Time) (bool, error) {
	age, err := AgeAt(dateOfBirth, serverTime)
	if err != nil {
		return false, err
	}
	return age >= MinimumAge, nil
}

type PurgeStatus string

const (
	PurgePending   PurgeStatus = "pending"
	PurgeCompleted PurgeStatus = "completed"
)

// Restriction is retained compliance proof. SubjectKey and SourceKey must be
// one-way keyed digests supplied by the application boundary.
type Restriction struct {
	id          string
	commandID   string
	subjectKey  string
	sourceKey   string
	blockedAt   time.Time
	purgeDueAt  time.Time
	purgeStatus PurgeStatus
	purgedAt    time.Time
	version     uint64
}

func NewRestriction(id, commandID, subjectKey, sourceKey string, serverTime time.Time) (Restriction, error) {
	id = strings.TrimSpace(id)
	commandID = strings.TrimSpace(commandID)
	subjectKey = strings.TrimSpace(subjectKey)
	sourceKey = strings.TrimSpace(sourceKey)
	if !validOpaque(id) || !validOpaque(commandID) || !validDigest(subjectKey) ||
		!validDigest(sourceKey) || serverTime.IsZero() {
		return Restriction{}, ErrInvalidRestriction
	}
	blockedAt := serverTime.UTC()
	return Restriction{
		id: id, commandID: commandID, subjectKey: subjectKey, sourceKey: sourceKey,
		blockedAt: blockedAt, purgeDueAt: blockedAt.Add(PurgeSLA),
		purgeStatus: PurgePending, version: 1,
	}, nil
}

func Rehydrate(id, commandID, subjectKey, sourceKey string, blockedAt, purgeDueAt time.Time, status PurgeStatus, purgedAt time.Time, version uint64) (Restriction, error) {
	restriction := Restriction{
		id: strings.TrimSpace(id), commandID: strings.TrimSpace(commandID),
		subjectKey: strings.TrimSpace(subjectKey), sourceKey: strings.TrimSpace(sourceKey),
		blockedAt: blockedAt.UTC(), purgeDueAt: purgeDueAt.UTC(),
		purgeStatus: status, purgedAt: purgedAt.UTC(), version: version,
	}
	if !validOpaque(restriction.id) || !validOpaque(restriction.commandID) ||
		!validDigest(restriction.subjectKey) || !validDigest(restriction.sourceKey) ||
		restriction.blockedAt.IsZero() || !restriction.purgeDueAt.Equal(restriction.blockedAt.Add(PurgeSLA)) ||
		version == 0 {
		return Restriction{}, ErrInvalidRestriction
	}
	switch status {
	case PurgePending:
		if !restriction.purgedAt.IsZero() {
			return Restriction{}, ErrInvalidRestriction
		}
	case PurgeCompleted:
		if restriction.purgedAt.IsZero() || restriction.purgedAt.Before(restriction.blockedAt) {
			return Restriction{}, ErrInvalidRestriction
		}
	default:
		return Restriction{}, ErrInvalidRestriction
	}
	return restriction, nil
}

func (restriction Restriction) MarkPurged(serverTime time.Time, expectedVersion uint64) (Restriction, error) {
	if serverTime.IsZero() || expectedVersion != restriction.version {
		return Restriction{}, ErrStaleVersion
	}
	if restriction.purgeStatus == PurgeCompleted {
		return restriction, nil
	}
	restriction.purgeStatus = PurgeCompleted
	restriction.purgedAt = serverTime.UTC()
	restriction.version++
	return restriction, nil
}

func (restriction Restriction) ID() string               { return restriction.id }
func (restriction Restriction) CommandID() string        { return restriction.commandID }
func (restriction Restriction) SubjectKey() string       { return restriction.subjectKey }
func (restriction Restriction) SourceKey() string        { return restriction.sourceKey }
func (restriction Restriction) BlockedAt() time.Time     { return restriction.blockedAt }
func (restriction Restriction) PurgeDueAt() time.Time    { return restriction.purgeDueAt }
func (restriction Restriction) PurgeStatus() PurgeStatus { return restriction.purgeStatus }
func (restriction Restriction) PurgedAt() time.Time      { return restriction.purgedAt }
func (restriction Restriction) Version() uint64          { return restriction.version }
func (restriction Restriction) Blocked() bool            { return true }
func (restriction Restriction) PurgedWithinSLA() bool {
	return restriction.purgeStatus == PurgeCompleted &&
		!restriction.purgedAt.After(restriction.purgeDueAt)
}

func validOpaque(value string) bool {
	return opaquePattern.MatchString(value)
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return false
		}
	}
	return true
}
