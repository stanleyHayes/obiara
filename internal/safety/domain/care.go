// Package domain models the care queue (E12-S05; Doc 09 §5): distress
// signals route immediately to trained humans with resource-first
// scripts — no diagnostic language, and no punitive action ever attaches
// to seeking help. Closure events can quieten notifications for 72 hours
// (heartbreak-aware offboarding).
package domain

import (
	"errors"
	"strings"
	"time"
)

// Signal is a care-triggering event class.
type Signal string

const (
	SignalDistressReport     Signal = "distress_report"
	SignalSelfHarmIndication Signal = "self_harm_indication"
	SignalVictimReport       Signal = "victim_report"
	SignalOkyeameEscalation  Signal = "okyeame_escalation"
	SignalClosure            Signal = "closure"
)

// CareStatus is the care case lifecycle.
type CareStatus string

const (
	CareOpen     CareStatus = "open"
	CareEngaged  CareStatus = "engaged"
	CareResolved CareStatus = "resolved"
)

// ScriptKey is an approved resource-first script (Doc 09 §5: local
// helpline directory per market, counselor referral, support content).
type ScriptKey string

const (
	ScriptHelplineDirectory ScriptKey = "helpline_directory_gh"
	ScriptCounselorReferral ScriptKey = "counselor_referral"
	ScriptSupportContent    ScriptKey = "support_content"
	ScriptClosureQuietening ScriptKey = "closure_quietening"
)

// QuieteningWindow is the 72-hour notification quietening after closure.
const QuieteningWindow = 72 * time.Hour

var (
	ErrInvalidSignal     = errors.New("unknown care signal")
	ErrCareSubjectNeeded = errors.New("care subject is required")
	ErrScriptRequired    = errors.New("resolution needs at least one script")
	ErrInvalidScript     = errors.New("unknown resource script")
)

// CareCase is one routed care case. It is never linked to enforcement:
// care and punishment are separate systems (Doc 09 §5).
type CareCase struct {
	id         string
	subjectID  string
	signal     Signal
	status     CareStatus
	scripts    []ScriptKey
	version    int64
	createdAt  time.Time
	resolvedAt *time.Time
}

// NewCareCase routes a signal. Care flags are always immediate priority.
func NewCareCase(id, subjectID string, signal Signal, now time.Time) (CareCase, error) {
	if strings.TrimSpace(subjectID) == "" {
		return CareCase{}, ErrCareSubjectNeeded
	}
	switch signal {
	case SignalDistressReport, SignalSelfHarmIndication, SignalVictimReport, SignalOkyeameEscalation, SignalClosure:
	default:
		return CareCase{}, ErrInvalidSignal
	}
	return CareCase{id: id, subjectID: subjectID, signal: signal, status: CareOpen, version: 1, createdAt: now.UTC()}, nil
}

// ReconstituteCareCase rebuilds a stored case without policy checks.
func ReconstituteCareCase(id, subjectID string, signal Signal, status CareStatus, scripts []ScriptKey, version int64, createdAt time.Time, resolvedAt *time.Time) CareCase {
	return CareCase{id: id, subjectID: subjectID, signal: signal, status: status, scripts: scripts, version: version, createdAt: createdAt, resolvedAt: resolvedAt}
}

// Engage marks a trained human on the case.
func (c *CareCase) Engage() error {
	if c.status != CareOpen {
		return ErrCaseNotOpen
	}
	c.status = CareEngaged
	c.version++
	return nil
}

// Resolve closes the case with the scripts the human used.
func (c *CareCase) Resolve(scripts []ScriptKey, now time.Time) error {
	if c.status != CareEngaged {
		return ErrCaseNotOpen
	}
	if len(scripts) == 0 {
		return ErrScriptRequired
	}
	for _, script := range scripts {
		switch script {
		case ScriptHelplineDirectory, ScriptCounselorReferral, ScriptSupportContent, ScriptClosureQuietening:
		default:
			return ErrInvalidScript
		}
	}
	c.status = CareResolved
	c.scripts = scripts
	resolved := now.UTC()
	c.resolvedAt = &resolved
	c.version++
	return nil
}

// NeedsQuietening reports whether this case's signal quiets notifications
// (Doc 09 §5: closure events trigger optional support content and a 72-h
// notification quietening).
func (c CareCase) NeedsQuietening() bool {
	return c.signal == SignalClosure
}

func (c CareCase) ID() string             { return c.id }
func (c CareCase) SubjectID() string      { return c.subjectID }
func (c CareCase) Signal() Signal         { return c.signal }
func (c CareCase) Status() CareStatus     { return c.status }
func (c CareCase) Scripts() []ScriptKey   { return c.scripts }
func (c CareCase) Version() int64         { return c.version }
func (c CareCase) CreatedAt() time.Time   { return c.createdAt }
func (c CareCase) ResolvedAt() *time.Time { return c.resolvedAt }
