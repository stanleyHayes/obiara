package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	MinCohort       = 200
	MinEvidence     = 1000
	MinSliceCohort  = 50
	MaxSlices       = 8
	MinQuality      = 0.65
	MaxQuality      = 0.99
	MaxErrorRate    = 0.35
	MaxDisparity    = 0.10
	MaxApprovalLife = 30 * 24 * time.Hour
)

var (
	ErrInvalid    = errors.New("invalid offline evaluation")
	ErrStale      = errors.New("stale offline evaluation command")
	ErrMismatch   = errors.New("offline evaluation command mismatch")
	ErrTransition = errors.New("invalid offline evaluation transition")
)
var opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var slug = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

type Command struct {
	ID               string    `bson:"id"`
	ExpectedRevision uint64    `bson:"expectedRevision"`
	At               time.Time `bson:"at"`
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
type Snapshot struct {
	ID             string    `bson:"id"`
	Version        uint64    `bson:"version"`
	ConsentVersion uint64    `bson:"consentVersion"`
	EvaluatedAt    time.Time `bson:"evaluatedAt"`
}
type SliceMetric struct {
	PolicyKey string  `bson:"policyKey"`
	Cohort    int     `bson:"cohort"`
	Quality   float64 `bson:"quality"`
	ErrorRate float64 `bson:"errorRate"`
}
type Metrics struct {
	Cohort       int           `bson:"cohort"`
	Evidence     int           `bson:"evidence"`
	Quality      float64       `bson:"quality"`
	ErrorRate    float64       `bson:"errorRate"`
	MaxDisparity float64       `bson:"maxDisparity"`
	Slices       []SliceMetric `bson:"slices"`
}
type ModelCard struct {
	Version        uint64 `bson:"version"`
	Purpose        string `bson:"purpose"`
	EvaluationRef  string `bson:"evaluationRef"`
	LimitationsRef string `bson:"limitationsRef"`
	Owner          string `bson:"owner"`
}
type Approval struct {
	ReviewerKey        string    `bson:"reviewerKey"`
	EvaluationRevision uint64    `bson:"evaluationRevision"`
	ApprovedAt         time.Time `bson:"approvedAt"`
	ExpiresAt          time.Time `bson:"expiresAt"`
}
type Evaluation struct {
	id, candidate    string
	candidateVersion uint64
	snapshot         Snapshot
	metrics          Metrics
	card             ModelCard
	approval         Approval
	revision         uint64
	events           []Event
	commands         []Applied
}
type State struct {
	ID, Candidate    string
	CandidateVersion uint64
	Snapshot         Snapshot
	Metrics          Metrics
	Card             ModelCard
	Approval         Approval
	Revision         uint64
	Events           []Event
	Commands         []Applied
}

func Create(id, candidate string, version uint64, c Command) (Evaluation, error) {
	e := Evaluation{id: id, candidate: candidate, candidateVersion: version}
	if !opaque.MatchString(id) || !slug.MatchString(candidate) || version == 0 || c.ExpectedRevision != 0 {
		return Evaluation{}, ErrInvalid
	}
	if err := e.apply(c, "create", candidate, strconv.FormatUint(version, 10)); err != nil {
		return Evaluation{}, err
	}
	return e, nil
}
func Rehydrate(s State) (Evaluation, error) {
	e := Evaluation{id: s.ID, candidate: s.Candidate, candidateVersion: s.CandidateVersion, snapshot: s.Snapshot, metrics: s.Metrics, card: s.Card, approval: s.Approval, revision: s.Revision, events: append([]Event(nil), s.Events...), commands: append([]Applied(nil), s.Commands...)}
	if !opaque.MatchString(e.id) || !slug.MatchString(e.candidate) || e.candidateVersion == 0 || e.revision == 0 || len(e.events) != int(e.revision) || len(e.commands) != int(e.revision) {
		return Evaluation{}, ErrInvalid
	}
	return e, nil
}
func (e Evaluation) Record(snapshot Snapshot, metrics Metrics, c Command) (Evaluation, error) {
	action := "evaluate"
	if replay, err := e.replay(c, action, snapshot.ID, strconv.FormatUint(snapshot.Version, 10), strconv.FormatUint(snapshot.ConsentVersion, 10)); replay || err != nil {
		return e, err
	}
	if !validSnapshot(snapshot, c.At) || !validMetrics(metrics) {
		return Evaluation{}, ErrInvalid
	}
	e = e.clone()
	e.snapshot, e.metrics, e.approval = snapshot, cloneMetrics(metrics), Approval{}
	return e, e.apply(c, action, snapshot.ID, strconv.FormatUint(snapshot.Version, 10), strconv.FormatUint(snapshot.ConsentVersion, 10))
}
func (e Evaluation) AttachCard(card ModelCard, c Command) (Evaluation, error) {
	action := "card:" + strconv.FormatUint(card.Version, 10)
	if replay, err := e.replay(c, action, card.Purpose, card.EvaluationRef, card.LimitationsRef, card.Owner); replay || err != nil {
		return e, err
	}
	if e.snapshot.ID == "" || !validCard(card) {
		return Evaluation{}, ErrTransition
	}
	e = e.clone()
	e.card, e.approval = card, Approval{}
	return e, e.apply(c, action, card.Purpose, card.EvaluationRef, card.LimitationsRef, card.Owner)
}
func (e Evaluation) Approve(reviewer string, expires time.Time, c Command) (Evaluation, error) {
	action := "approve:" + expires.UTC().Format(time.RFC3339Nano)
	if replay, err := e.replay(c, action, reviewer); replay || err != nil {
		return e, err
	}
	if !opaque.MatchString(reviewer) || !e.gatesPass() || expires.IsZero() || !expires.After(c.At) || expires.Sub(c.At) > MaxApprovalLife {
		return Evaluation{}, ErrTransition
	}
	e = e.clone()
	// The approval binds to the revision produced by this approval command.
	e.approval = Approval{ReviewerKey: reviewer, EvaluationRevision: e.revision + 1, ApprovedAt: c.At.UTC(), ExpiresAt: expires.UTC()}
	return e, e.apply(c, action, reviewer)
}
func (e Evaluation) Ready(at time.Time) bool {
	return !at.IsZero() && e.gatesPass() && e.approval.ReviewerKey != "" && e.approval.EvaluationRevision == e.revision && !at.UTC().Before(e.approval.ApprovedAt) && at.UTC().Before(e.approval.ExpiresAt)
}
func (e Evaluation) gatesPass() bool {
	return validSnapshot(e.snapshot, e.snapshot.EvaluatedAt) && validMetrics(e.metrics) && validCard(e.card)
}
func validSnapshot(s Snapshot, at time.Time) bool {
	return opaque.MatchString(s.ID) && s.Version > 0 && s.ConsentVersion > 0 && !s.EvaluatedAt.IsZero() && !s.EvaluatedAt.After(at.UTC())
}
func validMetrics(m Metrics) bool {
	if nonFinite(m.Quality, m.ErrorRate, m.MaxDisparity) || m.Cohort < MinCohort || m.Evidence < MinEvidence || m.Quality < MinQuality || m.Quality > MaxQuality || m.ErrorRate < 0 || m.ErrorRate > MaxErrorRate || m.MaxDisparity < 0 || m.MaxDisparity > MaxDisparity || len(m.Slices) == 0 || len(m.Slices) > MaxSlices {
		return false
	}
	keys := make([]string, 0, len(m.Slices))
	for _, x := range m.Slices {
		if nonFinite(x.Quality, x.ErrorRate) || !slug.MatchString(x.PolicyKey) || x.Cohort < MinSliceCohort || x.Quality < 0 || x.Quality > 1 || x.ErrorRate < 0 || x.ErrorRate > 1 {
			return false
		}
		keys = append(keys, x.PolicyKey)
	}
	slices.Sort(keys)
	return !hasDuplicate(keys)
}
func nonFinite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return true
		}
	}
	return false
}
func hasDuplicate(v []string) bool {
	for i := 1; i < len(v); i++ {
		if v[i] == v[i-1] {
			return true
		}
	}
	return false
}
func validCard(c ModelCard) bool {
	return c.Version > 0 && slug.MatchString(c.Purpose) && opaque.MatchString(c.EvaluationRef) && opaque.MatchString(c.LimitationsRef) && slug.MatchString(c.Owner)
}
func cloneMetrics(m Metrics) Metrics { m.Slices = append([]SliceMetric(nil), m.Slices...); return m }
func (e Evaluation) clone() Evaluation {
	e.metrics = cloneMetrics(e.metrics)
	e.events, e.commands = append([]Event(nil), e.events...), append([]Applied(nil), e.commands...)
	return e
}
func (e *Evaluation) apply(c Command, action string, values ...string) error {
	if !opaque.MatchString(c.ID) || c.At.IsZero() || c.ExpectedRevision != e.revision {
		return ErrStale
	}
	fp := fingerprint(e.id, c, action, values...)
	e.revision++
	e.events = append(e.events, Event{Sequence: e.revision, CommandID: c.ID, Action: action, At: c.At.UTC()})
	e.commands = append(e.commands, Applied{ID: c.ID, Fingerprint: fp, Revision: e.revision})
	return nil
}
func (e Evaluation) replay(c Command, action string, values ...string) (bool, error) {
	fp := fingerprint(e.id, c, action, values...)
	for _, a := range e.commands {
		if a.ID == c.ID {
			if a.Fingerprint != fp {
				return false, ErrMismatch
			}
			return true, nil
		}
	}
	return false, nil
}
func fingerprint(id string, c Command, action string, values ...string) string {
	sum := sha256.Sum256([]byte(id + "\x00" + c.ID + "\x00" + action + "\x00" + strings.Join(values, "\x00") + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10) + "\x00" + c.At.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(sum[:])
}

func (e Evaluation) ID() string               { return e.id }
func (e Evaluation) Candidate() string        { return e.candidate }
func (e Evaluation) CandidateVersion() uint64 { return e.candidateVersion }
func (e Evaluation) Snapshot() Snapshot       { return e.snapshot }
func (e Evaluation) Metrics() Metrics         { return cloneMetrics(e.metrics) }
func (e Evaluation) Card() ModelCard          { return e.card }
func (e Evaluation) Approval() Approval       { return e.approval }
func (e Evaluation) Revision() uint64         { return e.revision }
func (e Evaluation) Events() []Event          { return append([]Event(nil), e.events...) }
func (e Evaluation) Commands() []Applied      { return append([]Applied(nil), e.commands...) }
