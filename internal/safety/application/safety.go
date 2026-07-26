// Package application files reports and manages blocks (E12-S01). Every
// filed report also emits a durable outbox event for the T&S queue
// processors (E12-S02).
package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/internal/platform/outbox"
	"github.com/stanleyHayes/obiara/internal/safety/domain"
)

var (
	ErrReportNotFound = errors.New("report not found")
	ErrBlockExists    = errors.New("member is already blocked")
	ErrBlockNotFound  = errors.New("member is not blocked")
)

// ReportRepository persists reports.
type ReportRepository interface {
	Create(context.Context, domain.Report) error
	FindByID(context.Context, string) (domain.Report, error)
}

// BlockRepository persists blocks.
type BlockRepository interface {
	Add(context.Context, domain.Block) error
	Remove(context.Context, string, string) error
	Exists(context.Context, string, string) (bool, error)
}

// OutboxAppender is the durable-event port for queue processors.
type OutboxAppender interface {
	Append(context.Context, outbox.Record) error
}

// SafetyService is the intake boundary.
type SafetyService struct {
	reports ReportRepository
	blocks  BlockRepository
	outbox  OutboxAppender
	now     func() time.Time
	newID   func() string
}

func NewSafetyService(reports ReportRepository, blocks BlockRepository, outboxStore OutboxAppender, now func() time.Time, newID func() string) SafetyService {
	return SafetyService{reports: reports, blocks: blocks, outbox: outboxStore, now: now, newID: newID}
}

// File records a report and emits its queue event. The returned
// acknowledgement is the reporter-safe view.
func (service SafetyService) File(ctx context.Context, reporterID, subjectID string, category domain.Category, surface domain.Surface, contextRef, reason string) (string, domain.Tier, error) {
	report, err := domain.NewReport(service.newID(), reporterID, subjectID, category, surface, contextRef, reason, service.now())
	if err != nil {
		return "", "", err
	}
	if err := service.reports.Create(ctx, report); err != nil {
		return "", "", err
	}

	payload, err := json.Marshal(map[string]string{
		"reportId": report.ID(),
		"tier":     string(report.Tier()),
		"category": string(report.Category()),
	})
	if err != nil {
		return "", "", err
	}
	if err := service.outbox.Append(ctx, outbox.Record{
		ID:            "report_" + report.ID(),
		AggregateType: "report",
		AggregateID:   report.ID(),
		EventType:     "safety.report_filed",
		Payload:       payload,
		OccurredAt:    service.now(),
	}); err != nil {
		return "", "", err
	}

	id, tier, _ := report.Acknowledgement()
	return id, tier, nil
}

// Block blocks a member; duplicates are idempotent conflicts.
func (service SafetyService) Block(ctx context.Context, blockerID, blockedID string) error {
	block, err := domain.NewBlock(blockerID, blockedID, service.now())
	if err != nil {
		return err
	}
	return service.blocks.Add(ctx, block)
}

// Unblock removes a block.
func (service SafetyService) Unblock(ctx context.Context, blockerID, blockedID string) error {
	return service.blocks.Remove(ctx, blockerID, blockedID)
}

// IsBlocked reports whether a block edge exists.
func (service SafetyService) IsBlocked(ctx context.Context, blockerID, blockedID string) (bool, error) {
	return service.blocks.Exists(ctx, blockerID, blockedID)
}
