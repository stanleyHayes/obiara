package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid escrow")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

const MaxMilestones = 16

type MilestoneTerm struct {
	ID           string
	GrossPesewas uint64
	FeePesewas   uint64
}
type Terms struct {
	ID         string
	Version    uint64
	Milestones []MilestoneTerm
}
type EvidenceRole string

const (
	DeliveryEvidence   EvidenceRole = "delivery"
	AcceptanceEvidence EvidenceRole = "acceptance"
)

type Evidence struct {
	Role EvidenceRole
	Ref  string
	At   time.Time
}
type MilestoneState struct {
	Term         MilestoneTerm
	Evidence     []Evidence
	Settled      bool
	StatementRef string
}
type Dispute struct {
	Ref, EscalationRef string
	RaisedAt           time.Time
}
type Event struct {
	Sequence                                uint64
	Kind, CommandID, MilestoneID, Reference string
	At                                      time.Time
}
type State struct {
	ID, OwnerKey, EngagementID, FundingRef string
	FundedPesewas                          uint64
	TermsID                                string
	TermsVersion                           uint64
	Milestones                             []MilestoneState
	Dispute                                *Dispute
	SettledPesewas                         uint64
	Revision                               uint64
	Events                                 []Event
	AppliedIDs                             []string
}
type Escrow struct{ state State }

func Fund(id, ownerKey, engagementID, fundingRef string, amount uint64, terms Terms, command string, at time.Time) (Escrow, error) {
	terms.Milestones = append([]MilestoneTerm(nil), terms.Milestones...)
	if !opaque.MatchString(id) || !opaque.MatchString(ownerKey) || !opaque.MatchString(engagementID) || !opaque.MatchString(fundingRef) || amount == 0 || !validTerms(terms, amount) || !token.MatchString(command) || at.IsZero() {
		return Escrow{}, ErrInvalid
	}
	m := make([]MilestoneState, len(terms.Milestones))
	for i, x := range terms.Milestones {
		m[i] = MilestoneState{Term: x}
	}
	s := State{ID: id, OwnerKey: ownerKey, EngagementID: engagementID, FundingRef: fundingRef, FundedPesewas: amount, TermsID: terms.ID, TermsVersion: terms.Version, Milestones: m, Revision: 1, AppliedIDs: []string{command}}
	s.Events = []Event{{Sequence: 1, Kind: "funded", CommandID: command, Reference: fundingRef, At: at.UTC()}}
	return Escrow{s}, nil
}
func Rehydrate(s State) (Escrow, error) {
	s = clone(s)
	if !validState(s) {
		return Escrow{}, ErrInvalid
	}
	return Escrow{s}, nil
}
func (e Escrow) AddEvidence(milestone string, role EvidenceRole, ref, command string, at time.Time) (Escrow, error) {
	if e.state.Dispute != nil || !opaque.MatchString(ref) || (role != DeliveryEvidence && role != AcceptanceEvidence) {
		return Escrow{}, ErrInvalid
	}
	index := e.index(milestone)
	if index < 0 || e.state.Milestones[index].Settled {
		return Escrow{}, ErrInvalid
	}
	for _, x := range e.state.Milestones[index].Evidence {
		if x.Role == role {
			return Escrow{}, ErrInvalid
		}
	}
	n, x := e.transition("evidence", command, milestone, ref, at)
	if x == nil {
		n.state.Milestones[index].Evidence = append(n.state.Milestones[index].Evidence, Evidence{role, ref, at.UTC()})
	}
	return n, x
}

type Statement struct {
	Ref, EscrowID, MilestoneID           string
	TermsID                              string
	TermsVersion                         uint64
	GrossPesewas, FeePesewas, NetPesewas uint64
	SettledAt                            time.Time
}

func (e Escrow) Settle(milestone, statementRef, command string, at time.Time) (Escrow, Statement, error) {
	if e.state.Dispute != nil || !opaque.MatchString(statementRef) {
		return Escrow{}, Statement{}, ErrInvalid
	}
	index := e.index(milestone)
	if index < 0 || e.state.Milestones[index].Settled || !completeEvidence(e.state.Milestones[index]) {
		return Escrow{}, Statement{}, ErrInvalid
	}
	term := e.state.Milestones[index].Term
	if e.state.SettledPesewas > e.state.FundedPesewas-term.GrossPesewas {
		return Escrow{}, Statement{}, ErrInvalid
	}
	n, x := e.transition("settled", command, milestone, statementRef, at)
	if x != nil {
		return Escrow{}, Statement{}, x
	}
	n.state.Milestones[index].Settled = true
	n.state.Milestones[index].StatementRef = statementRef
	n.state.SettledPesewas += term.GrossPesewas
	statement := Statement{statementRef, e.state.ID, milestone, e.state.TermsID, e.state.TermsVersion, term.GrossPesewas, term.FeePesewas, term.GrossPesewas - term.FeePesewas, at.UTC()}
	return n, statement, nil
}
func (e Escrow) RaiseDispute(ref, escalationRef, command string, at time.Time) (Escrow, error) {
	if e.state.Dispute != nil || !opaque.MatchString(ref) || !opaque.MatchString(escalationRef) {
		return Escrow{}, ErrInvalid
	}
	n, x := e.transition("disputed", command, "", ref, at)
	if x == nil {
		n.state.Dispute = &Dispute{ref, escalationRef, at.UTC()}
	}
	return n, x
}
func (e Escrow) transition(kind, command, milestone, ref string, at time.Time) (Escrow, error) {
	if !token.MatchString(command) || at.IsZero() || slices.Contains(e.state.AppliedIDs, command) {
		return Escrow{}, ErrInvalid
	}
	n := e.State()
	n.Revision++
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Events = append(n.Events, Event{n.Revision, kind, command, milestone, ref, at.UTC()})
	return Escrow{n}, nil
}
func (e Escrow) index(id string) int {
	for i, x := range e.state.Milestones {
		if x.Term.ID == id {
			return i
		}
	}
	return -1
}
func (e Escrow) State() State     { return clone(e.state) }
func (e Escrow) ID() string       { return e.state.ID }
func (e Escrow) Revision() uint64 { return e.state.Revision }
func validTerms(t Terms, amount uint64) bool {
	if !token.MatchString(t.ID) || t.Version == 0 || len(t.Milestones) == 0 || len(t.Milestones) > MaxMilestones {
		return false
	}
	seen := map[string]bool{}
	var sum uint64
	for _, m := range t.Milestones {
		if !token.MatchString(m.ID) || m.GrossPesewas == 0 || m.FeePesewas > m.GrossPesewas || seen[m.ID] || ^uint64(0)-sum < m.GrossPesewas {
			return false
		}
		seen[m.ID] = true
		sum += m.GrossPesewas
	}
	return sum == amount
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.OwnerKey) || !opaque.MatchString(s.EngagementID) || !opaque.MatchString(s.FundingRef) || s.FundedPesewas == 0 || !token.MatchString(s.TermsID) || s.TermsVersion == 0 || s.SettledPesewas > s.FundedPesewas || s.Revision == 0 || len(s.Events) != int(s.Revision) || len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	for i, x := range s.Events {
		if x.Sequence != uint64(i+1) || x.CommandID != s.AppliedIDs[i] || !token.MatchString(x.CommandID) || x.At.IsZero() {
			return false
		}
	}
	return true
}
func completeEvidence(m MilestoneState) bool {
	delivery, acceptance := false, false
	for _, x := range m.Evidence {
		delivery = delivery || x.Role == DeliveryEvidence
		acceptance = acceptance || x.Role == AcceptanceEvidence
	}
	return delivery && acceptance
}
func clone(s State) State {
	s.Milestones = append([]MilestoneState(nil), s.Milestones...)
	for i := range s.Milestones {
		s.Milestones[i].Evidence = append([]Evidence(nil), s.Milestones[i].Evidence...)
	}
	s.Events = append([]Event(nil), s.Events...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	if s.Dispute != nil {
		x := *s.Dispute
		s.Dispute = &x
	}
	return s
}
