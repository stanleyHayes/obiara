package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/ai/gateway/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type ConsentGate interface {
	RequireCurrent(context.Context, string, string, []domain.DataClass) error
}

type Redactor interface {
	Redact(context.Context, string, []domain.DataClass) (string, int, error)
}

type Vendor interface {
	Invoke(context.Context, VendorRequest) (VendorResponse, error)
}

type AuditSink interface {
	Record(context.Context, AuditRecord) error
}

type Clock interface{ Now() time.Time }

type VendorRequest struct {
	RequestID       string
	Model           string
	Capability      domain.Capability
	Input           string
	MaxOutputTokens int
}

type VendorResponse struct {
	Output          string
	VendorRequestID string
}

type AuditRecord struct {
	RequestID   string
	ActorKey    string
	Capability  domain.Capability
	Purpose     string
	Vendor      string
	Model       string
	Region      string
	InputBytes  int
	OutputBytes int
	Redactions  int
	Outcome     string
	ReasonCode  string
	OccurredAt  time.Time
}
