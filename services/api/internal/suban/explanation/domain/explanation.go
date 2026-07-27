package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	suban "github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
)

var ErrInvalid = errors.New("invalid Suban explanation or appeal")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

const MaxVisibleEvents = 200

type Effect string

const (
	EffectKeepsWord            Effect = "supports_keeps_word"
	EffectGracious             Effect = "supports_gracious"
	EffectTrustedVoucher       Effect = "supports_trusted_voucher"
	EffectPractice             Effect = "supports_character_practice"
	EffectSuppresses           Effect = "suppresses_marks_during_review_window"
	EffectPermanentSuppression Effect = "suppresses_marks_permanently"
)

type VisibleEvent struct {
	ID             string
	Kind           suban.Kind
	Effect         Effect
	SourceCategory string
	OccurredAt     time.Time
}
type Explanation struct {
	SubjectKey  string
	Marks       []suban.Mark
	Events      []VisibleEvent
	GeneratedAt time.Time
}

func Explain(subjectKey string, events []suban.Event, at time.Time) (Explanation, error) {
	if strings.TrimSpace(subjectKey) == "" || len(events) > MaxVisibleEvents || at.IsZero() {
		return Explanation{}, ErrInvalid
	}
	copyEvents := append([]suban.Event(nil), events...)
	seen := map[string]bool{}
	slices.SortFunc(copyEvents, func(a, b suban.Event) int {
		if x := a.OccurredAt.Compare(b.OccurredAt); x != 0 {
			return x
		}
		return strings.Compare(a.ID, b.ID)
	})
	visible := make([]VisibleEvent, 0, len(copyEvents))
	for _, event := range copyEvents {
		if event.SubjectID != subjectKey || strings.TrimSpace(event.ID) == "" || seen[event.ID] || event.OccurredAt.IsZero() || !safeSource(event.Provenance.Source) {
			return Explanation{}, ErrInvalid
		}
		seen[event.ID] = true
		effect, ok := effectFor(event.Kind)
		if !ok {
			return Explanation{}, ErrInvalid
		}
		visible = append(visible, VisibleEvent{event.ID, event.Kind, effect, event.Provenance.Source, event.OccurredAt.UTC()})
	}
	marks := suban.ComputeMarks(copyEvents, at)
	slices.Sort(marks)
	return Explanation{subjectKey, marks, visible, at.UTC()}, nil
}
func effectFor(kind suban.Kind) (Effect, bool) {
	switch kind {
	case suban.KindMeetingFollowThrough:
		return EffectKeepsWord, true
	case suban.KindKindClosure, suban.KindGraciousDecline:
		return EffectGracious, true
	case suban.KindCleanVouch:
		return EffectTrustedVoucher, true
	case suban.KindPauseStone, suban.KindThemeCompleted:
		return EffectPractice, true
	case suban.KindFraudFinding:
		return EffectPermanentSuppression, true
	case suban.KindGhostPattern, suban.KindHarassmentFinding, suban.KindVouchStakeLoss:
		return EffectSuppresses, true
	}
	return "", false
}
func safeSource(source string) bool {
	return slices.Contains([]string{"meeting", "closure", "vouch", "panel", "room", "theme"}, source)
}

type Reason string

const (
	ReasonWrongSubject      Reason = "wrong_subject"
	ReasonEventInaccurate   Reason = "event_inaccurate"
	ReasonFindingOverturned Reason = "finding_overturned"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusUpheld     Status = "upheld"
	StatusOverturned Status = "overturned"
)

type Audit struct {
	Sequence                             uint64
	Kind, CommandID, ActorKey, Reference string
	At                                   time.Time
}
type State struct {
	ID, SubjectKey, EventID   string
	Reason                    Reason
	Status                    Status
	FiledAt, ResolvedAt       time.Time
	ReviewerKey, ReasoningRef string
	Revision                  uint64
	Audit                     []Audit
	AppliedIDs                []string
}
type Appeal struct{ state State }

func File(id, subject, eventID string, reason Reason, command string, at time.Time) (Appeal, error) {
	if !opaque.MatchString(id) || strings.TrimSpace(subject) == "" || strings.TrimSpace(eventID) == "" || !validReason(reason) || !token.MatchString(command) || at.IsZero() {
		return Appeal{}, ErrInvalid
	}
	s := State{ID: id, SubjectKey: subject, EventID: eventID, Reason: reason, Status: StatusPending, FiledAt: at.UTC(), Revision: 1, AppliedIDs: []string{command}}
	s.Audit = []Audit{{1, "filed", command, subject, eventID, at.UTC()}}
	return Appeal{s}, nil
}
func Rehydrate(s State) (Appeal, error) {
	s.Audit = append([]Audit(nil), s.Audit...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	if !validState(s) {
		return Appeal{}, ErrInvalid
	}
	return Appeal{s}, nil
}
func (a Appeal) Resolve(status Status, reviewer, reasoningRef, command string, at time.Time) (Appeal, error) {
	if a.state.Status != StatusPending || (status != StatusUpheld && status != StatusOverturned) || reviewer == a.state.SubjectKey || !opaque.MatchString(reviewer) || !opaque.MatchString(reasoningRef) || !token.MatchString(command) || at.IsZero() || slices.Contains(a.state.AppliedIDs, command) {
		return Appeal{}, ErrInvalid
	}
	n := a.State()
	n.Status = status
	n.ReviewerKey = reviewer
	n.ReasoningRef = reasoningRef
	n.ResolvedAt = at.UTC()
	n.Revision++
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Audit = append(n.Audit, Audit{n.Revision, "resolved", command, reviewer, reasoningRef, at.UTC()})
	return Appeal{n}, nil
}
func (a Appeal) State() State {
	s := a.state
	s.Audit = append([]Audit(nil), s.Audit...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	return s
}
func (a Appeal) ID() string       { return a.state.ID }
func (a Appeal) Revision() uint64 { return a.state.Revision }
func validReason(r Reason) bool {
	return r == ReasonWrongSubject || r == ReasonEventInaccurate || r == ReasonFindingOverturned
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || strings.TrimSpace(s.SubjectKey) == "" || strings.TrimSpace(s.EventID) == "" || !validReason(s.Reason) || (s.Status != StatusPending && s.Status != StatusUpheld && s.Status != StatusOverturned) || s.Revision == 0 || len(s.Audit) != int(s.Revision) || len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	for i, x := range s.Audit {
		if x.Sequence != uint64(i+1) || x.CommandID != s.AppliedIDs[i] || !token.MatchString(x.CommandID) || x.At.IsZero() {
			return false
		}
	}
	return true
}
func Appealable(kind suban.Kind) bool {
	return kind == suban.KindGhostPattern || kind == suban.KindHarassmentFinding || kind == suban.KindFraudFinding || kind == suban.KindVouchStakeLoss
}
