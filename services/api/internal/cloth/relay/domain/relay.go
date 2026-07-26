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

type Action string

const (
	ActionSubmit Action = "submit"
	ActionGrant  Action = "grant"
	ActionRevoke Action = "revoke"
)

var (
	ErrInvalid         = errors.New("relay unavailable")
	ErrDenied          = errors.New("relay unavailable")
	ErrStaleRevision   = errors.New("stale relay revision")
	ErrCommandMismatch = errors.New("relay replay mismatch")
	keyPattern         = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern          = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{7,127}$`)
	refPattern         = regexp.MustCompile(`^ref_[A-Za-z0-9_-]{20,160}$`)
)

type Command struct {
	ID, ActorKey, QuestionRef, PromptRef, ResponseRef string
	ExpectedRevision                                  uint64
	At                                                time.Time
}
type Question struct {
	QuestionRef, PromptRef string
	Grants                 map[string]string
}
type Audit struct {
	Sequence    uint64
	CommandID   string
	Action      Action
	QuestionRef string
	At          time.Time
}
type Applied struct {
	ID, Fingerprint string
	Revision        uint64
}
type Relay struct {
	id        string
	members   []string
	reviewer  string
	questions []Question
	audit     []Audit
	commands  []Applied
	revision  uint64
}
type State struct {
	ID        string
	Members   []string
	Reviewer  string
	Questions []Question
	Audit     []Audit
	Commands  []Applied
	Revision  uint64
}
type Projection struct{ QuestionRef, PromptRef, ResponseRef string }

func New(id string, members []string, reviewer string) (Relay, error) {
	m := append([]string(nil), members...)
	slices.Sort(m)
	m = slices.Compact(m)
	if !idPattern.MatchString(id) || len(m) != 2 || !keyPattern.MatchString(m[0]) || !keyPattern.MatchString(m[1]) || !keyPattern.MatchString(reviewer) || slices.Contains(m, reviewer) {
		return Relay{}, ErrInvalid
	}
	return Relay{id: id, members: m, reviewer: reviewer}, nil
}
func Rehydrate(s State) (Relay, error) {
	r, e := New(s.ID, s.Members, s.Reviewer)
	if e != nil {
		return Relay{}, e
	}
	r.questions = cloneQuestions(s.Questions)
	r.audit = append([]Audit(nil), s.Audit...)
	r.commands = append([]Applied(nil), s.Commands...)
	r.revision = s.Revision
	if len(r.audit) != int(r.revision) || len(r.commands) != int(r.revision) {
		return Relay{}, ErrInvalid
	}
	return r, nil
}
func (r Relay) Submit(c Command) (Relay, error) {
	if c.ActorKey != r.reviewer || !refPattern.MatchString(c.QuestionRef) || !refPattern.MatchString(c.PromptRef) || c.ResponseRef != "" {
		return Relay{}, ErrDenied
	}
	return r.apply(ActionSubmit, c)
}
func (r Relay) Grant(c Command) (Relay, error) {
	if !r.member(c.ActorKey) || !refPattern.MatchString(c.QuestionRef) || !refPattern.MatchString(c.ResponseRef) || c.PromptRef != "" || r.question(c.QuestionRef) == nil {
		return Relay{}, ErrDenied
	}
	return r.apply(ActionGrant, c)
}
func (r Relay) Revoke(c Command) (Relay, error) {
	if !r.member(c.ActorKey) || !refPattern.MatchString(c.QuestionRef) || c.PromptRef != "" || c.ResponseRef != "" || r.question(c.QuestionRef) == nil {
		return Relay{}, ErrDenied
	}
	return r.apply(ActionRevoke, c)
}
func (r Relay) apply(a Action, c Command) (Relay, error) {
	r.questions = cloneQuestions(r.questions)
	if !idPattern.MatchString(c.ID) || c.At.IsZero() {
		return Relay{}, ErrDenied
	}
	fp := fingerprint(r.id, a, c)
	for _, x := range r.commands {
		if x.ID == c.ID {
			if x.Fingerprint != fp {
				return Relay{}, ErrCommandMismatch
			}
			return r, nil
		}
	}
	if c.ExpectedRevision != r.revision {
		return Relay{}, ErrStaleRevision
	}
	if a == ActionSubmit {
		if r.question(c.QuestionRef) != nil {
			return Relay{}, ErrDenied
		}
		r.questions = append(r.questions, Question{QuestionRef: c.QuestionRef, PromptRef: c.PromptRef, Grants: map[string]string{}})
	} else {
		q := r.question(c.QuestionRef)
		if a == ActionGrant {
			q.Grants[c.ActorKey] = c.ResponseRef
		} else {
			delete(q.Grants, c.ActorKey)
		}
	}
	r.revision++
	r.audit = append(r.audit, Audit{r.revision, c.ID, a, c.QuestionRef, c.At.UTC()})
	r.commands = append(r.commands, Applied{c.ID, fp, r.revision})
	return r, nil
}
func (r Relay) Project(reviewer, questionRef string) (Projection, error) {
	if reviewer != r.reviewer {
		return Projection{}, ErrDenied
	}
	q := r.question(questionRef)
	if q == nil || len(q.Grants) != 2 {
		return Projection{}, ErrDenied
	}
	a, b := q.Grants[r.members[0]], q.Grants[r.members[1]]
	if a == "" || a != b {
		return Projection{}, ErrDenied
	}
	return Projection{q.QuestionRef, q.PromptRef, a}, nil
}
func (r Relay) question(ref string) *Question {
	for i := range r.questions {
		if r.questions[i].QuestionRef == ref {
			return &r.questions[i]
		}
	}
	return nil
}
func (r Relay) member(k string) bool { return slices.Contains(r.members, k) }
func fingerprint(id string, a Action, c Command) string {
	s := sha256.Sum256([]byte(id + "\x00" + string(a) + "\x00" + c.ID + "\x00" + c.ActorKey + "\x00" + c.QuestionRef + "\x00" + c.PromptRef + "\x00" + c.ResponseRef + "\x00" + strconv.FormatUint(c.ExpectedRevision, 10)))
	return hex.EncodeToString(s[:])
}
func cloneQuestions(in []Question) []Question {
	out := make([]Question, len(in))
	for i, q := range in {
		out[i] = Question{q.QuestionRef, q.PromptRef, map[string]string{}}
		for k, v := range q.Grants {
			out[i].Grants[k] = v
		}
	}
	return out
}
func (r Relay) ID() string            { return r.id }
func (r Relay) Members() []string     { return append([]string(nil), r.members...) }
func (r Relay) Reviewer() string      { return r.reviewer }
func (r Relay) Questions() []Question { return cloneQuestions(r.questions) }
func (r Relay) Audit() []Audit        { return append([]Audit(nil), r.audit...) }
func (r Relay) Commands() []Applied   { return append([]Applied(nil), r.commands...) }
func (r Relay) Revision() uint64      { return r.revision }
