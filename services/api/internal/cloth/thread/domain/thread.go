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

var (
	ErrInvalid         = errors.New("invalid thread")
	ErrDenied          = errors.New("thread unavailable")
	ErrAlreadyIssued   = errors.New("thread unavailable")
	ErrStaleRevision   = errors.New("stale thread revision")
	ErrCommandMismatch = errors.New("thread command replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)
	refPattern         = regexp.MustCompile(`^ref_[A-Za-z0-9_-]{20,160}$`)
)

type Command struct {
	ID, ActorKey, RevealRef, RecipeRef string
	BandVersion                        uint32
	ExpectedRevision                   uint64
	At                                 time.Time
}
type Provenance struct {
	BandVersion          uint32
	RevealRef, RecipeRef string
	IssuedAt             time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Thread struct {
	id         string
	members    []string
	provenance *Provenance
	revision   uint64
	commands   []Applied
}
type State struct {
	ID         string
	Members    []string
	Provenance *Provenance
	Revision   uint64
	Commands   []Applied
}
type View struct {
	ID         string
	Provenance Provenance
	Revision   uint64
}

func New(id string, members []string) (Thread, error) {
	m := append([]string(nil), members...)
	slices.Sort(m)
	m = slices.Compact(m)
	if !idPattern.MatchString(id) || len(m) != 2 || !keyPattern.MatchString(m[0]) || !keyPattern.MatchString(m[1]) {
		return Thread{}, ErrInvalid
	}
	return Thread{id: id, members: m}, nil
}
func Rehydrate(s State) (Thread, error) {
	t, err := New(s.ID, s.Members)
	if err != nil {
		return Thread{}, err
	}
	t.provenance = s.Provenance
	t.revision = s.Revision
	t.commands = append([]Applied(nil), s.Commands...)
	if (t.provenance == nil && (t.revision != 0 || len(t.commands) != 0)) || (t.provenance != nil && (t.revision != 1 || len(t.commands) != 1 || !validProvenance(*t.provenance))) {
		return Thread{}, ErrInvalid
	}
	return t, nil
}
func (t Thread) Issue(c Command) (Thread, error) {
	if !t.member(c.ActorKey) || !idPattern.MatchString(c.ID) || c.At.IsZero() || c.BandVersion == 0 || !refPattern.MatchString(c.RevealRef) || !refPattern.MatchString(c.RecipeRef) {
		return Thread{}, ErrDenied
	}
	fp := fingerprint(t.id, c)
	for _, seen := range t.commands {
		if seen.ID == c.ID {
			if seen.Fingerprint != fp {
				return Thread{}, ErrCommandMismatch
			}
			return t, nil
		}
	}
	if c.ExpectedRevision != t.revision {
		return Thread{}, ErrStaleRevision
	}
	if t.provenance != nil {
		return Thread{}, ErrAlreadyIssued
	}
	t.revision = 1
	t.provenance = &Provenance{BandVersion: c.BandVersion, RevealRef: c.RevealRef, RecipeRef: c.RecipeRef, IssuedAt: c.At.UTC()}
	t.commands = append(t.commands, Applied{ID: c.ID, Fingerprint: fp, Revision: 1})
	return t, nil
}
func (t Thread) View(actor string) (View, error) {
	if !t.member(actor) || t.provenance == nil {
		return View{}, ErrDenied
	}
	return View{ID: t.id, Provenance: *t.Provenance(), Revision: t.revision}, nil
}
func validProvenance(p Provenance) bool {
	return p.BandVersion > 0 && refPattern.MatchString(p.RevealRef) && refPattern.MatchString(p.RecipeRef) && !p.IssuedAt.IsZero()
}
func fingerprint(id string, c Command) string {
	sum := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + c.RevealRef + "\x00" + c.RecipeRef + "\x00" + strconv.FormatUint(uint64(c.BandVersion), 10) + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(sum[:])
}
func (t Thread) member(k string) bool { return slices.Contains(t.members, k) }
func (t Thread) ID() string           { return t.id }
func (t Thread) Members() []string    { return append([]string(nil), t.members...) }
func (t Thread) Provenance() *Provenance {
	if t.provenance == nil {
		return nil
	}
	p := *t.provenance
	return &p
}
func (t Thread) Revision() uint64    { return t.revision }
func (t Thread) Commands() []Applied { return append([]Applied(nil), t.commands...) }
