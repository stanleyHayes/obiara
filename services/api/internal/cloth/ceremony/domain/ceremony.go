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

const (
	AnnouncementKind = "gate_completed"
)

type Action string

const (
	ActionOpened                Action = "opened"
	ActionConfirmed             Action = "ceremony_confirmed"
	ActionAnnouncementProposed  Action = "announcement_proposed"
	ActionAnnouncementConsented Action = "announcement_consented"
	ActionAnnouncementPublished Action = "announcement_published"
)

var (
	ErrInvalid              = errors.New("invalid gate ceremony")
	ErrNotMember            = errors.New("ceremony unavailable")
	ErrNotComplete          = errors.New("ceremony is not complete")
	ErrAnnouncementExists   = errors.New("announcement already proposed")
	ErrAnnouncementNotReady = errors.New("announcement lacks fresh dual consent")
	ErrAlreadyPublished     = errors.New("announcement already published")
	ErrStaleRevision        = errors.New("stale ceremony revision")
	ErrCommandMismatch      = errors.New("ceremony command replay mismatch")
)
var keyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)

type Announcement struct {
	DestinationKey, Kind string
	Consents             []string
	Published            bool
}
type Command struct {
	ID, ActorKey, Fingerprint string
	ExpectedRevision          uint64
	At                        time.Time
}
type Event struct {
	Sequence                                         uint64
	Action                                           Action
	CommandID, ActorKey, Fingerprint, DestinationKey string
	At                                               time.Time
}
type State struct {
	ID            string
	Members       [2]string
	Confirmations []string
	Announcement  *Announcement
	Revision      uint64
	Events        []Event
}
type Ceremony struct{ state State }

func Open(id string, members [2]string, c Command) (Ceremony, error) {
	if !idPattern.MatchString(id) || !validMembers(members) || c.ExpectedRevision != 0 || (c.ActorKey != members[0] && c.ActorKey != members[1]) {
		return Ceremony{}, ErrInvalid
	}
	if members[1] < members[0] {
		members[0], members[1] = members[1], members[0]
	}
	x := Ceremony{State{ID: id, Members: members}}
	return x.apply(ActionOpened, "", c)
}
func Rehydrate(s State) (Ceremony, error) {
	if !idPattern.MatchString(s.ID) || !validMembers(s.Members) || s.Members[0] >= s.Members[1] || s.Revision == 0 || len(s.Events) != int(s.Revision) {
		return Ceremony{}, ErrInvalid
	}
	x := Ceremony{state: s}
	x.state.Confirmations = nil
	x.state.Announcement = nil
	x.state.Events = append([]Event(nil), s.Events...)
	for i, event := range x.state.Events {
		if event.Sequence != uint64(i+1) || !idPattern.MatchString(event.CommandID) || !keyPattern.MatchString(event.ActorKey) || !keyPattern.MatchString(event.Fingerprint) || (event.ActorKey != s.Members[0] && event.ActorKey != s.Members[1]) {
			return Ceremony{}, ErrInvalid
		}
		switch event.Action {
		case ActionOpened:
			if i != 0 {
				return Ceremony{}, ErrInvalid
			}
		case ActionConfirmed:
			x.state.Confirmations = add(x.state.Confirmations, event.ActorKey)
		case ActionAnnouncementProposed:
			if len(x.state.Confirmations) != 2 || x.state.Announcement != nil || !keyPattern.MatchString(event.DestinationKey) {
				return Ceremony{}, ErrInvalid
			}
			x.state.Announcement = &Announcement{DestinationKey: event.DestinationKey, Kind: AnnouncementKind}
		case ActionAnnouncementConsented:
			if x.state.Announcement == nil || x.state.Announcement.Published {
				return Ceremony{}, ErrInvalid
			}
			x.state.Announcement.Consents = add(x.state.Announcement.Consents, event.ActorKey)
		case ActionAnnouncementPublished:
			if x.state.Announcement == nil || len(x.state.Announcement.Consents) != 2 || x.state.Announcement.Published {
				return Ceremony{}, ErrInvalid
			}
			x.state.Announcement.Published = true
		default:
			return Ceremony{}, ErrInvalid
		}
	}
	if !slices.Equal(x.state.Confirmations, sortedCopy(s.Confirmations)) || !announcementEqual(x.state.Announcement, s.Announcement) {
		return Ceremony{}, ErrInvalid
	}
	return x, nil
}
func (x Ceremony) Confirm(c Command) (Ceremony, error) { return x.change(ActionConfirmed, "", c) }
func (x Ceremony) ProposeAnnouncement(destination string, c Command) (Ceremony, error) {
	if !x.Complete() {
		return x, ErrNotComplete
	}
	if x.state.Announcement != nil {
		return x, ErrAnnouncementExists
	}
	if !keyPattern.MatchString(destination) {
		return x, ErrInvalid
	}
	return x.change(ActionAnnouncementProposed, destination, c)
}
func (x Ceremony) ConsentAnnouncement(c Command) (Ceremony, error) {
	if x.state.Announcement == nil || x.state.Announcement.Published {
		return x, ErrAnnouncementNotReady
	}
	return x.change(ActionAnnouncementConsented, x.state.Announcement.DestinationKey, c)
}
func (x Ceremony) PublishAnnouncement(c Command) (Ceremony, error) {
	if x.state.Announcement == nil || len(x.state.Announcement.Consents) != 2 {
		return x, ErrAnnouncementNotReady
	}
	if x.state.Announcement.Published {
		return x, ErrAlreadyPublished
	}
	return x.change(ActionAnnouncementPublished, x.state.Announcement.DestinationKey, c)
}
func (x Ceremony) change(action Action, destination string, c Command) (Ceremony, error) {
	for _, event := range x.state.Events {
		if event.CommandID == c.ID {
			if event.Fingerprint != c.Fingerprint {
				return x, ErrCommandMismatch
			}
			return x, nil
		}
	}
	if c.ActorKey != x.state.Members[0] && c.ActorKey != x.state.Members[1] {
		return x, ErrNotMember
	}
	if action == ActionConfirmed && slices.Contains(x.state.Confirmations, c.ActorKey) {
		return x.apply(action, destination, c)
	}
	return x.apply(action, destination, c)
}
func (x Ceremony) apply(action Action, destination string, c Command) (Ceremony, error) {
	if !idPattern.MatchString(c.ID) || !keyPattern.MatchString(c.ActorKey) || !keyPattern.MatchString(c.Fingerprint) || c.At.IsZero() {
		return Ceremony{}, ErrInvalid
	}
	if c.ExpectedRevision != x.state.Revision {
		return Ceremony{}, ErrStaleRevision
	}
	x.state.Confirmations = append([]string(nil), x.state.Confirmations...)
	if x.state.Announcement != nil {
		copy := *x.state.Announcement
		copy.Consents = append([]string(nil), copy.Consents...)
		x.state.Announcement = &copy
	}
	switch action {
	case ActionConfirmed:
		x.state.Confirmations = add(x.state.Confirmations, c.ActorKey)
	case ActionAnnouncementProposed:
		x.state.Announcement = &Announcement{DestinationKey: destination, Kind: AnnouncementKind}
	case ActionAnnouncementConsented:
		x.state.Announcement.Consents = add(x.state.Announcement.Consents, c.ActorKey)
	case ActionAnnouncementPublished:
		x.state.Announcement.Published = true
	}
	x.state.Revision++
	x.state.Events = append(append([]Event(nil), x.state.Events...), Event{x.state.Revision, action, c.ID, c.ActorKey, c.Fingerprint, destination, c.At.UTC()})
	return x, nil
}
func Fingerprint(id string, action Action, actor, destination string, revision uint64) string {
	sum := sha256.Sum256([]byte(id + "\x00" + string(action) + "\x00" + actor + "\x00" + destination + "\x00" + strconv.FormatUint(revision, 10)))
	return hex.EncodeToString(sum[:])
}
func validMembers(m [2]string) bool {
	return keyPattern.MatchString(m[0]) && keyPattern.MatchString(m[1]) && m[0] != m[1]
}
func add(v []string, x string) []string {
	if !slices.Contains(v, x) {
		v = append(v, x)
		slices.Sort(v)
	}
	return v
}
func sortedCopy(v []string) []string { x := append([]string(nil), v...); slices.Sort(x); return x }
func announcementEqual(a, b *Announcement) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.DestinationKey == b.DestinationKey && a.Kind == b.Kind && a.Published == b.Published && slices.Equal(a.Consents, sortedCopy(b.Consents))
}
func (x Ceremony) Complete() bool { return len(x.state.Confirmations) == 2 }
func (x Ceremony) AnnouncementReady() bool {
	return x.state.Announcement != nil && len(x.state.Announcement.Consents) == 2 && !x.state.Announcement.Published
}
func (x Ceremony) State() State {
	s := x.state
	s.Confirmations = append([]string(nil), s.Confirmations...)
	s.Events = append([]Event(nil), s.Events...)
	if s.Announcement != nil {
		copy := *s.Announcement
		copy.Consents = append([]string(nil), copy.Consents...)
		s.Announcement = &copy
	}
	return s
}
func (x Ceremony) ID() string       { return x.state.ID }
func (x Ceremony) Revision() uint64 { return x.state.Revision }
