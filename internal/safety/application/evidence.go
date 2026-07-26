package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

// AccessAudit is the immutable evidence-access audit store.
type AccessAudit interface {
	Append(context.Context, domain.AccessRecord) error
	CountForCase(context.Context, string) (int, error)
}

// EvidenceService serves least-exposure bundles (E12-S03). The audit
// record is written before any data is read: even a failed access is
// accounted for (plan §15 insider-access controls).
type EvidenceService struct {
	reports ReportRepository
	audit   AccessAudit
	now     func() time.Time
	newID   func() string
}

func NewEvidenceService(reports ReportRepository, audit AccessAudit, now func() time.Time, newID func() string) EvidenceService {
	return EvidenceService{reports: reports, audit: audit, now: now, newID: newID}
}

// View returns the redacted bundle for a case after auditing the access.
func (service EvidenceService) View(ctx context.Context, caseID, agentID string, purpose domain.Purpose) (domain.Bundle, error) {
	record, err := domain.NewAccessRecord(service.newID(), caseID, agentID, purpose, service.now())
	if err != nil {
		return domain.Bundle{}, err
	}
	if err := service.audit.Append(ctx, record); err != nil {
		return domain.Bundle{}, err
	}

	report, err := service.reports.FindByID(ctx, caseID)
	if err != nil {
		return domain.Bundle{}, err
	}
	return domain.BuildBundle(report), nil
}
