package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"time"
)

type Category string

const (
	CategoryThreat     Category = "threat"
	CategoryHarassment Category = "harassment"
	CategoryIdentity   Category = "identity"
	CategoryMedical    Category = "medical"
	CategoryOther      Category = "other"
)

type Status string

const (
	StatusPending Status = "pending_route"
	StatusRouted  Status = "routed"
)

var (
	ErrInvalid    = errors.New("invalid incident")
	ErrTransition = errors.New("invalid incident transition")
	ErrStale      = errors.New("stale incident revision")
	ErrMismatch   = errors.New("incident command mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Command struct {
	ID               string
	ExpectedRevision uint64
	At               time.Time
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
type Case struct {
	CaseID, FireKey, ActorKey string
	Category                  Category
	EvidenceRef               string
	OccurredAt                time.Time
}
type Projection struct {
	Reference  string
	AcceptedAt time.Time
	Routed     bool
}
type Incident struct {
	caseID, fireKey, actorKey, evidenceRef string
	category                               Category
	occurredAt                             time.Time
	status                                 Status
	routedAt                               time.Time
	revision                               uint64
	events                                 []Event
	commands                               []Applied
}
type State struct {
	CaseID, FireKey, ActorKey, EvidenceRef string
	Category                               Category
	OccurredAt                             time.Time
	Status                                 Status
	RoutedAt                               time.Time
	Revision                               uint64
	Events                                 []Event
	Commands                               []Applied
}

func Create(caseID, fireKey, actorKey string, category Category, evidence string, occurred time.Time, c Command) (Incident, error) {
	i := Incident{caseID: caseID, fireKey: fireKey, actorKey: actorKey, category: category, evidenceRef: evidence, occurredAt: occurred.UTC(), status: StatusPending}
	if !valid(i) || c.ExpectedRevision != 0 {
		return Incident{}, ErrInvalid
	}
	if e := i.apply(c, "create"); e != nil {
		return Incident{}, e
	}
	return i, nil
}
func Rehydrate(s State) (Incident, error) {
	i := Incident{caseID: s.CaseID, fireKey: s.FireKey, actorKey: s.ActorKey, category: s.Category, evidenceRef: s.EvidenceRef, occurredAt: s.OccurredAt.UTC(), status: s.Status, routedAt: s.RoutedAt.UTC(), revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]Applied(nil), s.Commands...)}
	if !valid(i) || len(i.events) != int(i.revision) || len(i.commands) != int(i.revision) || i.revision == 0 || (i.status == StatusRouted && i.routedAt.IsZero()) {
		return Incident{}, ErrInvalid
	}
	return i, nil
}
func (i Incident) Route(now time.Time, c Command) (Incident, error) {
	if replay, e := i.replay(c, "route"); replay || e != nil {
		return i, e
	}
	if i.status != StatusPending || now.IsZero() {
		return Incident{}, ErrTransition
	}
	i.status, i.routedAt = StatusRouted, now.UTC()
	return i, i.apply(c, "route")
}
func (i Incident) Case() Case {
	return Case{i.caseID, i.fireKey, i.actorKey, i.category, i.evidenceRef, i.occurredAt}
}
func (i Incident) Project() Projection {
	return Projection{Reference: i.caseID, AcceptedAt: i.occurredAt, Routed: i.status == StatusRouted}
}
func (i *Incident) apply(c Command, action string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != i.revision {
		return ErrStale
	}
	f := fingerprint(i.caseID, c, action)
	i.revision++
	i.events = append(i.events, Event{i.revision, c.ID, action, c.At.UTC()})
	i.commands = append(i.commands, Applied{c.ID, f, i.revision})
	return nil
}
func (i Incident) replay(c Command, a string) (bool, error) {
	f := fingerprint(i.caseID, c, a)
	for _, x := range i.commands {
		if x.ID == c.ID {
			if x.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func valid(i Incident) bool {
	return opaque.MatchString(i.caseID) && key.MatchString(i.fireKey) && key.MatchString(i.actorKey) && (i.evidenceRef == "" || key.MatchString(i.evidenceRef)) && !i.occurredAt.IsZero() && validCategory(i.category) && (i.status == StatusPending || i.status == StatusRouted)
}
func validCategory(c Category) bool {
	return c == CategoryThreat || c == CategoryHarassment || c == CategoryIdentity || c == CategoryMedical || c == CategoryOther
}
func fingerprint(id string, c Command, a string) string {
	s := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(s[:])
}
func (i Incident) CaseID() string        { return i.caseID }
func (i Incident) FireKey() string       { return i.fireKey }
func (i Incident) ActorKey() string      { return i.actorKey }
func (i Incident) Category() Category    { return i.category }
func (i Incident) EvidenceRef() string   { return i.evidenceRef }
func (i Incident) OccurredAt() time.Time { return i.occurredAt }
func (i Incident) Status() Status        { return i.status }
func (i Incident) RoutedAt() time.Time   { return i.routedAt }
func (i Incident) Revision() uint64      { return i.revision }
func (i Incident) Events() []Event       { return append([]Event(nil), i.events...) }
func (i Incident) Commands() []Applied   { return append([]Applied(nil), i.commands...) }
