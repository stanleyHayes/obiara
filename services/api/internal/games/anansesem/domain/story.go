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
	"unicode/utf8"
)

const (
	MaxPassages     = 40
	MaxPassageRunes = 1200
	MaxRevisions    = 20
)

var (
	ErrInvalid    = errors.New("invalid anansesem story")
	ErrNotTurn    = errors.New("not author's turn")
	ErrTransition = errors.New("invalid story transition")
	ErrStale      = errors.New("stale story revision")
	ErrMismatch   = errors.New("story command mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)
var title = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{1,47}$`)

type Revision struct {
	Number    uint64    `bson:"number"`
	Content   string    `bson:"content"`
	EditedAt  time.Time `bson:"editedAt"`
	CommandID string    `bson:"commandId"`
}
type Passage struct {
	ID        string     `bson:"id"`
	Ordinal   int        `bson:"ordinal"`
	AuthorKey string     `bson:"authorKey"`
	CreatedAt time.Time  `bson:"createdAt"`
	Revisions []Revision `bson:"revisions"`
}
type EditionPassage struct {
	Ordinal int    `bson:"ordinal" json:"ordinal"`
	Content string `bson:"content" json:"content"`
}
type Edition struct {
	Version     uint64           `bson:"version" json:"version"`
	TitleCode   string           `bson:"titleCode" json:"titleCode"`
	Passages    []EditionPassage `bson:"passages" json:"passages"`
	PublishedAt time.Time        `bson:"publishedAt" json:"publishedAt"`
}
type Grant struct{ AuthorKey, DraftFingerprint string }
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
type Story struct {
	id, roomKey, titleCode string
	authors                []string
	passages               []Passage
	grants                 []Grant
	editions               []Edition
	revision               uint64
	events                 []Event
	commands               []Applied
}
type State struct {
	ID, RoomKey, TitleCode string
	Authors                []string
	Passages               []Passage
	Grants                 []Grant
	Editions               []Edition
	Revision               uint64
	Events                 []Event
	Commands               []Applied
}

func Create(id, room, titleCode string, authors []string, c Command) (Story, error) {
	a := append([]string(nil), authors...)
	slices.Sort(a)
	s := Story{id: id, roomKey: room, titleCode: titleCode, authors: a}
	if !opaque.MatchString(id) || !key.MatchString(room) || !title.MatchString(titleCode) || !validAuthors(a) || c.ExpectedRevision != 0 {
		return Story{}, ErrInvalid
	}
	if e := s.apply(c, "create", ""); e != nil {
		return Story{}, e
	}
	return s, nil
}
func Rehydrate(st State) (Story, error) {
	s := Story{id: st.ID, roomKey: st.RoomKey, titleCode: st.TitleCode, authors: append([]string(nil), st.Authors...), passages: clonePassages(st.Passages), grants: append([]Grant(nil), st.Grants...), editions: cloneEditions(st.Editions), revision: st.Revision, events: append([]Event(nil), st.Events...), commands: append([]Applied(nil), st.Commands...)}
	if !opaque.MatchString(s.id) || !key.MatchString(s.roomKey) || !title.MatchString(s.titleCode) || !validAuthors(s.authors) || !slices.IsSorted(s.authors) || len(s.passages) > MaxPassages || len(s.events) != int(s.revision) || len(s.commands) != int(s.revision) || s.revision == 0 {
		return Story{}, ErrInvalid
	}
	for i, p := range s.passages {
		if p.Ordinal != i || !opaque.MatchString(p.ID) || p.AuthorKey != s.authors[i%2] || len(p.Revisions) == 0 || len(p.Revisions) > MaxRevisions {
			return Story{}, ErrInvalid
		}
		for _, r := range p.Revisions {
			if !validContent(r.Content) || r.EditedAt.IsZero() {
				return Story{}, ErrInvalid
			}
		}
	}
	return s, nil
}
func (s Story) Add(passageID, actor, content string, now time.Time, c Command) (Story, error) {
	if replay, e := s.replay(c, "add", passageID, content); replay || e != nil {
		return s, e
	}
	if len(s.passages) >= MaxPassages || actor != s.authors[len(s.passages)%2] {
		return Story{}, ErrNotTurn
	}
	if !opaque.MatchString(passageID) || !validContent(content) {
		return Story{}, ErrInvalid
	}
	s.passages = append(s.passages, Passage{ID: passageID, Ordinal: len(s.passages), AuthorKey: actor, CreatedAt: now.UTC(), Revisions: []Revision{{Number: 1, Content: content, EditedAt: now.UTC(), CommandID: c.ID}}})
	s.grants = nil
	return s, s.apply(c, "add", passageID, content)
}
func (s Story) Edit(passageID, actor, content string, now time.Time, c Command) (Story, error) {
	if replay, e := s.replay(c, "edit", passageID, content); replay || e != nil {
		return s, e
	}
	if !validContent(content) {
		return Story{}, ErrInvalid
	}
	index := -1
	for i, p := range s.passages {
		if p.ID == passageID {
			index = i
		}
	}
	if index < 0 || s.passages[index].AuthorKey != actor || len(s.passages[index].Revisions) >= MaxRevisions {
		return Story{}, ErrTransition
	}
	p := s.passages[index]
	p.Revisions = append(p.Revisions, Revision{Number: uint64(len(p.Revisions) + 1), Content: content, EditedAt: now.UTC(), CommandID: c.ID})
	s.passages[index] = p
	s.grants = nil
	return s, s.apply(c, "edit", passageID, content)
}
func (s Story) Grant(actor string, c Command) (Story, error) {
	draft := s.DraftFingerprint()
	if replay, e := s.replay(c, "grant", actor, draft); replay || e != nil {
		return s, e
	}
	if len(s.passages) == 0 || !slices.Contains(s.authors, actor) {
		return Story{}, ErrTransition
	}
	for _, g := range s.grants {
		if g.AuthorKey == actor {
			return Story{}, ErrTransition
		}
	}
	s.grants = append(s.grants, Grant{actor, draft})
	slices.SortFunc(s.grants, func(a, b Grant) int { return strings.Compare(a.AuthorKey, b.AuthorKey) })
	return s, s.apply(c, "grant", actor, draft)
}
func (s Story) Publish(now time.Time, c Command) (Story, Edition, error) {
	draft := s.DraftFingerprint()
	if replay, e := s.replay(c, "publish", draft); replay || e != nil {
		if e == nil && len(s.editions) > 0 {
			return s, s.editions[len(s.editions)-1], nil
		}
		return s, Edition{}, e
	}
	if len(s.grants) != 2 || s.grants[0].DraftFingerprint != draft || s.grants[1].DraftFingerprint != draft {
		return Story{}, Edition{}, ErrTransition
	}
	edition := Edition{Version: uint64(len(s.editions) + 1), TitleCode: s.titleCode, PublishedAt: now.UTC()}
	for _, p := range s.passages {
		edition.Passages = append(edition.Passages, EditionPassage{Ordinal: p.Ordinal, Content: p.Revisions[len(p.Revisions)-1].Content})
	}
	s.editions = append(s.editions, edition)
	s.grants = nil
	if e := s.apply(c, "publish", draft); e != nil {
		return Story{}, Edition{}, e
	}
	return s, edition, nil
}
func (s Story) DraftFingerprint() string {
	x := sha256.New()
	_, _ = x.Write([]byte(s.titleCode))
	for _, p := range s.passages {
		r := p.Revisions[len(p.Revisions)-1]
		_, _ = x.Write([]byte("\x00" + strconv.Itoa(p.Ordinal) + "\x00" + r.Content + "\x00" + strconv.FormatUint(r.Number, 10)))
	}
	return hex.EncodeToString(x.Sum(nil))
}
func (s *Story) apply(c Command, a string, v ...string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != s.revision {
		return ErrStale
	}
	f := fingerprint(s.id, c, a, v...)
	s.revision++
	s.events = append(s.events, Event{s.revision, c.ID, a, c.At.UTC()})
	s.commands = append(s.commands, Applied{c.ID, f, s.revision})
	return nil
}
func (s Story) replay(c Command, a string, v ...string) (bool, error) {
	f := fingerprint(s.id, c, a, v...)
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
func validContent(v string) bool {
	return strings.TrimSpace(v) == v && utf8.RuneCountInString(v) > 0 && utf8.RuneCountInString(v) <= MaxPassageRunes && !strings.ContainsAny(v, "\x00\r")
}
func validAuthors(a []string) bool {
	return len(a) == 2 && a[0] != a[1] && key.MatchString(a[0]) && key.MatchString(a[1])
}
func fingerprint(id string, c Command, a string, v ...string) string {
	x := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strings.Join(v, "\x00") + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(x[:])
}
func clonePassages(v []Passage) []Passage {
	o := append([]Passage(nil), v...)
	for i := range o {
		o[i].Revisions = append([]Revision(nil), o[i].Revisions...)
	}
	return o
}
func cloneEditions(v []Edition) []Edition {
	o := append([]Edition(nil), v...)
	for i := range o {
		o[i].Passages = append([]EditionPassage(nil), o[i].Passages...)
	}
	return o
}
func (s Story) ID() string          { return s.id }
func (s Story) RoomKey() string     { return s.roomKey }
func (s Story) TitleCode() string   { return s.titleCode }
func (s Story) Authors() []string   { return append([]string(nil), s.authors...) }
func (s Story) Passages() []Passage { return clonePassages(s.passages) }
func (s Story) Grants() []Grant     { return append([]Grant(nil), s.grants...) }
func (s Story) Editions() []Edition { return cloneEditions(s.editions) }
func (s Story) Revision() uint64    { return s.revision }
func (s Story) Events() []Event     { return append([]Event(nil), s.events...) }
func (s Story) Commands() []Applied { return append([]Applied(nil), s.commands...) }
