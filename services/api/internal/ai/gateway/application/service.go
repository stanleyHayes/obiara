package application

import (
	"context"
	"errors"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/ai/gateway/domain"
)

type Service struct {
	consent  ConsentGate
	redactor Redactor
	vendor   Vendor
	audit    AuditSink
	clock    Clock
}

type ExecuteCommand struct {
	RequestID       string
	ActorKey        string
	Input           string
	MaxOutputTokens int
	Policy          domain.Policy
}

func NewService(consent ConsentGate, redactor Redactor, vendor Vendor, audit AuditSink, clock Clock) Service {
	return Service{consent: consent, redactor: redactor, vendor: vendor, audit: audit, clock: clock}
}

func (s Service) Execute(ctx context.Context, cmd ExecuteCommand) (string, error) {
	if strings.TrimSpace(cmd.RequestID) == "" || strings.TrimSpace(cmd.ActorKey) == "" {
		return "", domain.ErrInvalid
	}
	if err := cmd.Policy.Authorize(len([]byte(cmd.Input)), cmd.MaxOutputTokens); err != nil {
		s.record(ctx, cmd, 0, 0, "denied", reason(err))
		return "", err
	}
	if err := s.consent.RequireCurrent(ctx, cmd.ActorKey, cmd.Policy.Purpose, cmd.Policy.DataClasses); err != nil {
		s.record(ctx, cmd, 0, 0, "denied", "consent")
		return "", domain.ErrNotConsented
	}
	safeInput, redactions, err := s.redactor.Redact(ctx, cmd.Input, cmd.Policy.DataClasses)
	if err != nil || strings.TrimSpace(safeInput) == "" {
		s.record(ctx, cmd, redactions, 0, "denied", "redaction")
		return "", domain.ErrDenied
	}
	response, err := s.vendor.Invoke(ctx, VendorRequest{
		RequestID: cmd.RequestID, Model: cmd.Policy.Model.Name,
		Capability: cmd.Policy.Capability, Input: safeInput,
		MaxOutputTokens: cmd.MaxOutputTokens,
	})
	if err != nil {
		s.record(ctx, cmd, redactions, 0, "failed", "vendor")
		return "", err
	}
	if strings.TrimSpace(response.Output) == "" {
		s.record(ctx, cmd, redactions, 0, "failed", "empty_output")
		return "", domain.ErrDenied
	}
	if err := s.record(ctx, cmd, redactions, len([]byte(response.Output)), "completed", ""); err != nil {
		return "", err
	}
	return response.Output, nil
}

func (s Service) record(ctx context.Context, cmd ExecuteCommand, redactions, outputBytes int, outcome, code string) error {
	return s.audit.Record(ctx, AuditRecord{
		RequestID: cmd.RequestID, ActorKey: cmd.ActorKey,
		Capability: cmd.Policy.Capability, Purpose: cmd.Policy.Purpose,
		Vendor: cmd.Policy.Model.Vendor, Model: cmd.Policy.Model.Name, Region: cmd.Policy.Region,
		InputBytes: len([]byte(cmd.Input)), OutputBytes: outputBytes, Redactions: redactions,
		Outcome: outcome, ReasonCode: code, OccurredAt: s.clock.Now().UTC(),
	})
}

func reason(err error) string {
	switch {
	case errors.Is(err, domain.ErrTooLarge):
		return "bounds"
	case errors.Is(err, domain.ErrDenied):
		return "policy"
	default:
		return "invalid"
	}
}
