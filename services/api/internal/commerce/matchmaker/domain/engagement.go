package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid matchmaker engagement")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

const MaxMilestones = 8

type License struct {
	ID, MatchmakerKey, Jurisdiction      string
	Version                              uint64
	ValidFrom, ValidUntil                time.Time
	MinimumFeePesewas, MaximumFeePesewas uint64
}

func (license License) Current(at time.Time) bool {
	return token.MatchString(license.ID) && opaque.MatchString(license.MatchmakerKey) &&
		token.MatchString(license.Jurisdiction) && license.Version > 0 &&
		!license.ValidFrom.IsZero() && license.ValidUntil.After(license.ValidFrom) &&
		(at.Equal(license.ValidFrom) || at.After(license.ValidFrom)) && at.Before(license.ValidUntil) &&
		license.MinimumFeePesewas > 0 && license.MaximumFeePesewas >= license.MinimumFeePesewas
}

type Milestone struct {
	ID         string
	FeePesewas uint64
	DueAfter   time.Duration
}
type Terms struct {
	ID              string
	Version         uint64
	TotalFeePesewas uint64
	Milestones      []Milestone
}

type EventKind string

const (
	EventBooked             EventKind = "booked"
	EventProposalCurated    EventKind = "proposal_curated"
	EventMemberConsented    EventKind = "member_consented"
	EventCandidateConsented EventKind = "candidate_consented"
	EventProposalExposed    EventKind = "proposal_exposed"
	EventCompleted          EventKind = "completed"
)

type Event struct {
	Sequence  uint64
	Kind      EventKind
	CommandID string
	At        time.Time
}

type State struct {
	ID, MemberKey, MatchmakerKey                            string
	LicenseID                                               string
	LicenseVersion                                          uint64
	Terms                                                   Terms
	BookedAt                                                time.Time
	SealedProposalRef                                       string
	MemberConsented, CandidateConsented, Exposed, Completed bool
	Revision                                                uint64
	Events                                                  []Event
	AppliedIDs                                              []string
}
type Engagement struct{ state State }

func Book(id, memberKey string, license License, terms Terms, commandID string, at time.Time) (Engagement, error) {
	terms.Milestones = append([]Milestone(nil), terms.Milestones...)
	if !opaque.MatchString(id) || !opaque.MatchString(memberKey) || !license.Current(at) || !validTerms(terms, license) || !token.MatchString(commandID) {
		return Engagement{}, ErrInvalid
	}
	s := State{ID: id, MemberKey: memberKey, MatchmakerKey: license.MatchmakerKey, LicenseID: license.ID, LicenseVersion: license.Version, Terms: terms, BookedAt: at.UTC(), Revision: 1, AppliedIDs: []string{commandID}}
	s.Events = []Event{{Sequence: 1, Kind: EventBooked, CommandID: commandID, At: at.UTC()}}
	return Engagement{s}, nil
}
func Rehydrate(s State) (Engagement, error) {
	s.Terms.Milestones = append([]Milestone(nil), s.Terms.Milestones...)
	s.Events = append([]Event(nil), s.Events...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	if !validState(s) {
		return Engagement{}, ErrInvalid
	}
	return Engagement{s}, nil
}
func (e Engagement) Curate(commandID, sealedRef string, at time.Time) (Engagement, error) {
	if e.state.SealedProposalRef != "" || !opaque.MatchString(sealedRef) {
		return Engagement{}, ErrInvalid
	}
	n, x := e.transition(EventProposalCurated, commandID, at)
	if x == nil {
		n.state.SealedProposalRef = sealedRef
	}
	return n, x
}

type ConsentRole string

const (
	ConsentMember    ConsentRole = "member"
	ConsentCandidate ConsentRole = "candidate"
)

func (e Engagement) Consent(role ConsentRole, commandID string, at time.Time) (Engagement, error) {
	if e.state.SealedProposalRef == "" || e.state.Exposed {
		return Engagement{}, ErrInvalid
	}
	kind := EventMemberConsented
	if role == ConsentMember {
		if e.state.MemberConsented {
			return Engagement{}, ErrInvalid
		}
	} else if role == ConsentCandidate {
		kind = EventCandidateConsented
		if e.state.CandidateConsented {
			return Engagement{}, ErrInvalid
		}
	} else {
		return Engagement{}, ErrInvalid
	}
	n, x := e.transition(kind, commandID, at)
	if x == nil {
		if role == ConsentMember {
			n.state.MemberConsented = true
		} else {
			n.state.CandidateConsented = true
		}
	}
	return n, x
}
func (e Engagement) Expose(commandID string, at time.Time) (Engagement, error) {
	if !e.state.MemberConsented || !e.state.CandidateConsented || e.state.Exposed {
		return Engagement{}, ErrInvalid
	}
	n, x := e.transition(EventProposalExposed, commandID, at)
	if x == nil {
		n.state.Exposed = true
	}
	return n, x
}
func (e Engagement) Complete(commandID string, at time.Time) (Engagement, error) {
	if !e.state.Exposed || e.state.Completed {
		return Engagement{}, ErrInvalid
	}
	n, x := e.transition(EventCompleted, commandID, at)
	if x == nil {
		n.state.Completed = true
	}
	return n, x
}
func (e Engagement) ProposalRef() (string, bool) {
	if !e.state.Exposed {
		return "", false
	}
	return e.state.SealedProposalRef, true
}
func (e Engagement) ReviewEligible() bool { return e.state.Completed }
func (e Engagement) transition(kind EventKind, id string, at time.Time) (Engagement, error) {
	if !token.MatchString(id) || at.IsZero() || slices.Contains(e.state.AppliedIDs, id) {
		return Engagement{}, ErrInvalid
	}
	n := e.State()
	n.Revision++
	n.AppliedIDs = append(n.AppliedIDs, id)
	n.Events = append(n.Events, Event{Sequence: n.Revision, Kind: kind, CommandID: id, At: at.UTC()})
	return Engagement{n}, nil
}
func (e Engagement) State() State {
	s := e.state
	s.Terms.Milestones = append([]Milestone(nil), s.Terms.Milestones...)
	s.Events = append([]Event(nil), s.Events...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	return s
}
func (e Engagement) ID() string       { return e.state.ID }
func (e Engagement) Revision() uint64 { return e.state.Revision }
func validTerms(t Terms, l License) bool {
	if !token.MatchString(t.ID) || t.Version == 0 || t.TotalFeePesewas < l.MinimumFeePesewas || t.TotalFeePesewas > l.MaximumFeePesewas || len(t.Milestones) == 0 || len(t.Milestones) > MaxMilestones {
		return false
	}
	var sum uint64
	seen := map[string]bool{}
	var due time.Duration
	for _, m := range t.Milestones {
		if !token.MatchString(m.ID) || m.FeePesewas == 0 || m.DueAfter < due || seen[m.ID] {
			return false
		}
		seen[m.ID] = true
		due = m.DueAfter
		if ^uint64(0)-sum < m.FeePesewas {
			return false
		}
		sum += m.FeePesewas
	}
	return sum == t.TotalFeePesewas
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.MemberKey) || !opaque.MatchString(s.MatchmakerKey) || !token.MatchString(s.LicenseID) || s.LicenseVersion == 0 || s.Revision == 0 || len(s.Events) != int(s.Revision) || len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	for i, x := range s.Events {
		if x.Sequence != uint64(i+1) || x.CommandID != s.AppliedIDs[i] || !token.MatchString(x.CommandID) || x.At.IsZero() {
			return false
		}
	}
	return true
}
