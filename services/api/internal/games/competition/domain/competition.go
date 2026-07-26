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
)

const (
	MinCohort = 4
	MaxCohort = 16
)

var (
	ErrInvalid    = errors.New("invalid competition")
	ErrTransition = errors.New("invalid competition transition")
	ErrStale      = errors.New("stale competition")
	ErrMismatch   = errors.New("competition command mismatch")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var key = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Status string

const (
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
)

type ReviewStatus string

const (
	ReviewOpen     ReviewStatus = "open"
	ReviewResolved ReviewStatus = "resolved"
	ReviewAppealed ReviewStatus = "appealed"
	ReviewFinal    ReviewStatus = "final"
)

type Decision string

const (
	DecisionNone        Decision = "none"
	DecisionNoAction    Decision = "no_action"
	DecisionRulesAction Decision = "rules_action"
)

type Match struct {
	ID        string `bson:"id"`
	Round     int    `bson:"round"`
	Slot      int    `bson:"slot"`
	FirstKey  string `bson:"firstKey"`
	SecondKey string `bson:"secondKey"`
	WinnerKey string `bson:"winnerKey,omitempty"`
	ResultKey string `bson:"resultKey,omitempty"`
}
type LadderEntry struct {
	MemberKey string `bson:"memberKey"`
	Played    int    `bson:"played"`
	Wins      int    `bson:"wins"`
}
type Review struct {
	ID, MatchID, EvidenceKey, OpenedByKey string
	Status                                ReviewStatus
	Decision                              Decision
	OpenedAt, ResolvedAt                  time.Time
}
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
	ID       string
	Status   Status
	Entrants []string
	Matches  []Match
	Ladder   []LadderEntry
	Reviews  []Review
}
type Competition struct {
	id, cohortKey string
	entrants      []string
	matches       []Match
	ladder        []LadderEntry
	reviews       []Review
	status        Status
	revision      uint64
	events        []Event
	commands      []Applied
}
type State struct {
	ID, CohortKey string
	Entrants      []string
	Matches       []Match
	Ladder        []LadderEntry
	Reviews       []Review
	Status        Status
	Revision      uint64
	Events        []Event
	Commands      []Applied
}

func Create(id, cohort string, entrants []string, c Command) (Competition, error) {
	e := append([]string(nil), entrants...)
	slices.Sort(e)
	x := Competition{id: id, cohortKey: cohort, entrants: e, status: StatusActive}
	if !opaque.MatchString(id) || !key.MatchString(cohort) || !validEntrants(e) || c.ExpectedRevision != 0 {
		return Competition{}, ErrInvalid
	}
	for i := 0; i < len(e); i += 2 {
		x.matches = append(x.matches, Match{ID: matchID(1, i/2), Round: 1, Slot: i / 2, FirstKey: e[i], SecondKey: e[i+1]})
	}
	for _, m := range e {
		x.ladder = append(x.ladder, LadderEntry{MemberKey: m})
	}
	if err := x.apply(c, "create", ""); err != nil {
		return Competition{}, err
	}
	return x, nil
}
func Rehydrate(st State) (Competition, error) {
	x := Competition{id: st.ID, cohortKey: st.CohortKey, entrants: append([]string(nil), st.Entrants...), matches: append([]Match(nil), st.Matches...), ladder: append([]LadderEntry(nil), st.Ladder...), reviews: append([]Review(nil), st.Reviews...), status: st.Status, revision: st.Revision, events: append([]Event(nil), st.Events...), commands: append([]Applied(nil), st.Commands...)}
	if !opaque.MatchString(x.id) || !key.MatchString(x.cohortKey) || !validEntrants(x.entrants) || !slices.IsSorted(x.entrants) || len(x.events) != int(x.revision) || len(x.commands) != int(x.revision) || x.revision == 0 {
		return Competition{}, ErrInvalid
	}
	return x, nil
}
func (x Competition) RecordResult(matchID, winner, resultKey string, c Command) (Competition, error) {
	if replay, e := x.replay(c, "result", matchID, winner, resultKey); replay || e != nil {
		return x, e
	}
	x = x.clone()
	if x.status != StatusActive || !key.MatchString(resultKey) {
		return Competition{}, ErrTransition
	}
	idx := -1
	for i, m := range x.matches {
		if m.ID == matchID {
			idx = i
		}
	}
	if idx < 0 || x.matches[idx].WinnerKey != "" || (winner != x.matches[idx].FirstKey && winner != x.matches[idx].SecondKey) {
		return Competition{}, ErrTransition
	}
	m := x.matches[idx]
	m.WinnerKey, m.ResultKey = winner, resultKey
	x.matches[idx] = m
	for i := range x.ladder {
		if x.ladder[i].MemberKey == m.FirstKey || x.ladder[i].MemberKey == m.SecondKey {
			x.ladder[i].Played++
		}
		if x.ladder[i].MemberKey == winner {
			x.ladder[i].Wins++
		}
	}
	x.advanceRound(m.Round)
	return x, x.apply(c, "result", matchID, winner, resultKey)
}
func (x *Competition) advanceRound(round int) {
	var current []Match
	for _, m := range x.matches {
		if m.Round == round {
			current = append(current, m)
		}
	}
	for _, m := range current {
		if m.WinnerKey == "" {
			return
		}
	}
	if len(current) == 1 {
		x.status = StatusCompleted
		return
	}
	for i := 0; i < len(current); i += 2 {
		x.matches = append(x.matches, Match{ID: matchID(round+1, i/2), Round: round + 1, Slot: i / 2, FirstKey: current[i].WinnerKey, SecondKey: current[i+1].WinnerKey})
	}
}
func (x Competition) OpenReview(reviewID, matchID, evidence, actor string, now time.Time, c Command) (Competition, error) {
	if replay, e := x.replay(c, "review-open", reviewID, matchID, evidence); replay || e != nil {
		return x, e
	}
	x = x.clone()
	if !opaque.MatchString(reviewID) || !key.MatchString(evidence) || !slices.Contains(x.entrants, actor) || !hasMatch(x.matches, matchID) {
		return Competition{}, ErrInvalid
	}
	x.reviews = append(x.reviews, Review{ID: reviewID, MatchID: matchID, EvidenceKey: evidence, OpenedByKey: actor, Status: ReviewOpen, Decision: DecisionNone, OpenedAt: now.UTC()})
	return x, x.apply(c, "review-open", reviewID, matchID, evidence)
}
func (x Competition) ResolveReview(reviewID string, decision Decision, now time.Time, c Command) (Competition, error) {
	action := "review-resolve:" + string(decision)
	if replay, e := x.replay(c, action, reviewID); replay || e != nil {
		return x, e
	}
	x = x.clone()
	idx := reviewIndex(x.reviews, reviewID)
	if idx < 0 || x.reviews[idx].Status != ReviewOpen || (decision != DecisionNoAction && decision != DecisionRulesAction) {
		return Competition{}, ErrTransition
	}
	x.reviews[idx].Status, x.reviews[idx].Decision, x.reviews[idx].ResolvedAt = ReviewResolved, decision, now.UTC()
	return x, x.apply(c, action, reviewID)
}
func (x Competition) Appeal(reviewID, actor string, c Command) (Competition, error) {
	if replay, e := x.replay(c, "review-appeal", reviewID, actor); replay || e != nil {
		return x, e
	}
	x = x.clone()
	idx := reviewIndex(x.reviews, reviewID)
	if idx < 0 || x.reviews[idx].Status != ReviewResolved || !slices.Contains(x.entrants, actor) {
		return Competition{}, ErrTransition
	}
	x.reviews[idx].Status = ReviewAppealed
	return x, x.apply(c, "review-appeal", reviewID, actor)
}
func (x Competition) ResolveAppeal(reviewID string, decision Decision, now time.Time, c Command) (Competition, error) {
	action := "appeal-resolve:" + string(decision)
	if replay, e := x.replay(c, action, reviewID); replay || e != nil {
		return x, e
	}
	x = x.clone()
	idx := reviewIndex(x.reviews, reviewID)
	if idx < 0 || x.reviews[idx].Status != ReviewAppealed || (decision != DecisionNoAction && decision != DecisionRulesAction) {
		return Competition{}, ErrTransition
	}
	x.reviews[idx].Status, x.reviews[idx].Decision, x.reviews[idx].ResolvedAt = ReviewFinal, decision, now.UTC()
	return x, x.apply(c, action, reviewID)
}
func (x Competition) Project() Projection {
	ladder := append([]LadderEntry(nil), x.ladder...)
	slices.SortFunc(ladder, func(a, b LadderEntry) int {
		if a.Wins != b.Wins {
			return b.Wins - a.Wins
		}
		return strings.Compare(a.MemberKey, b.MemberKey)
	})
	return Projection{ID: x.id, Status: x.status, Entrants: append([]string(nil), x.entrants...), Matches: append([]Match(nil), x.matches...), Ladder: ladder, Reviews: append([]Review(nil), x.reviews...)}
}
func (x Competition) clone() Competition {
	x.entrants = append([]string(nil), x.entrants...)
	x.matches = append([]Match(nil), x.matches...)
	x.ladder = append([]LadderEntry(nil), x.ladder...)
	x.reviews = append([]Review(nil), x.reviews...)
	x.events = append([]Event(nil), x.events...)
	x.commands = append([]Applied(nil), x.commands...)
	return x
}
func (x *Competition) apply(c Command, a string, v ...string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != x.revision {
		return ErrStale
	}
	f := fingerprint(x.id, c, a, v...)
	x.revision++
	x.events = append(x.events, Event{x.revision, c.ID, a, c.At.UTC()})
	x.commands = append(x.commands, Applied{c.ID, f, x.revision})
	return nil
}
func (x Competition) replay(c Command, a string, v ...string) (bool, error) {
	f := fingerprint(x.id, c, a, v...)
	for _, q := range x.commands {
		if q.ID == c.ID {
			if q.Fingerprint != f {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func validEntrants(v []string) bool {
	if len(v) < MinCohort || len(v) > MaxCohort || len(v)&(len(v)-1) != 0 {
		return false
	}
	for i, x := range v {
		if !key.MatchString(x) || (i > 0 && v[i-1] == x) {
			return false
		}
	}
	return true
}
func hasMatch(v []Match, id string) bool {
	for _, m := range v {
		if m.ID == id {
			return true
		}
	}
	return false
}
func reviewIndex(v []Review, id string) int {
	for i, r := range v {
		if r.ID == id {
			return i
		}
	}
	return -1
}
func matchID(r, s int) string { return "round-" + strconv.Itoa(r) + ":match-" + strconv.Itoa(s) }
func fingerprint(id string, c Command, a string, v ...string) string {
	h := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + a + "\x00" + strings.Join(v, "\x00") + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h[:])
}
func (x Competition) ID() string            { return x.id }
func (x Competition) CohortKey() string     { return x.cohortKey }
func (x Competition) Entrants() []string    { return append([]string(nil), x.entrants...) }
func (x Competition) Matches() []Match      { return append([]Match(nil), x.matches...) }
func (x Competition) Ladder() []LadderEntry { return append([]LadderEntry(nil), x.ladder...) }
func (x Competition) Reviews() []Review     { return append([]Review(nil), x.reviews...) }
func (x Competition) Status() Status        { return x.status }
func (x Competition) Revision() uint64      { return x.revision }
func (x Competition) Events() []Event       { return append([]Event(nil), x.events...) }
func (x Competition) Commands() []Applied   { return append([]Applied(nil), x.commands...) }
