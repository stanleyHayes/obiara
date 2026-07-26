// Package domain defines the deny-by-default E11-S07 counsel egress policy.
// Counsel content, topics, attendance, and outcomes have no representation in
// the only permitted outbound event.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

type Destination string

const (
	DestinationSafetyEscalation Destination = "safety_escalation"
	DestinationMatchingFeature  Destination = "matching_feature"
	DestinationExplanation      Destination = "matching_explanation"
	DestinationRanking          Destination = "ranking"
	DestinationTrust            Destination = "trust"
	DestinationAIPrompt         Destination = "ai_prompt"
)

type Field string

const (
	FieldEventID        Field = "event_id"
	FieldSubjectKey     Field = "subject_key"
	FieldReasonCode     Field = "reason_code"
	FieldOccurredAt     Field = "occurred_at"
	FieldConsentVersion Field = "consent_version"
)

type ReasonCode string

const ReasonExplicitSafetySupport ReasonCode = "explicit_safety_support"

var (
	ErrDenied       = errors.New("counsel egress denied")
	ErrInvalidEvent = errors.New("invalid counsel safety event")
)

var opaqueKeyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

var minimalSafetyFields = []Field{
	FieldEventID,
	FieldSubjectKey,
	FieldReasonCode,
	FieldOccurredAt,
	FieldConsentVersion,
}

// Permit allows only the exact minimal safety shape. Every matching, ranking,
// explanation, trust, and AI destination is denied regardless of fields.
func Permit(destination Destination, fields []Field) error {
	if destination != DestinationSafetyEscalation || len(fields) != len(minimalSafetyFields) {
		return ErrDenied
	}
	provided := slices.Clone(fields)
	slices.Sort(provided)
	required := slices.Clone(minimalSafetyFields)
	slices.Sort(required)
	if !slices.Equal(provided, required) {
		return ErrDenied
	}
	return nil
}

// SafetyEvent is intentionally unable to carry counsel content, topic,
// attendance, session identity, outcome, actor identity, or free text.
type SafetyEvent struct {
	ID             string
	SubjectKey     string
	Reason         ReasonCode
	OccurredAt     time.Time
	ConsentVersion uint64
}

func NewSafetyEvent(id, subjectKey string, reason ReasonCode, occurredAt time.Time, consentVersion uint64) (SafetyEvent, error) {
	event := SafetyEvent{
		ID:             id,
		SubjectKey:     subjectKey,
		Reason:         reason,
		OccurredAt:     occurredAt.UTC(),
		ConsentVersion: consentVersion,
	}
	if !opaqueKeyPattern.MatchString(event.ID) ||
		!opaqueKeyPattern.MatchString(event.SubjectKey) ||
		event.Reason != ReasonExplicitSafetySupport ||
		event.OccurredAt.IsZero() ||
		event.ConsentVersion == 0 {
		return SafetyEvent{}, ErrInvalidEvent
	}
	if err := Permit(DestinationSafetyEscalation, minimalSafetyFields); err != nil {
		return SafetyEvent{}, err
	}
	return event, nil
}
