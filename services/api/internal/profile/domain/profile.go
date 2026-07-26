package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalidProfile  = errors.New("invalid profile")
	ErrUnsafeProfile   = errors.New("profile contains disallowed personal data")
	ErrStaleRevision   = errors.New("stale profile revision")
	ErrCommandMismatch = errors.New("profile command replay does not match original")
	ErrInvalidAudience = errors.New("invalid profile audience")
	ErrConsentRequired = errors.New("community visibility requires consent")
)

type Visibility string

const (
	VisibilityPrivate   Visibility = "private"
	VisibilityCircles   Visibility = "circles"
	VisibilityCommunity Visibility = "community"
)

type Audience string

const (
	AudienceSelf      Audience = "self"
	AudienceCircle    Audience = "circle"
	AudienceCommunity Audience = "community"
)

var (
	opaqueIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	emailPattern    = regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
	urlPattern      = regexp.MustCompile(`(?i)\b(?:https?://|www\.)\S+`)
	phonePattern    = regexp.MustCompile(`(?:\+?\d[\d ()-]{7,}\d)`)
)

// Field is a privacy-scoped value. ConsentRef is an opaque consent record
// reference and never contains evidence, policy copy, or subject identifiers.
type Field struct {
	value      string
	visibility Visibility
	consentRef string
}

func NewField(value string, visibility Visibility, consentRef string, limit int, required bool) (Field, error) {
	value = strings.TrimSpace(value)
	consentRef = strings.TrimSpace(consentRef)
	if (required && value == "") || !utf8.ValidString(value) || utf8.RuneCountInString(value) > limit {
		return Field{}, ErrInvalidProfile
	}
	if value != "" && containsDisallowedPersonalData(value) {
		return Field{}, ErrUnsafeProfile
	}
	switch visibility {
	case VisibilityPrivate, VisibilityCircles:
		if consentRef != "" {
			return Field{}, ErrInvalidProfile
		}
	case VisibilityCommunity:
		if !opaqueIDPattern.MatchString(consentRef) {
			return Field{}, ErrConsentRequired
		}
	default:
		return Field{}, ErrInvalidProfile
	}
	return Field{value: value, visibility: visibility, consentRef: consentRef}, nil
}

func (field Field) Value() string          { return field.value }
func (field Field) Visibility() Visibility { return field.visibility }
func (field Field) ConsentRef() string     { return field.consentRef }

func (field Field) VisibleTo(audience Audience) (bool, error) {
	switch audience {
	case AudienceSelf:
		return true, nil
	case AudienceCircle:
		return field.visibility == VisibilityCircles || field.visibility == VisibilityCommunity, nil
	case AudienceCommunity:
		return field.visibility == VisibilityCommunity, nil
	default:
		return false, ErrInvalidAudience
	}
}

type AppliedCommand struct {
	id          string
	fingerprint string
	revision    uint64
}

func NewAppliedCommand(id, fingerprint string, revision uint64) (AppliedCommand, error) {
	id = strings.TrimSpace(id)
	fingerprint = strings.TrimSpace(fingerprint)
	if !opaqueIDPattern.MatchString(id) || len(fingerprint) != sha256.Size*2 || revision == 0 {
		return AppliedCommand{}, ErrInvalidProfile
	}
	if _, err := hex.DecodeString(fingerprint); err != nil {
		return AppliedCommand{}, ErrInvalidProfile
	}
	return AppliedCommand{id: id, fingerprint: fingerprint, revision: revision}, nil
}

func (command AppliedCommand) ID() string          { return command.id }
func (command AppliedCommand) Fingerprint() string { return command.fingerprint }
func (command AppliedCommand) Revision() uint64    { return command.revision }

type Profile struct {
	memberID     string
	displayName  Field
	introduction Field
	revision     uint64
	updatedAt    time.Time
	commands     []AppliedCommand
}

type State struct {
	MemberID     string
	DisplayName  Field
	Introduction Field
	Revision     uint64
	UpdatedAt    time.Time
	Commands     []AppliedCommand
}

func Rehydrate(state State) (Profile, error) {
	state.MemberID = strings.TrimSpace(state.MemberID)
	if !opaqueIDPattern.MatchString(state.MemberID) || state.Revision == 0 || state.UpdatedAt.IsZero() ||
		state.DisplayName.value == "" {
		return Profile{}, ErrInvalidProfile
	}
	if _, err := NewField(state.DisplayName.value, state.DisplayName.visibility, state.DisplayName.consentRef, 80, true); err != nil {
		return Profile{}, err
	}
	if _, err := NewField(state.Introduction.value, state.Introduction.visibility, state.Introduction.consentRef, 280, false); err != nil {
		return Profile{}, err
	}
	seen := make(map[string]struct{}, len(state.Commands))
	if uint64(len(state.Commands)) != state.Revision {
		return Profile{}, ErrInvalidProfile
	}
	for index, command := range state.Commands {
		if command.revision != uint64(index+1) || command.id == "" || command.fingerprint == "" {
			return Profile{}, ErrInvalidProfile
		}
		if _, duplicate := seen[command.id]; duplicate {
			return Profile{}, ErrInvalidProfile
		}
		seen[command.id] = struct{}{}
	}
	return Profile{
		memberID: state.MemberID, displayName: state.DisplayName, introduction: state.Introduction,
		revision: state.Revision, updatedAt: state.UpdatedAt.UTC(),
		commands: append([]AppliedCommand(nil), state.Commands...),
	}, nil
}

type Change struct {
	CommandID        string
	ExpectedRevision uint64
	DisplayName      Field
	Introduction     Field
	RecordedAt       time.Time
}

func Create(memberID string, change Change) (Profile, error) {
	memberID = strings.TrimSpace(memberID)
	if !opaqueIDPattern.MatchString(memberID) || change.ExpectedRevision != 0 {
		return Profile{}, ErrInvalidProfile
	}
	return apply(Profile{memberID: memberID}, change)
}

func (profile Profile) Update(change Change) (Profile, error) {
	return apply(profile, change)
}

func apply(profile Profile, change Change) (Profile, error) {
	change.CommandID = strings.TrimSpace(change.CommandID)
	if !opaqueIDPattern.MatchString(change.CommandID) || change.RecordedAt.IsZero() ||
		change.DisplayName.value == "" {
		return Profile{}, ErrInvalidProfile
	}
	fingerprint := fingerprintFor(profile.memberID, change)
	for _, command := range profile.commands {
		if command.id == change.CommandID {
			if command.fingerprint != fingerprint {
				return Profile{}, ErrCommandMismatch
			}
			return profile, nil
		}
	}
	if change.ExpectedRevision != profile.revision {
		return Profile{}, ErrStaleRevision
	}
	revision := profile.revision + 1
	applied, _ := NewAppliedCommand(change.CommandID, fingerprint, revision)
	return Profile{
		memberID: profile.memberID, displayName: change.DisplayName, introduction: change.Introduction,
		revision: revision, updatedAt: change.RecordedAt.UTC(),
		commands: append(append([]AppliedCommand(nil), profile.commands...), applied),
	}, nil
}

func (profile Profile) MemberID() string     { return profile.memberID }
func (profile Profile) DisplayName() Field   { return profile.displayName }
func (profile Profile) Introduction() Field  { return profile.introduction }
func (profile Profile) Revision() uint64     { return profile.revision }
func (profile Profile) UpdatedAt() time.Time { return profile.updatedAt }
func (profile Profile) Commands() []AppliedCommand {
	return append([]AppliedCommand(nil), profile.commands...)
}

func (profile Profile) HasCommand(commandID string) bool {
	for _, command := range profile.commands {
		if command.id == commandID {
			return true
		}
	}
	return false
}

func containsDisallowedPersonalData(value string) bool {
	return emailPattern.MatchString(value) || urlPattern.MatchString(value) || phonePattern.MatchString(value)
}

func fingerprintFor(memberID string, change Change) string {
	payload := strings.Join([]string{
		memberID,
		change.DisplayName.value, string(change.DisplayName.visibility), change.DisplayName.consentRef,
		change.Introduction.value, string(change.Introduction.visibility), change.Introduction.consentRef,
	}, "\x00")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
