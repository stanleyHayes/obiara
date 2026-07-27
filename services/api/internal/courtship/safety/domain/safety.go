package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"slices"
	"strconv"
	"time"
)

type Category string
type Action string

const (
	CategoryHarassment Category = "harassment"
	CategoryIdentity   Category = "identity"
	CategoryThreat     Category = "threat"
	CategoryOther      Category = "other"
	ActionBlock        Action   = "block"
	ActionReport       Action   = "report"
)

var (
	ErrInvalid         = errors.New("invalid safety command")
	ErrDenied          = errors.New("safety command unavailable")
	ErrContactBlocked  = errors.New("room contact unavailable")
	ErrStaleRevision   = errors.New("stale safety revision")
	ErrCommandMismatch = errors.New("safety command replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)
	evidencePattern    = regexp.MustCompile(`^enc_[A-Za-z0-9_-]{20,200}$`)
)

type Command struct {
	ID, ActorKey, EvidenceRef string
	Category                  Category
	ExpectedRevision          uint64
	At                        time.Time
}
type Event struct {
	Sequence  uint64
	CommandID string
	Action    Action
	At        time.Time
}
type CareReview struct {
	ID, ReporterKey, EvidenceRef string
	Category                     Category
	CreatedAt                    time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Safety struct {
	id        string
	members   []string
	blocked   bool
	blockedAt time.Time
	revision  uint64
	events    []Event
	reviews   []CareReview
	commands  []Applied
}
type State struct {
	ID        string
	Members   []string
	Blocked   bool
	BlockedAt time.Time
	Revision  uint64
	Events    []Event
	Reviews   []CareReview
	Commands  []Applied
}

func New(id string, members []string) (Safety, error) {
	m := append([]string(nil), members...)
	slices.Sort(m)
	m = slices.Compact(m)
	if !idPattern.MatchString(id) || len(m) != 2 || !keyPattern.MatchString(m[0]) || !keyPattern.MatchString(m[1]) {
		return Safety{}, ErrInvalid
	}
	return Safety{id: id, members: m}, nil
}
func Rehydrate(s State) (Safety, error) {
	v, err := New(s.ID, s.Members)
	if err != nil {
		return Safety{}, err
	}
	v.blocked, v.blockedAt, v.revision = s.Blocked, s.BlockedAt, s.Revision
	v.events = append([]Event(nil), s.Events...)
	v.reviews = append([]CareReview(nil), s.Reviews...)
	v.commands = append([]Applied(nil), s.Commands...)
	if len(v.events) != int(v.revision) || len(v.commands) != int(v.revision) || (!v.blocked && !v.blockedAt.IsZero()) || (v.blocked && v.blockedAt.IsZero()) {
		return Safety{}, ErrInvalid
	}
	for _, r := range v.reviews {
		if !v.member(r.ReporterKey) || !validCategory(r.Category) || !evidencePattern.MatchString(r.EvidenceRef) || r.CreatedAt.IsZero() {
			return Safety{}, ErrInvalid
		}
	}
	return v, nil
}
func (s Safety) Block(cmd Command) (Safety, error)  { return s.apply(ActionBlock, cmd) }
func (s Safety) Report(cmd Command) (Safety, error) { return s.apply(ActionReport, cmd) }
func (s Safety) apply(action Action, cmd Command) (Safety, error) {
	if !s.member(cmd.ActorKey) || !idPattern.MatchString(cmd.ID) || cmd.At.IsZero() {
		return Safety{}, ErrDenied
	}
	if action == ActionBlock && (cmd.Category != "" || cmd.EvidenceRef != "") {
		return Safety{}, ErrInvalid
	}
	if action == ActionReport && (!validCategory(cmd.Category) || !evidencePattern.MatchString(cmd.EvidenceRef)) {
		return Safety{}, ErrInvalid
	}
	fp := fingerprint(s.id, action, cmd)
	for _, seen := range s.commands {
		if seen.ID == cmd.ID {
			if seen.Fingerprint != fp {
				return Safety{}, ErrCommandMismatch
			}
			return s, nil
		}
	}
	if cmd.ExpectedRevision != s.revision {
		return Safety{}, ErrStaleRevision
	}
	if action == ActionBlock && s.blocked {
		return Safety{}, ErrContactBlocked
	}
	s.revision++
	if action == ActionBlock {
		s.blocked = true
		s.blockedAt = cmd.At.UTC()
	}
	if action == ActionReport {
		s.reviews = append(append([]CareReview(nil), s.reviews...), CareReview{ID: cmd.ID, ReporterKey: cmd.ActorKey, EvidenceRef: cmd.EvidenceRef, Category: cmd.Category, CreatedAt: cmd.At.UTC()})
	}
	s.events = append(append([]Event(nil), s.events...), Event{Sequence: s.revision, CommandID: cmd.ID, Action: action, At: cmd.At.UTC()})
	s.commands = append(append([]Applied(nil), s.commands...), Applied{ID: cmd.ID, Fingerprint: fp, Revision: s.revision})
	return s, nil
}
func validCategory(c Category) bool {
	return c == CategoryHarassment || c == CategoryIdentity || c == CategoryThreat || c == CategoryOther
}
func fingerprint(id string, action Action, c Command) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + string(c.Category) + "\x00" + c.EvidenceRef + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(sum[:])
}
func (s Safety) member(k string) bool { return slices.Contains(s.members, k) }
func (s Safety) CanContact(actor string) error {
	if !s.member(actor) {
		return ErrDenied
	}
	if s.blocked {
		return ErrContactBlocked
	}
	return nil
}
func (s Safety) ID() string            { return s.id }
func (s Safety) Members() []string     { return append([]string(nil), s.members...) }
func (s Safety) Blocked() bool         { return s.blocked }
func (s Safety) BlockedAt() time.Time  { return s.blockedAt }
func (s Safety) Revision() uint64      { return s.revision }
func (s Safety) Events() []Event       { return append([]Event(nil), s.events...) }
func (s Safety) Reviews() []CareReview { return append([]CareReview(nil), s.reviews...) }
func (s Safety) Commands() []Applied   { return append([]Applied(nil), s.commands...) }
