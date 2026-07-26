package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"time"
)

type GameEvent string

const (
	EventAbandoned            GameEvent = "session_abandoned"
	EventInactivity           GameEvent = "repeated_inactivity"
	EventModerationRemoval    GameEvent = "moderation_removal"
	EventRespectfulCompletion GameEvent = "respectful_completion"
)

type Reason string

const (
	ReasonAbandonment Reason = "conduct_abandonment"
	ReasonInactivity  Reason = "conduct_inactivity"
	ReasonModerated   Reason = "conduct_moderated"
	ReasonRespectful  Reason = "conduct_respectful"
)

type Provenance string

const (
	ProvenanceServerEvent     Provenance = "server_event"
	ProvenanceModeratorAction Provenance = "moderator_action"
)

type Kind string

const (
	KindConcern     Kind = "concern"
	KindAffirmation Kind = "affirmation"
)

type AppealState string

const (
	AppealNone       AppealState = "none"
	AppealPending    AppealState = "pending"
	AppealUpheld     AppealState = "upheld"
	AppealOverturned AppealState = "overturned"
)

var (
	ErrInvalid    = errors.New("invalid conduct signal")
	ErrTransition = errors.New("invalid conduct transition")
	ErrStale      = errors.New("stale conduct signal")
	ErrMismatch   = errors.New("conduct command mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Command struct {
	ID               string
	ExpectedRevision uint64
	At               time.Time
}
type Event struct {
	Sequence          uint64 `bson:"sequence"`
	CommandID, Action string
	At                time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Projection struct {
	Reference  string
	Kind       Kind
	Reason     Reason
	Provenance Provenance
	Appeal     AppealState
	RecordedAt time.Time
}
type Signal struct {
	id, gameKey, subjectKey, eventKey string
	gameEvent                         GameEvent
	kind                              Kind
	reason                            Reason
	provenance                        Provenance
	recordedAt                        time.Time
	appeal                            AppealState
	appealedAt, resolvedAt            time.Time
	revision                          uint64
	events                            []Event
	commands                          []Applied
}
type State struct {
	ID, GameKey, SubjectKey, EventKey string
	GameEvent                         GameEvent
	Kind                              Kind
	Reason                            Reason
	Provenance                        Provenance
	RecordedAt                        time.Time
	Appeal                            AppealState
	AppealedAt, ResolvedAt            time.Time
	Revision                          uint64
	Events                            []Event
	Commands                          []Applied
}

func Record(id, game, subject, event string, eventType GameEvent, at time.Time, c Command) (Signal, error) {
	kind, reason, provenance, ok := mapping(eventType)
	s := Signal{id: id, gameKey: game, subjectKey: subject, eventKey: event, gameEvent: eventType, kind: kind, reason: reason, provenance: provenance, recordedAt: at.UTC(), appeal: AppealNone}
	if !ok || !valid(s) || c.ExpectedRevision != 0 {
		return Signal{}, ErrInvalid
	}
	if e := s.apply(c, "record"); e != nil {
		return Signal{}, e
	}
	return s, nil
}
func Rehydrate(st State) (Signal, error) {
	s := Signal{id: st.ID, gameKey: st.GameKey, subjectKey: st.SubjectKey, eventKey: st.EventKey, gameEvent: st.GameEvent, kind: st.Kind, reason: st.Reason, provenance: st.Provenance, recordedAt: st.RecordedAt.UTC(), appeal: st.Appeal, appealedAt: st.AppealedAt.UTC(), resolvedAt: st.ResolvedAt.UTC(), revision: st.Revision, events: append([]Event(nil), st.Events...), commands: append([]Applied(nil), st.Commands...)}
	k, r, p, ok := mapping(s.gameEvent)
	if !ok || !valid(s) || k != s.kind || r != s.reason || p != s.provenance || len(s.events) != int(s.revision) || len(s.commands) != int(s.revision) || s.revision == 0 {
		return Signal{}, ErrInvalid
	}
	return s, nil
}
func (s Signal) Appeal(now time.Time, c Command) (Signal, error) {
	if replay, e := s.replay(c, "appeal"); replay || e != nil {
		return s, e
	}
	if s.appeal != AppealNone {
		return Signal{}, ErrTransition
	}
	s.appeal, s.appealedAt = AppealPending, now.UTC()
	return s, s.apply(c, "appeal")
}
func (s Signal) Resolve(result AppealState, now time.Time, c Command) (Signal, error) {
	action := "resolve:" + string(result)
	if replay, e := s.replay(c, action); replay || e != nil {
		return s, e
	}
	if s.appeal != AppealPending || (result != AppealUpheld && result != AppealOverturned) {
		return Signal{}, ErrTransition
	}
	s.appeal, s.resolvedAt = result, now.UTC()
	return s, s.apply(c, action)
}
func (s Signal) Project() Projection {
	return Projection{Reference: s.id, Kind: s.kind, Reason: s.reason, Provenance: s.provenance, Appeal: s.appeal, RecordedAt: s.recordedAt}
}
func (s *Signal) apply(c Command, a string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != s.revision {
		return ErrStale
	}
	f := fingerprint(s.id, c, a)
	s.revision++
	s.events = append(s.events, Event{s.revision, c.ID, a, c.At.UTC()})
	s.commands = append(s.commands, Applied{c.ID, f, s.revision})
	return nil
}
func (s Signal) replay(c Command, a string) (bool, error) {
	f := fingerprint(s.id, c, a)
	for _, x := range s.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func mapping(e GameEvent) (Kind, Reason, Provenance, bool) {
	switch e {
	case EventAbandoned:
		return KindConcern, ReasonAbandonment, ProvenanceServerEvent, true
	case EventInactivity:
		return KindConcern, ReasonInactivity, ProvenanceServerEvent, true
	case EventModerationRemoval:
		return KindConcern, ReasonModerated, ProvenanceModeratorAction, true
	case EventRespectfulCompletion:
		return KindAffirmation, ReasonRespectful, ProvenanceServerEvent, true
	}
	return "", "", "", false
}
func valid(s Signal) bool {
	return opaque.MatchString(s.id) && key.MatchString(s.gameKey) && key.MatchString(s.subjectKey) && key.MatchString(s.eventKey) && !s.recordedAt.IsZero() && (s.appeal == AppealNone || s.appeal == AppealPending || s.appeal == AppealUpheld || s.appeal == AppealOverturned)
}
func fingerprint(id string, c Command, a string) string {
	x := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(x[:])
}
func (s Signal) ID() string               { return s.id }
func (s Signal) GameKey() string          { return s.gameKey }
func (s Signal) SubjectKey() string       { return s.subjectKey }
func (s Signal) EventKey() string         { return s.eventKey }
func (s Signal) GameEvent() GameEvent     { return s.gameEvent }
func (s Signal) Kind() Kind               { return s.kind }
func (s Signal) Reason() Reason           { return s.reason }
func (s Signal) Provenance() Provenance   { return s.provenance }
func (s Signal) RecordedAt() time.Time    { return s.recordedAt }
func (s Signal) AppealState() AppealState { return s.appeal }
func (s Signal) AppealedAt() time.Time    { return s.appealedAt }
func (s Signal) ResolvedAt() time.Time    { return s.resolvedAt }
func (s Signal) Revision() uint64         { return s.revision }
func (s Signal) Events() []Event          { return append([]Event(nil), s.events...) }
func (s Signal) Commands() []Applied      { return append([]Applied(nil), s.commands...) }
