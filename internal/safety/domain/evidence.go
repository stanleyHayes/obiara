// Package domain models evidence access for T&S desks (E12-S03; FR-801;
// Doc 09 §3): least-exposure bundles with automatic redaction of
// non-case parties, and an immutable audit record for every access.
package domain

import (
	"errors"
	"strings"
	"time"
)

// Purpose is the declared reason for evidence access (just-in-time,
// purpose-scoped per plan §15 insider controls).
type Purpose string

const (
	PurposeTriage Purpose = "triage"
	PurposeAppeal Purpose = "appeal"
	PurposeLegal  Purpose = "legal"
)

var ErrInvalidPurpose = errors.New("evidence access purpose must be triage, appeal or legal")

// AccessRecord is the immutable audit entry for one evidence access
// (agent_plan.md §4: all privileged actions are auditable).
type AccessRecord struct {
	ID         string
	CaseID     string
	AgentID    string
	Purpose    Purpose
	AccessedAt time.Time
}

// NewAccessRecord validates an access.
func NewAccessRecord(id, caseID, agentID string, purpose Purpose, now time.Time) (AccessRecord, error) {
	switch purpose {
	case PurposeTriage, PurposeAppeal, PurposeLegal:
	default:
		return AccessRecord{}, ErrInvalidPurpose
	}
	if strings.TrimSpace(agentID) == "" {
		return AccessRecord{}, ErrAgentRequired
	}
	return AccessRecord{ID: id, CaseID: caseID, AgentID: agentID, Purpose: purpose, AccessedAt: now.UTC()}, nil
}

// Bundle is the least-exposure evidence projection for one case.
type Bundle struct {
	CaseID      string
	Tier        Tier
	Category    Category
	Surface     Surface
	ContextRef  string
	SubjectID   string
	Description string // redacted free text
}

const redactionMask = "[redacted]"

// Redact masks identifiers of non-case parties in free text: anything
// looking like a phone number, email address, or @handle. The reported
// subject stays visible to the desk; bystanders do not (Doc 09 §3:
// auto-redaction of third parties).
func Redact(text string) string {
	for _, pattern := range identifierPatterns {
		text = pattern.ReplaceAllString(text, redactionMask)
	}
	return text
}

// BuildBundle projects a report into a desk bundle with redacted text.
func BuildBundle(report Report) Bundle {
	return Bundle{
		CaseID:      report.ID(),
		Tier:        report.Tier(),
		Category:    report.Category(),
		Surface:     report.Surface(),
		ContextRef:  report.ContextRef(),
		SubjectID:   report.SubjectID(),
		Description: Redact(report.Reason()),
	}
}
