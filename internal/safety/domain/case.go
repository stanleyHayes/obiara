package domain

import (
	"errors"
	"strings"
	"time"
)

// Queue is a T&S work queue (Doc 09 §3).
type Queue string

const (
	QueueTriage Queue = "triage"
	QueueCare   Queue = "care"
)

// CaseStatus is the case lifecycle.
type CaseStatus string

const (
	CaseQueued   CaseStatus = "queued"
	CaseInReview CaseStatus = "in_review"
	CaseResolved CaseStatus = "resolved"
)

var (
	ErrCaseIDRequired      = errors.New("case id is required")
	ErrCaseNotOpen         = errors.New("case is not open")
	ErrAgentRequired       = errors.New("reviewing agent is required")
	ErrOutcomeRequired     = errors.New("resolution outcome is required")
	ErrReportAlreadyQueued = errors.New("a case already exists for this report")
)

// SLADueAt computes the routing deadline from Doc 09 §3: Tier A p95 <8h,
// B <24h, C <72h, care immediate.
func SLADueAt(tier Tier, now time.Time) time.Time {
	switch tier {
	case TierA:
		return now.UTC().Add(8 * time.Hour)
	case TierB:
		return now.UTC().Add(24 * time.Hour)
	case TierC:
		return now.UTC().Add(72 * time.Hour)
	default: // TierD care: immediate
		return now.UTC()
	}
}

// QueueFor routes a tier to its queue; care goes straight to trained staff.
func QueueFor(tier Tier) Queue {
	if tier == TierD {
		return QueueCare
	}
	return QueueTriage
}

// Case is one queued T&S case built from a report.
type Case struct {
	id         string
	reportID   string
	subjectID  string
	tier       Tier
	queue      Queue
	slaDueAt   time.Time
	status     CaseStatus
	assignedTo string
	version    int64
	createdAt  time.Time
	resolvedAt *time.Time
}

// NewCaseFromReport builds a case with its SLA deadline.
func NewCaseFromReport(id string, report Report, now time.Time) (Case, error) {
	if strings.TrimSpace(id) == "" {
		return Case{}, ErrCaseIDRequired
	}
	return Case{
		id:        id,
		reportID:  report.ID(),
		subjectID: report.SubjectID(),
		tier:      report.Tier(),
		queue:     QueueFor(report.Tier()),
		slaDueAt:  SLADueAt(report.Tier(), now),
		status:    CaseQueued,
		version:   1,
		createdAt: now.UTC(),
	}, nil
}

// ReconstituteCase rebuilds a stored case without policy checks.
func ReconstituteCase(id, reportID, subjectID string, tier Tier, queue Queue, slaDueAt time.Time, status CaseStatus, assignedTo string, version int64, createdAt time.Time, resolvedAt *time.Time) Case {
	return Case{
		id: id, reportID: reportID, subjectID: subjectID, tier: tier, queue: queue,
		slaDueAt: slaDueAt, status: status, assignedTo: assignedTo,
		version: version, createdAt: createdAt, resolvedAt: resolvedAt,
	}
}

// Assign moves the case into review under an agent.
func (c *Case) Assign(agentID string, now time.Time) error {
	if c.status != CaseQueued {
		return ErrCaseNotOpen
	}
	if strings.TrimSpace(agentID) == "" {
		return ErrAgentRequired
	}
	c.status = CaseInReview
	c.assignedTo = agentID
	c.version++
	return nil
}

// Resolve closes the case with an outcome.
func (c *Case) Resolve(outcome, agentID string, now time.Time) error {
	if c.status != CaseInReview {
		return ErrCaseNotOpen
	}
	if strings.TrimSpace(agentID) == "" {
		return ErrAgentRequired
	}
	if strings.TrimSpace(outcome) == "" {
		return ErrOutcomeRequired
	}
	c.status = CaseResolved
	c.assignedTo = agentID
	resolved := now.UTC()
	c.resolvedAt = &resolved
	c.version++
	return nil
}

// Breached reports whether the SLA deadline passed without resolution.
func (c Case) Breached(now time.Time) bool {
	return c.status != CaseResolved && now.UTC().After(c.slaDueAt)
}

func (c Case) ID() string             { return c.id }
func (c Case) ReportID() string       { return c.reportID }
func (c Case) SubjectID() string      { return c.subjectID }
func (c Case) Tier() Tier             { return c.tier }
func (c Case) Queue() Queue           { return c.queue }
func (c Case) SLADueAt() time.Time    { return c.slaDueAt }
func (c Case) Status() CaseStatus     { return c.status }
func (c Case) AssignedTo() string     { return c.assignedTo }
func (c Case) Version() int64         { return c.version }
func (c Case) CreatedAt() time.Time   { return c.createdAt }
func (c Case) ResolvedAt() *time.Time { return c.resolvedAt }
