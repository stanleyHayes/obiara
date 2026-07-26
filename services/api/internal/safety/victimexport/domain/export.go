package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	MaxReferences    = 20
	AuthorizationTTL = 72 * time.Hour
)

var (
	ErrInvalid    = errors.New("invalid victim export")
	ErrStale      = errors.New("stale victim export command")
	ErrMismatch   = errors.New("victim export command mismatch")
	ErrTransition = errors.New("invalid victim export transition")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Purpose string

const (
	PurposeVictimSupport  Purpose = "victim_support"
	PurposeLegalSupport   Purpose = "legal_support"
	PurposePersonalRecord Purpose = "personal_record"
)

type ReferenceKind string

const (
	KindIncidentSummary      ReferenceKind = "incident_summary"
	KindMessageMetadata      ReferenceKind = "message_metadata"
	KindTransactionReference ReferenceKind = "transaction_reference"
)

func validPurpose(p Purpose) bool {
	return p == PurposeVictimSupport || p == PurposeLegalSupport || p == PurposePersonalRecord
}
func validKind(k ReferenceKind) bool {
	return k == KindIncidentSummary || k == KindMessageMetadata || k == KindTransactionReference
}

type Reference struct {
	Kind         ReferenceKind `bson:"kind"`
	RefKey       string        `bson:"refKey"`
	RedactionKey string        `bson:"redactionKey,omitempty"`
}
type Status string

const (
	StatusRequested  Status = "requested"
	StatusAuthorized Status = "authorized"
	StatusUsed       Status = "used"
	StatusRevoked    Status = "revoked"
)

type Authorization struct {
	TokenKey     string    `bson:"tokenKey"`
	AuthorizedAt time.Time `bson:"authorizedAt"`
	ExpiresAt    time.Time `bson:"expiresAt"`
	UsedAt       time.Time `bson:"usedAt,omitempty"`
	RevokedAt    time.Time `bson:"revokedAt,omitempty"`
}
type Command struct {
	ID               string    `bson:"id"`
	ExpectedRevision uint64    `bson:"expectedRevision"`
	At               time.Time `bson:"at"`
}
type Event struct {
	Sequence  uint64    `bson:"sequence"`
	CommandID string    `bson:"commandId"`
	Action    string    `bson:"action"`
	At        time.Time `bson:"at"`
}
type Applied struct {
	ID          string `bson:"id"`
	Fingerprint string `bson:"fingerprint"`
	Revision    uint64 `bson:"revision"`
}
type Export struct {
	id, memberKey string
	purpose       Purpose
	refs          []Reference
	status        Status
	auth          Authorization
	revision      uint64
	events        []Event
	commands      []Applied
}
type State struct {
	ID, MemberKey string
	Purpose       Purpose
	References    []Reference
	Status        Status
	Authorization Authorization
	Revision      uint64
	Events        []Event
	Commands      []Applied
}

func Request(id, member string, p Purpose, refs []Reference, c Command) (Export, error) {
	r := normalized(refs, true)
	e := Export{id: id, memberKey: member, purpose: p, refs: r, status: StatusRequested}
	if !opaque.MatchString(id) || !key.MatchString(member) || !validPurpose(p) || !validRefs(r, true) || c.ExpectedRevision != 0 {
		return Export{}, ErrInvalid
	}
	if x := e.apply(c, "request"); x != nil {
		return Export{}, x
	}
	return e, nil
}
func Rehydrate(s State) (Export, error) {
	e := Export{id: s.ID, memberKey: s.MemberKey, purpose: s.Purpose, refs: normalized(s.References, true), status: s.Status, auth: s.Authorization, revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]Applied(nil), s.Commands...)}
	if !opaque.MatchString(e.id) || !key.MatchString(e.memberKey) || !validPurpose(e.purpose) || !validRefs(e.refs, true) || !validState(e.status, e.auth) || e.revision == 0 || len(e.events) != int(e.revision) || len(e.commands) != int(e.revision) {
		return Export{}, ErrInvalid
	}
	return e, nil
}
func (e Export) Authorize(redacted []Reference, token string, c Command) (Export, error) {
	if replay, x := e.replay(c, "authorize", token); replay || x != nil {
		return e, x
	}
	r := normalized(redacted, true)
	if e.status != StatusRequested || !key.MatchString(token) || !validRefs(r, true) || !sameRequested(e.refs, r) {
		return Export{}, ErrTransition
	}
	e = e.clone()
	e.refs = r
	e.status = StatusAuthorized
	e.auth = Authorization{TokenKey: token, AuthorizedAt: c.At.UTC(), ExpiresAt: c.At.UTC().Add(AuthorizationTTL)}
	return e, e.apply(c, "authorize", token)
}
func (e Export) Use(token string, c Command) (Export, error) {
	if replay, x := e.replay(c, "use", token); replay || x != nil {
		return e, x
	}
	if e.status != StatusAuthorized || e.auth.TokenKey != token || !c.At.Before(e.auth.ExpiresAt) {
		return Export{}, ErrTransition
	}
	e = e.clone()
	e.status = StatusUsed
	e.auth.UsedAt = c.At.UTC()
	return e, e.apply(c, "use", token)
}
func (e Export) Revoke(c Command) (Export, error) {
	if replay, x := e.replay(c, "revoke"); replay || x != nil {
		return e, x
	}
	if e.status != StatusAuthorized {
		return Export{}, ErrTransition
	}
	e = e.clone()
	e.status = StatusRevoked
	e.auth.RevokedAt = c.At.UTC()
	return e, e.apply(c, "revoke")
}
func normalized(v []Reference, redacted bool) []Reference {
	r := append([]Reference(nil), v...)
	slices.SortFunc(r, func(a, b Reference) int {
		if a.Kind != b.Kind {
			return strings.Compare(string(a.Kind), string(b.Kind))
		}
		return strings.Compare(a.RefKey, b.RefKey)
	})
	return r
}
func validRefs(v []Reference, redacted bool) bool {
	if len(v) == 0 || len(v) > MaxReferences {
		return false
	}
	for i, r := range v {
		if !validKind(r.Kind) || !key.MatchString(r.RefKey) || (redacted && !key.MatchString(r.RedactionKey)) || (!redacted && r.RedactionKey != "") || (i > 0 && v[i-1].Kind == r.Kind && v[i-1].RefKey == r.RefKey) {
			return false
		}
	}
	return true
}
func sameRequested(a, b []Reference) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func validState(status Status, auth Authorization) bool {
	switch status {
	case StatusRequested:
		return auth == (Authorization{})
	case StatusAuthorized:
		return key.MatchString(auth.TokenKey) && !auth.AuthorizedAt.IsZero() && auth.ExpiresAt.Sub(auth.AuthorizedAt) == AuthorizationTTL && auth.UsedAt.IsZero() && auth.RevokedAt.IsZero()
	case StatusUsed:
		return key.MatchString(auth.TokenKey) && !auth.UsedAt.IsZero() && auth.RevokedAt.IsZero()
	case StatusRevoked:
		return key.MatchString(auth.TokenKey) && auth.UsedAt.IsZero() && !auth.RevokedAt.IsZero()
	default:
		return false
	}
}
func (e Export) clone() Export {
	e.refs = append([]Reference(nil), e.refs...)
	e.events = append([]Event(nil), e.events...)
	e.commands = append([]Applied(nil), e.commands...)
	return e
}
func (e *Export) apply(c Command, a string, v ...string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != e.revision {
		return ErrStale
	}
	f := fingerprint(e.id, c, a, v...)
	e.revision++
	e.events = append(e.events, Event{e.revision, c.ID, a, c.At.UTC()})
	e.commands = append(e.commands, Applied{c.ID, f, e.revision})
	return nil
}
func (e Export) replay(c Command, a string, v ...string) (bool, error) {
	f := fingerprint(e.id, c, a, v...)
	for _, x := range e.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func fingerprint(id string, c Command, a string, v ...string) string {
	s := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strings.Join(v, "\x00") + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(s[:])
}
func (e Export) ID() string                   { return e.id }
func (e Export) MemberKey() string            { return e.memberKey }
func (e Export) Purpose() Purpose             { return e.purpose }
func (e Export) References() []Reference      { return append([]Reference(nil), e.refs...) }
func (e Export) Status() Status               { return e.status }
func (e Export) Authorization() Authorization { return e.auth }
func (e Export) Revision() uint64             { return e.revision }
func (e Export) Events() []Event              { return append([]Event(nil), e.events...) }
func (e Export) Commands() []Applied          { return append([]Applied(nil), e.commands...) }
