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

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
	StatusDeleted  Status = "deleted"
)

var (
	ErrInvalid         = errors.New("invalid cloth lifecycle")
	ErrDenied          = errors.New("cloth lifecycle unavailable")
	ErrStaleRevision   = errors.New("stale lifecycle revision")
	ErrCommandMismatch = errors.New("lifecycle command replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)
	refPattern         = regexp.MustCompile(`^ref_[A-Za-z0-9_-]{20,160}$`)
)

type Command struct {
	ID, ActorKey, ArchiveRef, ReceiptKey string
	ExpectedRevision                     uint64
	At                                   time.Time
}
type Provenance struct {
	BandVersion uint32
	RecipeRef   string
}
type Tombstone struct {
	ReceiptKey string
	DeletedAt  time.Time
	Provenance Provenance
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Lifecycle struct {
	id         string
	members    []string
	status     Status
	provenance Provenance
	archiveRef string
	archivedAt time.Time
	tombstone  *Tombstone
	revision   uint64
	commands   []Applied
}
type State struct {
	ID         string
	Members    []string
	Status     Status
	Provenance Provenance
	ArchiveRef string
	ArchivedAt time.Time
	Tombstone  *Tombstone
	Revision   uint64
	Commands   []Applied
}
type Export struct {
	ArchiveRef string
	Provenance Provenance
	ArchivedAt time.Time
}

func New(id string, members []string, p Provenance) (Lifecycle, error) {
	m := append([]string(nil), members...)
	slices.Sort(m)
	m = slices.Compact(m)
	if !idPattern.MatchString(id) || len(m) != 2 || !keyPattern.MatchString(m[0]) || !keyPattern.MatchString(m[1]) || p.BandVersion == 0 || !refPattern.MatchString(p.RecipeRef) {
		return Lifecycle{}, ErrInvalid
	}
	return Lifecycle{id: id, members: m, status: StatusActive, provenance: p}, nil
}
func Rehydrate(s State) (Lifecycle, error) {
	v, e := New(s.ID, s.Members, s.Provenance)
	if e != nil {
		return Lifecycle{}, e
	}
	v.status, v.archiveRef, v.archivedAt, v.tombstone, v.revision, v.commands = s.Status, s.ArchiveRef, s.ArchivedAt, s.Tombstone, s.Revision, append([]Applied(nil), s.Commands...)
	if len(v.commands) != int(v.revision) || v.status == StatusActive && v.revision != 0 || v.status == StatusArchived && (!refPattern.MatchString(v.archiveRef) || v.archivedAt.IsZero()) || v.status == StatusDeleted && (v.tombstone == nil || !keyPattern.MatchString(v.tombstone.ReceiptKey)) {
		return Lifecycle{}, ErrInvalid
	}
	return v, nil
}
func (v Lifecycle) Archive(c Command) (Lifecycle, error) {
	if !v.member(c.ActorKey) || !refPattern.MatchString(c.ArchiveRef) || c.ReceiptKey != "" {
		return Lifecycle{}, ErrDenied
	}
	return v.apply("archive", c)
}
func (v Lifecycle) Delete(c Command) (Lifecycle, error) {
	if !v.member(c.ActorKey) || !keyPattern.MatchString(c.ReceiptKey) {
		return Lifecycle{}, ErrDenied
	}
	return v.apply("delete", c)
}
func (v Lifecycle) apply(action string, c Command) (Lifecycle, error) {
	if !idPattern.MatchString(c.ID) || c.At.IsZero() {
		return Lifecycle{}, ErrDenied
	}
	fp := fingerprint(v.id, action, c)
	for _, x := range v.commands {
		if x.ID == c.ID {
			if x.Fingerprint != fp {
				return Lifecycle{}, ErrCommandMismatch
			}
			return v, nil
		}
	}
	if c.ExpectedRevision != v.revision {
		return Lifecycle{}, ErrStaleRevision
	}
	if action == "archive" {
		if v.status != StatusActive {
			return Lifecycle{}, ErrDenied
		}
		v.status = StatusArchived
		v.archiveRef = c.ArchiveRef
		v.archivedAt = c.At.UTC()
	} else {
		if v.status == StatusDeleted {
			return Lifecycle{}, ErrDenied
		}
		v.status = StatusDeleted
		v.tombstone = &Tombstone{ReceiptKey: c.ReceiptKey, DeletedAt: c.At.UTC(), Provenance: v.provenance}
		v.archiveRef = ""
		v.archivedAt = time.Time{}
	}
	v.revision++
	v.commands = append(v.commands, Applied{c.ID, fp, v.revision})
	return v, nil
}
func (v Lifecycle) Export(actor, archiveRef string) (Export, error) {
	if !v.member(actor) || v.status != StatusArchived || archiveRef != v.archiveRef {
		return Export{}, ErrDenied
	}
	return Export{v.archiveRef, v.provenance, v.archivedAt}, nil
}
func fingerprint(id, action string, c Command) string {
	s := sha256.Sum256([]byte(id + "\x00" + action + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + c.ArchiveRef + "\x00" + c.ReceiptKey + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(s[:])
}
func (v Lifecycle) member(k string) bool   { return slices.Contains(v.members, k) }
func (v Lifecycle) ID() string             { return v.id }
func (v Lifecycle) Members() []string      { return append([]string(nil), v.members...) }
func (v Lifecycle) Status() Status         { return v.status }
func (v Lifecycle) Provenance() Provenance { return v.provenance }
func (v Lifecycle) ArchiveRef() string     { return v.archiveRef }
func (v Lifecycle) ArchivedAt() time.Time  { return v.archivedAt }
func (v Lifecycle) Tombstone() *Tombstone {
	if v.tombstone == nil {
		return nil
	}
	t := *v.tombstone
	return &t
}
func (v Lifecycle) Revision() uint64    { return v.revision }
func (v Lifecycle) Commands() []Applied { return append([]Applied(nil), v.commands...) }
