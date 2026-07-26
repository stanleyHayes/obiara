package domain

import (
	"errors"
	"net"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidCommand   = errors.New("invalid consent command")
	ErrInvalidIdentity  = errors.New("invalid consent identity")
	ErrInvalidEvidence  = errors.New("invalid consent evidence")
	ErrPurposeMismatch  = errors.New("consent purpose mismatch")
	ErrStaleRevision    = errors.New("stale consent revision")
	ErrAlreadyWithdrawn = errors.New("consent is already withdrawn")
	ErrCommandMismatch  = errors.New("consent command replay does not match original")
)

type ActorKind string

const (
	ActorSubject  ActorKind = "subject"
	ActorGuardian ActorKind = "guardian"
	ActorAdmin    ActorKind = "admin"
	ActorSystem   ActorKind = "system"
)

type Source string

const (
	SourceWeb    Source = "web"
	SourceMobile Source = "mobile"
	SourceAdmin  Source = "admin"
	SourceSystem Source = "system"
)

type Action string

const (
	ActionGranted   Action = "granted"
	ActionWithdrawn Action = "withdrawn"
)

type EvidenceKind string

const (
	EvidenceAcknowledgement EvidenceKind = "acknowledgement"
	EvidenceAgeAffirmation  EvidenceKind = "age_affirmation"
)

var opaqueIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

// Evidence intentionally contains references, never a date of birth, age,
// email address, free-form note, device identifier, or network address.
type Evidence struct {
	kind          EvidenceKind
	policyVersion uint64
	reference     string
}

func NewEvidence(kind EvidenceKind, policyVersion uint64, reference string) (Evidence, error) {
	reference = strings.TrimSpace(reference)
	switch kind {
	case EvidenceAcknowledgement, EvidenceAgeAffirmation:
	default:
		return Evidence{}, ErrInvalidEvidence
	}
	if policyVersion == 0 || !validOpaqueReference(reference) {
		return Evidence{}, ErrInvalidEvidence
	}
	return Evidence{kind: kind, policyVersion: policyVersion, reference: reference}, nil
}

func (evidence Evidence) Kind() EvidenceKind    { return evidence.kind }
func (evidence Evidence) PolicyVersion() uint64 { return evidence.policyVersion }
func (evidence Evidence) Reference() string     { return evidence.reference }

type Event struct {
	revision       uint64
	commandID      string
	action         Action
	purposeVersion uint64
	actorID        string
	actorKind      ActorKind
	source         Source
	evidence       Evidence
	recordedAt     time.Time
}

func (event Event) Revision() uint64       { return event.revision }
func (event Event) CommandID() string      { return event.commandID }
func (event Event) Action() Action         { return event.action }
func (event Event) PurposeVersion() uint64 { return event.purposeVersion }
func (event Event) ActorID() string        { return event.actorID }
func (event Event) ActorKind() ActorKind   { return event.actorKind }
func (event Event) Source() Source         { return event.source }
func (event Event) Evidence() Evidence     { return event.evidence }
func (event Event) RecordedAt() time.Time  { return event.recordedAt }

type EventParams struct {
	Revision       uint64
	CommandID      string
	Action         Action
	PurposeVersion uint64
	ActorID        string
	ActorKind      ActorKind
	Source         Source
	Evidence       Evidence
	RecordedAt     time.Time
}

// NewEvent is the persistence rehydration boundary. It does not append the
// event to a record; Rehydrate additionally validates ordering.
func NewEvent(params EventParams) (Event, error) {
	params.CommandID = strings.TrimSpace(params.CommandID)
	params.ActorID = strings.TrimSpace(params.ActorID)
	if params.Revision == 0 || !validOpaqueID(params.CommandID) ||
		!validOpaqueID(params.ActorID) || params.PurposeVersion == 0 ||
		params.RecordedAt.IsZero() || !validActorAndSource(params.ActorKind, params.Source) ||
		params.Evidence.kind == "" {
		return Event{}, ErrInvalidCommand
	}
	switch params.Action {
	case ActionGranted, ActionWithdrawn:
	default:
		return Event{}, ErrInvalidCommand
	}
	return Event{
		revision:       params.Revision,
		commandID:      params.CommandID,
		action:         params.Action,
		purposeVersion: params.PurposeVersion,
		actorID:        params.ActorID,
		actorKind:      params.ActorKind,
		source:         params.Source,
		evidence:       params.Evidence,
		recordedAt:     params.RecordedAt.UTC(),
	}, nil
}

// Record is an immutable aggregate. Every state transition appends an event
// and returns a new value, preserving the complete withdrawal history.
type Record struct {
	subjectID string
	purposeID string
	revision  uint64
	history   []Event
}

func NewRecord(subjectID, purposeID string) (Record, error) {
	subjectID = strings.TrimSpace(subjectID)
	purposeID = strings.TrimSpace(purposeID)
	if !validOpaqueID(subjectID) {
		return Record{}, ErrInvalidIdentity
	}
	if !slugPattern.MatchString(purposeID) {
		return Record{}, ErrInvalidPurpose
	}
	return Record{subjectID: subjectID, purposeID: purposeID}, nil
}

// Rehydrate validates persisted history before it crosses into the domain.
func Rehydrate(subjectID, purposeID string, events []Event) (Record, error) {
	record, err := NewRecord(subjectID, purposeID)
	if err != nil {
		return Record{}, err
	}
	commands := make(map[string]struct{}, len(events))
	for index, event := range events {
		if event.revision != uint64(index+1) || event.commandID == "" ||
			event.purposeVersion == 0 || event.recordedAt.IsZero() ||
			!validOpaqueID(event.actorID) || !validActorAndSource(event.actorKind, event.source) ||
			event.evidence.policyVersion != event.purposeVersion {
			return Record{}, ErrInvalidCommand
		}
		if _, duplicate := commands[event.commandID]; duplicate {
			return Record{}, ErrInvalidCommand
		}
		if event.actorKind == ActorSubject && event.actorID != record.subjectID {
			return Record{}, ErrInvalidIdentity
		}
		if index == 0 && event.action == ActionWithdrawn {
			return Record{}, ErrInvalidCommand
		}
		if index > 0 {
			previous := events[index-1]
			if event.recordedAt.Before(previous.recordedAt) ||
				(event.action == ActionWithdrawn && previous.action == ActionWithdrawn) {
				return Record{}, ErrInvalidCommand
			}
		}
		commands[event.commandID] = struct{}{}
		record.history = append(record.history, event)
		record.revision = event.revision
	}
	return record, nil
}

func (record Record) SubjectID() string { return record.subjectID }
func (record Record) PurposeID() string { return record.purposeID }
func (record Record) Revision() uint64  { return record.revision }
func (record Record) History() []Event  { return append([]Event(nil), record.history...) }

func (record Record) HasCommand(commandID string) bool {
	for _, event := range record.history {
		if event.commandID == commandID {
			return true
		}
	}
	return false
}

// ReplayMatches prevents a reused idempotency key from changing the meaning of
// an earlier command. RecordedAt is server-owned and intentionally excluded.
func (record Record) ReplayMatches(change Change, action Action) bool {
	for _, event := range record.history {
		if event.commandID != change.CommandID {
			continue
		}
		return event.action == action &&
			event.purposeVersion == change.Purpose.Version() &&
			event.actorID == change.ActorID &&
			event.actorKind == change.ActorKind &&
			event.source == change.Source &&
			event.evidence == change.Evidence
	}
	return false
}

type Change struct {
	CommandID        string
	ExpectedRevision uint64
	Purpose          Purpose
	ActorID          string
	ActorKind        ActorKind
	Source           Source
	Evidence         Evidence
	RecordedAt       time.Time
}

func (record Record) Grant(change Change) (Record, error) {
	return record.apply(change, ActionGranted)
}

func (record Record) Withdraw(change Change) (Record, error) {
	if len(record.history) == 0 || record.history[len(record.history)-1].action == ActionWithdrawn {
		return Record{}, ErrAlreadyWithdrawn
	}
	return record.apply(change, ActionWithdrawn)
}

func (record Record) apply(change Change, action Action) (Record, error) {
	change.CommandID = strings.TrimSpace(change.CommandID)
	change.ActorID = strings.TrimSpace(change.ActorID)
	if change.CommandID == "" || !validOpaqueID(change.CommandID) ||
		!validOpaqueID(change.ActorID) || change.RecordedAt.IsZero() ||
		!validActorAndSource(change.ActorKind, change.Source) ||
		change.Purpose.ID() != record.purposeID || change.Purpose.Version() == 0 {
		return Record{}, ErrInvalidCommand
	}
	if record.HasCommand(change.CommandID) {
		if record.ReplayMatches(change, action) {
			return record, nil
		}
		return Record{}, ErrCommandMismatch
	}
	if change.ExpectedRevision != record.revision {
		return Record{}, ErrStaleRevision
	}
	if change.Purpose.Kind() == PurposeAge && change.Evidence.kind != EvidenceAgeAffirmation {
		return Record{}, ErrInvalidEvidence
	}
	if change.Purpose.Kind() != PurposeAge && change.Evidence.kind != EvidenceAcknowledgement {
		return Record{}, ErrInvalidEvidence
	}
	if change.Evidence.policyVersion != change.Purpose.Version() {
		return Record{}, ErrInvalidEvidence
	}
	if action == ActionGranted && !change.Purpose.IsActive(change.RecordedAt) {
		return Record{}, ErrPurposeInactive
	}
	if change.ActorKind == ActorSubject && change.ActorID != record.subjectID {
		return Record{}, ErrInvalidIdentity
	}
	event := Event{
		revision:       record.revision + 1,
		commandID:      change.CommandID,
		action:         action,
		purposeVersion: change.Purpose.Version(),
		actorID:        change.ActorID,
		actorKind:      change.ActorKind,
		source:         change.Source,
		evidence:       change.Evidence,
		recordedAt:     change.RecordedAt.UTC(),
	}
	next := Record{
		subjectID: record.subjectID,
		purposeID: record.purposeID,
		revision:  event.revision,
		history:   append(record.History(), event),
	}
	return next, nil
}

// Effective denies by default and requires the latest grant to match the
// current active purpose version. Publishing new terms therefore requires a
// fresh explicit grant.
func (record Record) Effective(current Purpose, at time.Time) bool {
	if len(record.history) == 0 || current.ID() != record.purposeID || !current.IsActive(at) {
		return false
	}
	latest := record.history[len(record.history)-1]
	return latest.action == ActionGranted && latest.purposeVersion == current.Version()
}

func validOpaqueID(value string) bool {
	return opaqueIDPattern.MatchString(value) && net.ParseIP(value) == nil
}

func validOpaqueReference(value string) bool {
	return value != "" && len(value) <= 160 && opaqueIDPattern.MatchString(value)
}

func validActorAndSource(actor ActorKind, source Source) bool {
	switch actor {
	case ActorSubject, ActorGuardian:
		return source == SourceWeb || source == SourceMobile
	case ActorAdmin:
		return source == SourceAdmin
	case ActorSystem:
		return source == SourceSystem
	}
	return false
}
