package domain

import (
	"errors"
	"regexp"
	"time"
)

var ErrInvalid = errors.New("invalid analytics retention input")
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
var subject = regexp.MustCompile(`^[a-f0-9]{32,64}$`)

const (
	PseudonymizeAfter = 90 * 24 * time.Hour
	MaxBatch          = 500
)

type PolicySpec struct {
	ID, ReviewID, ReviewerKey    string
	Version, PseudonymKeyVersion uint64
	ReviewedAt                   time.Time
	BatchSize                    int
}
type Policy struct{ spec PolicySpec }

func NewPolicy(spec PolicySpec) (Policy, error) {
	spec.ReviewedAt = spec.ReviewedAt.UTC()
	if !token.MatchString(spec.ID) || !token.MatchString(spec.ReviewID) || !subject.MatchString(spec.ReviewerKey) || spec.Version == 0 || spec.PseudonymKeyVersion == 0 || spec.ReviewedAt.IsZero() || spec.BatchSize < 1 || spec.BatchSize > MaxBatch {
		return Policy{}, ErrInvalid
	}
	return Policy{spec}, nil
}
func (p Policy) Spec() PolicySpec { return p.spec }

type Candidate struct {
	ID, Name, SubjectRef string
	OccurredAt           time.Time
	PseudonymizedAt      time.Time
	PseudonymKeyVersion  uint64
}
type Action string

const (
	ActionKeep           Action = "keep"
	ActionPseudonymize   Action = "pseudonymize"
	ActionAggregateErase Action = "aggregate_erase"
)

type Decision struct {
	EventID                            string
	Action                             Action
	PolicyID                           string
	PolicyVersion, PseudonymKeyVersion uint64
	AggregateMonth, ProcessedAt        time.Time
}

func Decide(candidate Candidate, policy Policy, at time.Time) (Decision, error) {
	if !token.MatchString(candidate.Name) || candidate.ID == "" || !subject.MatchString(candidate.SubjectRef) || candidate.OccurredAt.IsZero() || at.IsZero() || candidate.OccurredAt.After(at) {
		return Decision{}, ErrInvalid
	}
	spec := policy.Spec()
	decision := Decision{EventID: candidate.ID, Action: ActionKeep, PolicyID: spec.ID, PolicyVersion: spec.Version, PseudonymKeyVersion: spec.PseudonymKeyVersion, ProcessedAt: at.UTC()}
	if !at.Before(candidate.OccurredAt.AddDate(0, 13, 0)) {
		decision.Action = ActionAggregateErase
		decision.AggregateMonth = time.Date(candidate.OccurredAt.Year(), candidate.OccurredAt.Month(), 1, 0, 0, 0, 0, time.UTC)
		return decision, nil
	}
	if at.Sub(candidate.OccurredAt) >= PseudonymizeAfter && candidate.PseudonymizedAt.IsZero() {
		decision.Action = ActionPseudonymize
	}
	return decision, nil
}

type AggregateKey struct {
	Name  string
	Month time.Time
}

func (d Decision) AggregateKey(name string) AggregateKey {
	return AggregateKey{Name: name, Month: d.AggregateMonth}
}
