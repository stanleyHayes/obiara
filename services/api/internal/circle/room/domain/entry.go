package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const MaxRetention = 90 * 24 * time.Hour

var (
	ErrInvalidEntry = errors.New("invalid circle room entry")
	ErrRetention    = errors.New("invalid circle room retention")
)

type Kind string

const (
	KindVoice  Kind = "voice"
	KindEvent  Kind = "event"
	KindNotice Kind = "notice"
)

var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,159}$`)

type MediaRef struct {
	assetID, transcriptID, contentType string
	duration                           time.Duration
}

func NewMediaRef(assetID, transcriptID, contentType string, duration time.Duration) (MediaRef, error) {
	assetID, transcriptID, contentType = strings.TrimSpace(assetID), strings.TrimSpace(transcriptID), strings.TrimSpace(contentType)
	if !opaque.MatchString(assetID) || (transcriptID != "" && !opaque.MatchString(transcriptID)) ||
		!strings.HasPrefix(contentType, "audio/") || duration <= 0 {
		return MediaRef{}, ErrInvalidEntry
	}
	return MediaRef{assetID, transcriptID, contentType, duration}, nil
}
func (m MediaRef) AssetID() string         { return m.assetID }
func (m MediaRef) TranscriptID() string    { return m.transcriptID }
func (m MediaRef) ContentType() string     { return m.contentType }
func (m MediaRef) Duration() time.Duration { return m.duration }

type Audit struct {
	commandID, actorKey string
	at                  time.Time
}

func (a Audit) CommandID() string { return a.commandID }
func (a Audit) ActorKey() string  { return a.actorKey }
func (a Audit) At() time.Time     { return a.at }

type Entry struct {
	id, circleID, authorKey, contentRef               string
	kind                                              Kind
	media                                             MediaRef
	startsAt, endsAt, createdAt, expiresAt, deletedAt time.Time
	revision                                          uint64
	audit                                             []Audit
}
type Params struct {
	ID, CircleID, AuthorKey, ContentRef, CommandID string
	Kind                                           Kind
	Media                                          MediaRef
	StartsAt, EndsAt, CreatedAt, ExpiresAt         time.Time
}

func New(p Params) (Entry, error) {
	p.ID, p.CircleID, p.AuthorKey, p.ContentRef, p.CommandID = strings.TrimSpace(p.ID), strings.TrimSpace(p.CircleID), strings.TrimSpace(p.AuthorKey), strings.TrimSpace(p.ContentRef), strings.TrimSpace(p.CommandID)
	if !opaque.MatchString(p.ID) || !opaque.MatchString(p.CircleID) || !validKey(p.AuthorKey) || !opaque.MatchString(p.CommandID) || p.CreatedAt.IsZero() ||
		p.ExpiresAt.IsZero() || !p.ExpiresAt.After(p.CreatedAt) || p.ExpiresAt.Sub(p.CreatedAt) > MaxRetention {
		return Entry{}, ErrInvalidEntry
	}
	switch p.Kind {
	case KindVoice:
		if p.Media.assetID == "" || p.ContentRef != "" {
			return Entry{}, ErrInvalidEntry
		}
	case KindEvent:
		if !opaque.MatchString(p.ContentRef) || p.StartsAt.IsZero() || !p.EndsAt.After(p.StartsAt) {
			return Entry{}, ErrInvalidEntry
		}
	case KindNotice:
		if !opaque.MatchString(p.ContentRef) {
			return Entry{}, ErrInvalidEntry
		}
	default:
		return Entry{}, ErrInvalidEntry
	}
	return Entry{id: p.ID, circleID: p.CircleID, authorKey: p.AuthorKey, contentRef: p.ContentRef, kind: p.Kind, media: p.Media,
		startsAt: p.StartsAt.UTC(), endsAt: p.EndsAt.UTC(), createdAt: p.CreatedAt.UTC(), expiresAt: p.ExpiresAt.UTC(), revision: 1,
		audit: []Audit{{p.CommandID, p.AuthorKey, p.CreatedAt.UTC()}}}, nil
}
func Rehydrate(p Params, deletedAt time.Time, revision uint64, audit []Audit) (Entry, error) {
	e, err := New(p)
	if err != nil {
		return Entry{}, err
	}
	if revision == 0 || uint64(len(audit)) != revision {
		return Entry{}, ErrInvalidEntry
	}
	e.deletedAt = deletedAt.UTC()
	e.revision = revision
	e.audit = append([]Audit(nil), audit...)
	return e, nil
}
func NewAudit(commandID, actorKey string, at time.Time) (Audit, error) {
	if !opaque.MatchString(commandID) || !validKey(actorKey) || at.IsZero() {
		return Audit{}, ErrInvalidEntry
	}
	return Audit{commandID, actorKey, at.UTC()}, nil
}
func (e Entry) Delete(commandID, actorKey string, at time.Time, expected uint64) (Entry, error) {
	for _, a := range e.audit {
		if a.commandID == commandID {
			return e, nil
		}
	}
	if expected != e.revision || !e.deletedAt.IsZero() {
		return Entry{}, ErrInvalidEntry
	}
	a, err := NewAudit(commandID, actorKey, at)
	if err != nil {
		return Entry{}, err
	}
	e.deletedAt = at.UTC()
	e.revision++
	e.audit = append(e.Audit(), a)
	return e, nil
}
func (e Entry) Visible(at time.Time) bool {
	return e.deletedAt.IsZero() && at.UTC().Before(e.expiresAt)
}
func (e Entry) ID() string           { return e.id }
func (e Entry) CircleID() string     { return e.circleID }
func (e Entry) AuthorKey() string    { return e.authorKey }
func (e Entry) ContentRef() string   { return e.contentRef }
func (e Entry) Kind() Kind           { return e.kind }
func (e Entry) Media() MediaRef      { return e.media }
func (e Entry) StartsAt() time.Time  { return e.startsAt }
func (e Entry) EndsAt() time.Time    { return e.endsAt }
func (e Entry) CreatedAt() time.Time { return e.createdAt }
func (e Entry) ExpiresAt() time.Time { return e.expiresAt }
func (e Entry) DeletedAt() time.Time { return e.deletedAt }
func (e Entry) Revision() uint64     { return e.revision }
func (e Entry) Audit() []Audit       { return append([]Audit(nil), e.audit...) }
func validKey(v string) bool {
	return v == "system" || (len(v) == 64 && regexp.MustCompile(`^[a-f0-9]+$`).MatchString(v))
}
