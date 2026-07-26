// Package safety builds T&S cases from filed-report outbox events
// (E12-S02). Every safety.report_filed event becomes one tiered case with
// its SLA deadline (Doc 09 §3). The builder is replay-safe at two levels:
// inbox dedup per event and the unique reportId case index.
package safety

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/stanleyHayes/obiara/internal/platform/inbox"
	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/safety/application"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
	jobsapplication "github.com/stanleyHayes/obiara/services/worker/internal/jobs/application"
)

// ReportReader loads filed reports (safety context read model).
type ReportReader interface {
	FindByID(context.Context, string) (domain.Report, error)
}

// EventLister queries outbox records by event type.
type EventLister interface {
	FindByEventType(context.Context, string, int) ([]outbox.Record, error)
}

// CaseBuilder consumes report events into tiered cases.
type CaseBuilder struct {
	events  EventLister
	reports ReportReader
	cases   application.CaseService
	inbox   *inbox.Store
}

func NewCaseBuilder(events EventLister, reports ReportReader, cases application.CaseService, inboxStore *inbox.Store) CaseBuilder {
	return CaseBuilder{events: events, reports: reports, cases: cases, inbox: inboxStore}
}

type reportPayload struct {
	ReportID string `json:"reportId"`
}

// RunOnce processes a batch of filed-report events.
func (builder CaseBuilder) RunOnce(ctx context.Context, batchSize int) error {
	records, err := builder.events.FindByEventType(ctx, "safety.report_filed", batchSize)
	if err != nil {
		return err
	}
	for _, record := range records {
		seen, err := builder.inbox.AlreadyProcessed(ctx, "safety.casebuilder", record.ID)
		if err != nil || seen {
			return err
		}

		var payload reportPayload
		if err := json.Unmarshal(record.Payload, &payload); err != nil {
			return fmt.Errorf("decode report event %q: %w", record.ID, err)
		}
		report, err := builder.reports.FindByID(ctx, payload.ReportID)
		if err != nil {
			return fmt.Errorf("load report %q: %w", payload.ReportID, err)
		}
		if _, err := builder.cases.Open(ctx, report); err != nil && !errors.Is(err, domain.ErrReportAlreadyQueued) {
			return fmt.Errorf("open case for report %q: %w", payload.ReportID, err)
		}
	}
	return nil
}

// NewBuilderJob builds the scheduled case-builder job.
func NewBuilderJob(builder CaseBuilder, batchSize int, interval time.Duration) jobsapplication.Job {
	return jobsapplication.Job{
		Name:        "safety.casebuilder",
		Interval:    interval,
		Timeout:     time.Minute,
		MaxAttempts: 5,
		Run: func(ctx context.Context) error {
			return builder.RunOnce(ctx, batchSize)
		},
	}
}
