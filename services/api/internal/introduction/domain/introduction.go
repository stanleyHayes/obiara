// Package domain models the consent-bound Voice of Introduction lifecycle.
// It retains media/transcript references and metadata, never audio bytes or
// transcript text.
package domain

import (
	"errors"
	"mime"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidIntroduction = errors.New("invalid voice introduction")
	ErrInvalidTransition   = errors.New("invalid voice introduction transition")
	ErrStaleVersion        = errors.New("stale voice introduction version")
	ErrCommandConflict     = errors.New("voice introduction command conflicts with prior use")
	ErrRetentionActive     = errors.New("voice introduction retention period is active")
	ErrLegalHold           = errors.New("voice introduction is under legal hold")
)

type Status string

const (
	StatusDraft                  Status = "draft"
	StatusUploadAuthorized       Status = "upload_authorized"
	StatusUploaded               Status = "uploaded"
	StatusTranscribing           Status = "transcribing"
	StatusReady                  Status = "ready"
	StatusTranscriptionUncertain Status = "transcription_uncertain"
	StatusTranscriptionFailed    Status = "transcription_failed"
	StatusCancelled              Status = "cancelled"
	StatusRevoked                Status = "revoked"
)

type DataStatus string

const (
	DataRetained     DataStatus = "retained"
	DataPurgePending DataStatus = "purge_pending"
	DataPurged       DataStatus = "purged"
)

type Action string

const (
	ActionCreated            Action = "created"
	ActionUploadAuthorized   Action = "upload_authorized"
	ActionUploadConfirmed    Action = "upload_confirmed"
	ActionTranscriptionStart Action = "transcription_started"
	ActionTranscriptionReady Action = "transcription_ready"
	ActionTranscriptionQueue Action = "transcription_uncertain"
	ActionTranscriptionFail  Action = "transcription_failed"
	ActionCancelled          Action = "cancelled"
	ActionRevoked            Action = "revoked"
	ActionPurged             Action = "purged"
)

var opaquePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)

type ConsentSnapshot struct {
	purposeID   string
	version     uint64
	evaluatedAt time.Time
}

func NewConsentSnapshot(purposeID string, version uint64, evaluatedAt time.Time) (ConsentSnapshot, error) {
	purposeID = strings.TrimSpace(purposeID)
	if !opaquePattern.MatchString(purposeID) || version == 0 || evaluatedAt.IsZero() {
		return ConsentSnapshot{}, ErrInvalidIntroduction
	}
	return ConsentSnapshot{purposeID: purposeID, version: version, evaluatedAt: evaluatedAt.UTC()}, nil
}

func (snapshot ConsentSnapshot) PurposeID() string      { return snapshot.purposeID }
func (snapshot ConsentSnapshot) Version() uint64        { return snapshot.version }
func (snapshot ConsentSnapshot) EvaluatedAt() time.Time { return snapshot.evaluatedAt }

type MediaRef struct {
	assetID     string
	contentType string
	size        int64
	duration    time.Duration
	checksum    string
}

func NewMediaRef(assetID, contentType string, size int64, duration time.Duration, checksum string) (MediaRef, error) {
	assetID = strings.TrimSpace(assetID)
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	checksum = strings.ToLower(strings.TrimSpace(checksum))
	parsed, _, err := mime.ParseMediaType(contentType)
	if !opaquePattern.MatchString(assetID) || err != nil || !strings.HasPrefix(parsed, "audio/") ||
		size < 0 || duration < 0 || (checksum != "" && !validDigest(checksum)) {
		return MediaRef{}, ErrInvalidIntroduction
	}
	return MediaRef{
		assetID: assetID, contentType: contentType, size: size,
		duration: duration, checksum: checksum,
	}, nil
}

func (ref MediaRef) AssetID() string         { return ref.assetID }
func (ref MediaRef) ContentType() string     { return ref.contentType }
func (ref MediaRef) Size() int64             { return ref.size }
func (ref MediaRef) Duration() time.Duration { return ref.duration }
func (ref MediaRef) Checksum() string        { return ref.checksum }
func (ref MediaRef) Complete() bool {
	return ref.size > 0 && ref.duration > 0 && ref.checksum != ""
}

type TranscriptRef struct {
	id         string
	language   string
	confidence uint8
}

func NewTranscriptRef(id, language string, confidence uint8) (TranscriptRef, error) {
	id = strings.TrimSpace(id)
	language = strings.ToLower(strings.TrimSpace(language))
	if !opaquePattern.MatchString(id) || !opaquePattern.MatchString(language) || confidence > 100 {
		return TranscriptRef{}, ErrInvalidIntroduction
	}
	return TranscriptRef{id: id, language: language, confidence: confidence}, nil
}

func (ref TranscriptRef) ID() string        { return ref.id }
func (ref TranscriptRef) Language() string  { return ref.language }
func (ref TranscriptRef) Confidence() uint8 { return ref.confidence }

type Retention struct {
	until     time.Time
	legalHold bool
}

func NewRetention(until time.Time, legalHold bool) Retention {
	return Retention{until: until.UTC(), legalHold: legalHold}
}

func (retention Retention) Until() time.Time { return retention.until }
func (retention Retention) LegalHold() bool  { return retention.legalHold }

type Event struct {
	commandID   string
	fingerprint string
	action      Action
	occurredAt  time.Time
	version     uint64
}

func (event Event) CommandID() string     { return event.commandID }
func (event Event) Fingerprint() string   { return event.fingerprint }
func (event Event) Action() Action        { return event.action }
func (event Event) OccurredAt() time.Time { return event.occurredAt }
func (event Event) Version() uint64       { return event.version }

type Command struct {
	ID          string
	Fingerprint string
	At          time.Time
}

// Introduction is immutable. All state changes return a new value and append
// only enum/timestamp/keyed-fingerprint audit data.
// Prompt is one of the three questions a Voice of Introduction answers
// (S-06). The server owns this vocabulary because completeness became a
// server fact when it started earning a rung: counting recordings rather than
// prompts would let three takes of one question earn Tier 2.
type Prompt string

const (
	PromptArrival  Prompt = "arrival"
	PromptOrdinary Prompt = "ordinary"
	PromptWelcome  Prompt = "welcome"
)

// Prompts is the complete set, in the order a member is asked.
var Prompts = []Prompt{PromptArrival, PromptOrdinary, PromptWelcome}

// RecordedStatuses are the states in which a member's recording actually
// exists and counts toward a finished introduction.
//
// Draft and upload_authorized are promises rather than recordings; cancelled
// and revoked are recordings the member took back. Every transcription
// outcome is included on purpose: transcription is deferred and the
// configured provider reports uncertain for everything, so requiring "ready"
// would mean no member ever finished an introduction. A recording nobody
// could transcribe is still the member's voice.
func RecordedStatuses() []Status {
	return []Status{
		StatusUploaded, StatusTranscribing, StatusReady,
		StatusTranscriptionUncertain, StatusTranscriptionFailed,
	}
}

// Complete reports whether every prompt has a recording. This is what earns
// Tier 2, so it asks for all three questions rather than three recordings.
func Complete(recorded []Prompt) bool {
	seen := make(map[Prompt]bool, len(recorded))
	for _, prompt := range recorded {
		seen[prompt] = true
	}
	for _, required := range Prompts {
		if !seen[required] {
			return false
		}
	}
	return true
}

func (p Prompt) Valid() bool {
	for _, known := range Prompts {
		if known == p {
			return true
		}
	}
	return false
}

type Introduction struct {
	id      string
	ownerID string
	// prompt is which of the three questions this recording answers.
	prompt        Prompt
	consent       ConsentSnapshot
	media         MediaRef
	transcript    TranscriptRef
	status        Status
	dataStatus    DataStatus
	retention     Retention
	deletionDueAt time.Time
	createdAt     time.Time
	updatedAt     time.Time
	version       uint64
	events        []Event
}

func New(id, ownerID string, prompt Prompt, consent ConsentSnapshot, media MediaRef, retention Retention, command Command) (Introduction, error) {
	id = strings.TrimSpace(id)
	ownerID = strings.TrimSpace(ownerID)
	if !opaquePattern.MatchString(id) || !opaquePattern.MatchString(ownerID) ||
		!prompt.Valid() ||
		consent.purposeID == "" || media.assetID == "" || command.At.IsZero() ||
		(!retention.until.IsZero() && retention.until.Before(command.At.UTC())) {
		return Introduction{}, ErrInvalidIntroduction
	}
	introduction := Introduction{
		id: id, ownerID: ownerID, prompt: prompt, consent: consent, media: media,
		status: StatusDraft, dataStatus: DataRetained, retention: retention,
		createdAt: command.At.UTC(), updatedAt: command.At.UTC(), version: 1,
	}
	event, err := newEvent(command, ActionCreated, introduction.version)
	if err != nil {
		return Introduction{}, err
	}
	introduction.events = []Event{event}
	return introduction, nil
}

// NewEvent rebuilds one audit entry from storage. The aggregate's history is
// part of its identity — replay safety and the command-mismatch checks are
// both decided from it — so a store that could not restore events would
// restore an aggregate that accepts a command it has already applied.
func NewEvent(commandID, fingerprint string, action Action, occurredAt time.Time, version uint64) Event {
	return Event{
		commandID: commandID, fingerprint: fingerprint, action: action,
		occurredAt: occurredAt.UTC(), version: version,
	}
}

// ReconstituteParams carries stored state into Reconstitute. A struct rather
// than thirteen positional arguments: the four timestamps and two enums are
// indistinguishable by type, and a transposed pair would restore a plausible
// aggregate in the wrong state.
type ReconstituteParams struct {
	ID            string
	OwnerID       string
	Prompt        Prompt
	Consent       ConsentSnapshot
	Media         MediaRef
	Transcript    TranscriptRef
	Status        Status
	DataStatus    DataStatus
	Retention     Retention
	DeletionDueAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       uint64
	Events        []Event
}

// Reconstitute rebuilds a stored aggregate without re-running its invariants,
// which is correct only because they were enforced on the way in. It is for
// repositories; application code creates through New and transitions through
// the methods below.
func Reconstitute(params ReconstituteParams) Introduction {
	return Introduction{
		id: params.ID, ownerID: params.OwnerID, prompt: params.Prompt, consent: params.Consent,
		media: params.Media, transcript: params.Transcript,
		status: params.Status, dataStatus: params.DataStatus,
		retention: params.Retention, deletionDueAt: params.DeletionDueAt.UTC(),
		createdAt: params.CreatedAt.UTC(), updatedAt: params.UpdatedAt.UTC(),
		version: params.Version, events: append([]Event(nil), params.Events...),
	}
}

// ReconstituteConsentSnapshot, ReconstituteMediaRef and ReconstituteTranscriptRef
// rebuild the value objects from storage. The New* constructors validate, and
// a stored value that has already been validated must not be re-rejected by a
// rule that has since tightened — that would make old rows unreadable.
func ReconstituteConsentSnapshot(purposeID string, version uint64, evaluatedAt time.Time) ConsentSnapshot {
	return ConsentSnapshot{purposeID: purposeID, version: version, evaluatedAt: evaluatedAt.UTC()}
}

func ReconstituteMediaRef(assetID, contentType string, size int64, duration time.Duration, checksum string) MediaRef {
	return MediaRef{
		assetID: assetID, contentType: contentType, size: size,
		duration: duration, checksum: checksum,
	}
}

func ReconstituteTranscriptRef(id, language string, confidence uint8) TranscriptRef {
	return TranscriptRef{id: id, language: language, confidence: confidence}
}

func (introduction Introduction) AuthorizeUpload(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status != StatusDraft {
		return Introduction{}, ErrInvalidTransition
	}
	return introduction.transition(command, ActionUploadAuthorized, StatusUploadAuthorized, expectedVersion)
}

func (introduction Introduction) ConfirmUpload(media MediaRef, command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status != StatusUploadAuthorized || !media.Complete() ||
		media.AssetID() != introduction.media.AssetID() {
		return Introduction{}, ErrInvalidTransition
	}
	next, err := introduction.transition(command, ActionUploadConfirmed, StatusUploaded, expectedVersion)
	if err != nil {
		return Introduction{}, err
	}
	next.media = media
	return next, nil
}

func (introduction Introduction) StartTranscription(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status != StatusUploaded && introduction.status != StatusTranscriptionUncertain &&
		introduction.status != StatusTranscriptionFailed {
		return Introduction{}, ErrInvalidTransition
	}
	return introduction.transition(command, ActionTranscriptionStart, StatusTranscribing, expectedVersion)
}

func (introduction Introduction) CompleteTranscription(transcript TranscriptRef, command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status != StatusTranscribing || transcript.id == "" {
		return Introduction{}, ErrInvalidTransition
	}
	next, err := introduction.transition(command, ActionTranscriptionReady, StatusReady, expectedVersion)
	if err != nil {
		return Introduction{}, err
	}
	next.transcript = transcript
	return next, nil
}

func (introduction Introduction) TranscriptionUncertain(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status != StatusTranscribing {
		return Introduction{}, ErrInvalidTransition
	}
	return introduction.transition(command, ActionTranscriptionQueue, StatusTranscriptionUncertain, expectedVersion)
}

func (introduction Introduction) TranscriptionFailed(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status != StatusTranscribing {
		return Introduction{}, ErrInvalidTransition
	}
	return introduction.transition(command, ActionTranscriptionFail, StatusTranscriptionFailed, expectedVersion)
}

func (introduction Introduction) Cancel(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status == StatusCancelled || introduction.status == StatusRevoked ||
		introduction.dataStatus == DataPurged {
		return Introduction{}, ErrInvalidTransition
	}
	next, err := introduction.transition(command, ActionCancelled, StatusCancelled, expectedVersion)
	if err != nil {
		return Introduction{}, err
	}
	return next.scheduleDeletion(command.At), nil
}

func (introduction Introduction) Revoke(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.status == StatusRevoked || introduction.dataStatus == DataPurged {
		return Introduction{}, ErrInvalidTransition
	}
	next, err := introduction.transition(command, ActionRevoked, StatusRevoked, expectedVersion)
	if err != nil {
		return Introduction{}, err
	}
	return next.scheduleDeletion(command.At), nil
}

func (introduction Introduction) MarkPurged(command Command, expectedVersion uint64) (Introduction, error) {
	if introduction.retention.legalHold {
		return Introduction{}, ErrLegalHold
	}
	if introduction.dataStatus != DataPurgePending {
		return Introduction{}, ErrInvalidTransition
	}
	if command.At.UTC().Before(introduction.deletionDueAt) {
		return Introduction{}, ErrRetentionActive
	}
	next, err := introduction.transition(command, ActionPurged, introduction.status, expectedVersion)
	if err != nil {
		return Introduction{}, err
	}
	next.dataStatus = DataPurged
	next.media = MediaRef{}
	next.transcript = TranscriptRef{}
	return next, nil
}

func (introduction Introduction) transition(command Command, action Action, status Status, expectedVersion uint64) (Introduction, error) {
	if event, exists := introduction.command(command.ID); exists {
		if event.action == action && event.fingerprint == command.Fingerprint {
			return introduction, nil
		}
		return Introduction{}, ErrCommandConflict
	}
	if expectedVersion != introduction.version {
		return Introduction{}, ErrStaleVersion
	}
	event, err := newEvent(command, action, introduction.version+1)
	if err != nil {
		return Introduction{}, err
	}
	next := introduction
	next.status = status
	next.updatedAt = command.At.UTC()
	next.version++
	next.events = append(introduction.Events(), event)
	return next, nil
}

func (introduction Introduction) scheduleDeletion(now time.Time) Introduction {
	introduction.dataStatus = DataPurgePending
	if introduction.retention.legalHold {
		introduction.deletionDueAt = time.Time{}
		return introduction
	}
	introduction.deletionDueAt = now.UTC()
	if introduction.retention.until.After(introduction.deletionDueAt) {
		introduction.deletionDueAt = introduction.retention.until
	}
	return introduction
}

func newEvent(command Command, action Action, version uint64) (Event, error) {
	command.ID = strings.TrimSpace(command.ID)
	command.Fingerprint = strings.TrimSpace(command.Fingerprint)
	if !opaquePattern.MatchString(command.ID) || !validDigest(command.Fingerprint) || command.At.IsZero() {
		return Event{}, ErrInvalidIntroduction
	}
	return Event{
		commandID: command.ID, fingerprint: command.Fingerprint, action: action,
		occurredAt: command.At.UTC(), version: version,
	}, nil
}

func (introduction Introduction) command(id string) (Event, bool) {
	for _, event := range introduction.events {
		if event.commandID == id {
			return event, true
		}
	}
	return Event{}, false
}

func (introduction Introduction) ID() string                { return introduction.id }
func (introduction Introduction) OwnerID() string           { return introduction.ownerID }
func (introduction Introduction) Prompt() Prompt            { return introduction.prompt }
func (introduction Introduction) Consent() ConsentSnapshot  { return introduction.consent }
func (introduction Introduction) Media() MediaRef           { return introduction.media }
func (introduction Introduction) Transcript() TranscriptRef { return introduction.transcript }
func (introduction Introduction) Status() Status            { return introduction.status }
func (introduction Introduction) DataStatus() DataStatus    { return introduction.dataStatus }
func (introduction Introduction) Retention() Retention      { return introduction.retention }
func (introduction Introduction) DeletionDueAt() time.Time  { return introduction.deletionDueAt }
func (introduction Introduction) CreatedAt() time.Time      { return introduction.createdAt }
func (introduction Introduction) UpdatedAt() time.Time      { return introduction.updatedAt }
func (introduction Introduction) Version() uint64           { return introduction.version }
func (introduction Introduction) Events() []Event {
	return append([]Event(nil), introduction.events...)
}
func (introduction Introduction) HasCommand(id string) bool {
	_, exists := introduction.command(strings.TrimSpace(id))
	return exists
}

func (introduction Introduction) MatchesCommand(command Command, action Action) bool {
	event, exists := introduction.command(strings.TrimSpace(command.ID))
	return exists && event.action == action && event.fingerprint == strings.TrimSpace(command.Fingerprint)
}

func (introduction Introduction) SameCommand(other Introduction, id string) bool {
	left, leftExists := introduction.command(strings.TrimSpace(id))
	right, rightExists := other.command(strings.TrimSpace(id))
	return leftExists && rightExists && left.action == right.action &&
		left.fingerprint == right.fingerprint
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
