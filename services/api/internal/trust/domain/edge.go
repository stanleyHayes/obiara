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
	MaxProjectionDepth = 4
	MaxProjectionNodes = 100
)

var (
	ErrInvalidEdge      = errors.New("invalid trust edge")
	ErrStaleRevision    = errors.New("stale trust edge revision")
	ErrCommandMismatch  = errors.New("trust edge command replay mismatch")
	ErrAlreadyRevoked   = errors.New("trust edge already revoked")
	ErrProjectionBounds = errors.New("trust projection exceeds bounds")
	ErrAccessDenied     = errors.New("trust projection access denied")
)

var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type EdgeType string

const (
	EdgeCircleMember EdgeType = "circle_member"
	EdgeVouch        EdgeType = "vouch"
	EdgeKnown        EdgeType = "known"
	EdgeHost         EdgeType = "host"
)

type Visibility string

const (
	VisibilityOwnerOnly     Visibility = "owner_only"
	VisibilityParticipants  Visibility = "participants"
	VisibilityConsentedPath Visibility = "consented_path"
)

type Action string

const (
	ActionCreated Action = "created"
	ActionRevoked Action = "revoked"
)

// Event is a privacy-minimal audit record. It identifies the edge and actor by
// opaque references but never copies endpoints, consent evidence or provenance.
type Event struct {
	revision  uint64
	commandID string
	action    Action
	actorRef  string
	at        time.Time
}

func NewEvent(revision uint64, commandID string, action Action, actorRef string, at time.Time) (Event, error) {
	commandID, actorRef = strings.TrimSpace(commandID), strings.TrimSpace(actorRef)
	if revision == 0 || !opaquePattern.MatchString(commandID) || !opaquePattern.MatchString(actorRef) ||
		(action != ActionCreated && action != ActionRevoked) || at.IsZero() {
		return Event{}, ErrInvalidEdge
	}
	return Event{revision: revision, commandID: commandID, action: action, actorRef: actorRef, at: at.UTC()}, nil
}

func (event Event) Revision() uint64  { return event.revision }
func (event Event) CommandID() string { return event.commandID }
func (event Event) Action() Action    { return event.action }
func (event Event) ActorRef() string  { return event.actorRef }
func (event Event) At() time.Time     { return event.at }

type AppliedCommand struct {
	id          string
	fingerprint string
	revision    uint64
}

func NewAppliedCommand(id, fingerprint string, revision uint64) (AppliedCommand, error) {
	id, fingerprint = strings.TrimSpace(id), strings.TrimSpace(fingerprint)
	if !opaquePattern.MatchString(id) || len(fingerprint) != sha256.Size*2 || revision == 0 {
		return AppliedCommand{}, ErrInvalidEdge
	}
	if _, err := hex.DecodeString(fingerprint); err != nil {
		return AppliedCommand{}, ErrInvalidEdge
	}
	return AppliedCommand{id: id, fingerprint: fingerprint, revision: revision}, nil
}

func (command AppliedCommand) ID() string          { return command.id }
func (command AppliedCommand) Fingerprint() string { return command.fingerprint }
func (command AppliedCommand) Revision() uint64    { return command.revision }

type Edge struct {
	id            string
	sourceID      string
	targetID      string
	edgeType      EdgeType
	provenanceRef string
	consentRef    string
	visibility    Visibility
	createdAt     time.Time
	expiresAt     *time.Time
	revokedAt     *time.Time
	revision      uint64
	history       []Event
	commands      []AppliedCommand
}

type Params struct {
	ID            string
	SourceID      string
	TargetID      string
	Type          EdgeType
	ProvenanceRef string
	ConsentRef    string
	Visibility    Visibility
	CreatedAt     time.Time
	ExpiresAt     *time.Time
}

type Command struct {
	ID               string
	ExpectedRevision uint64
	ActorRef         string
	Kind             string
	Payload          string
	RecordedAt       time.Time
}

func Create(params Params, command Command) (Edge, error) {
	params.ID, params.SourceID, params.TargetID = strings.TrimSpace(params.ID), strings.TrimSpace(params.SourceID), strings.TrimSpace(params.TargetID)
	params.ProvenanceRef, params.ConsentRef = strings.TrimSpace(params.ProvenanceRef), strings.TrimSpace(params.ConsentRef)
	if !opaquePattern.MatchString(params.ID) || !opaquePattern.MatchString(params.SourceID) ||
		!opaquePattern.MatchString(params.TargetID) || params.SourceID == params.TargetID ||
		!opaquePattern.MatchString(params.ProvenanceRef) || !opaquePattern.MatchString(params.ConsentRef) ||
		!validType(params.Type) || !validVisibility(params.Visibility) || params.CreatedAt.IsZero() ||
		command.ExpectedRevision != 0 || !command.RecordedAt.UTC().Equal(params.CreatedAt.UTC()) {
		return Edge{}, ErrInvalidEdge
	}
	expiresAt := utcPointer(params.ExpiresAt)
	if expiresAt != nil && !expiresAt.After(params.CreatedAt.UTC()) {
		return Edge{}, ErrInvalidEdge
	}
	edge := Edge{
		id: params.ID, sourceID: params.SourceID, targetID: params.TargetID,
		edgeType: params.Type, provenanceRef: params.ProvenanceRef, consentRef: params.ConsentRef,
		visibility: params.Visibility, createdAt: params.CreatedAt.UTC(), expiresAt: expiresAt,
	}
	return edge.apply(command, ActionCreated)
}

type State struct {
	Params    Params
	RevokedAt *time.Time
	Revision  uint64
	History   []Event
	Commands  []AppliedCommand
}

func Rehydrate(state State) (Edge, error) {
	if state.Revision == 0 || state.Revision > 2 || uint64(len(state.History)) != state.Revision ||
		uint64(len(state.Commands)) != state.Revision {
		return Edge{}, ErrInvalidEdge
	}
	create := Command{
		ID: state.History[0].commandID, ActorRef: state.History[0].actorRef,
		Kind: "edge.create", Payload: state.Commands[0].fingerprint, RecordedAt: state.Params.CreatedAt,
	}
	// Validate immutable fields independently; persisted command fingerprints
	// are validated below and are not recomputed from unavailable command input.
	state.Params.ID, state.Params.SourceID, state.Params.TargetID = strings.TrimSpace(state.Params.ID), strings.TrimSpace(state.Params.SourceID), strings.TrimSpace(state.Params.TargetID)
	state.Params.ProvenanceRef, state.Params.ConsentRef = strings.TrimSpace(state.Params.ProvenanceRef), strings.TrimSpace(state.Params.ConsentRef)
	if !opaquePattern.MatchString(state.Params.ID) || !opaquePattern.MatchString(state.Params.SourceID) ||
		!opaquePattern.MatchString(state.Params.TargetID) || state.Params.SourceID == state.Params.TargetID ||
		!opaquePattern.MatchString(state.Params.ProvenanceRef) || !opaquePattern.MatchString(state.Params.ConsentRef) ||
		!validType(state.Params.Type) || !validVisibility(state.Params.Visibility) ||
		state.Params.CreatedAt.IsZero() || create.ID == "" {
		return Edge{}, ErrInvalidEdge
	}
	edge := Edge{
		id: state.Params.ID, sourceID: state.Params.SourceID, targetID: state.Params.TargetID,
		edgeType: state.Params.Type, provenanceRef: state.Params.ProvenanceRef,
		consentRef: state.Params.ConsentRef, visibility: state.Params.Visibility,
		createdAt: state.Params.CreatedAt.UTC(), expiresAt: utcPointer(state.Params.ExpiresAt),
		revokedAt: utcPointer(state.RevokedAt), revision: state.Revision,
		history: append([]Event(nil), state.History...), commands: append([]AppliedCommand(nil), state.Commands...),
	}
	if edge.expiresAt != nil && !edge.expiresAt.After(edge.createdAt) {
		return Edge{}, ErrInvalidEdge
	}
	seen := map[string]struct{}{}
	for index, event := range edge.history {
		command := edge.commands[index]
		if event.revision != uint64(index+1) || command.revision != uint64(index+1) ||
			event.commandID != command.id || event.at.Before(edge.createdAt) {
			return Edge{}, ErrInvalidEdge
		}
		expectedAction := ActionCreated
		if index > 0 {
			expectedAction = ActionRevoked
		}
		if event.action != expectedAction {
			return Edge{}, ErrInvalidEdge
		}
		if _, duplicate := seen[command.id]; duplicate {
			return Edge{}, ErrInvalidEdge
		}
		seen[command.id] = struct{}{}
	}
	if (edge.revokedAt == nil) != (edge.revision == 1) ||
		(edge.revokedAt != nil && !edge.revokedAt.Equal(edge.history[len(edge.history)-1].at)) {
		return Edge{}, ErrInvalidEdge
	}
	return edge, nil
}

func (edge Edge) Revoke(command Command) (Edge, error) {
	for _, applied := range edge.commands {
		if applied.id != strings.TrimSpace(command.ID) {
			continue
		}
		if applied.fingerprint != commandFingerprint(edge.id, command) {
			return Edge{}, ErrCommandMismatch
		}
		return edge, nil
	}
	if edge.revokedAt != nil {
		return Edge{}, ErrAlreadyRevoked
	}
	if command.ExpectedRevision != edge.revision {
		return Edge{}, ErrStaleRevision
	}
	at := command.RecordedAt.UTC()
	next := edge
	next.revokedAt = &at
	return next.apply(command, ActionRevoked)
}

func (edge Edge) apply(command Command, action Action) (Edge, error) {
	command.ID, command.ActorRef, command.Kind, command.Payload = strings.TrimSpace(command.ID), strings.TrimSpace(command.ActorRef), strings.TrimSpace(command.Kind), strings.TrimSpace(command.Payload)
	if !opaquePattern.MatchString(command.ID) || !opaquePattern.MatchString(command.ActorRef) ||
		!opaquePattern.MatchString(command.Kind) || command.Payload == "" || command.RecordedAt.IsZero() ||
		command.RecordedAt.UTC().Before(edge.createdAt) {
		return Edge{}, ErrInvalidEdge
	}
	revision := edge.revision + 1
	event, _ := NewEvent(revision, command.ID, action, command.ActorRef, command.RecordedAt)
	applied, _ := NewAppliedCommand(command.ID, commandFingerprint(edge.id, command), revision)
	edge.revision = revision
	edge.history = append(append([]Event(nil), edge.history...), event)
	edge.commands = append(append([]AppliedCommand(nil), edge.commands...), applied)
	return edge, nil
}

func (edge Edge) Active(at time.Time) bool {
	at = at.UTC()
	return edge.revokedAt == nil && (edge.expiresAt == nil || at.Before(*edge.expiresAt))
}

func (edge Edge) VisibleTo(ownerID string, consented bool, at time.Time) bool {
	if !edge.Active(at) {
		return false
	}
	switch edge.visibility {
	case VisibilityOwnerOnly:
		return ownerID == edge.sourceID
	case VisibilityParticipants:
		return ownerID == edge.sourceID || ownerID == edge.targetID
	case VisibilityConsentedPath:
		return consented
	default:
		return false
	}
}

func (edge Edge) ID() string                 { return edge.id }
func (edge Edge) SourceID() string           { return edge.sourceID }
func (edge Edge) TargetID() string           { return edge.targetID }
func (edge Edge) Type() EdgeType             { return edge.edgeType }
func (edge Edge) ProvenanceRef() string      { return edge.provenanceRef }
func (edge Edge) ConsentRef() string         { return edge.consentRef }
func (edge Edge) Visibility() Visibility     { return edge.visibility }
func (edge Edge) CreatedAt() time.Time       { return edge.createdAt }
func (edge Edge) ExpiresAt() *time.Time      { return utcPointer(edge.expiresAt) }
func (edge Edge) RevokedAt() *time.Time      { return utcPointer(edge.revokedAt) }
func (edge Edge) Revision() uint64           { return edge.revision }
func (edge Edge) History() []Event           { return append([]Event(nil), edge.history...) }
func (edge Edge) Commands() []AppliedCommand { return append([]AppliedCommand(nil), edge.commands...) }
func (edge Edge) HasCommand(id string) bool {
	for _, command := range edge.commands {
		if command.id == strings.TrimSpace(id) {
			return true
		}
	}
	return false
}
func (edge Edge) VerifyReplay(command Command) error {
	for _, applied := range edge.commands {
		if applied.id == strings.TrimSpace(command.ID) {
			if applied.fingerprint == commandFingerprint(edge.id, command) {
				return nil
			}
			return ErrCommandMismatch
		}
	}
	return ErrCommandMismatch
}

func validType(edgeType EdgeType) bool {
	switch edgeType {
	case EdgeCircleMember, EdgeVouch, EdgeKnown, EdgeHost:
		return true
	default:
		return false
	}
}

func validVisibility(visibility Visibility) bool {
	switch visibility {
	case VisibilityOwnerOnly, VisibilityParticipants, VisibilityConsentedPath:
		return true
	default:
		return false
	}
}

func commandFingerprint(edgeID string, command Command) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		edgeID, command.ActorRef, command.Kind, command.Payload,
	}, "\x00")))
	return hex.EncodeToString(sum[:])
}

func utcPointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
