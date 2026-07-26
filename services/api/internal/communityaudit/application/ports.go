package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/communityaudit/domain"
	"time"
)

var (
	ErrDenied      = errors.New("community audit denied")
	ErrMFA         = errors.New("recent MFA required")
	ErrNotFound    = errors.New("community audit case not found")
	ErrConflict    = errors.New("community audit conflict")
	ErrUnavailable = errors.New("community audit unavailable")
)

type Capability string

const (
	CapQueue    Capability = "queue"
	CapRead     Capability = "read"
	CapEvidence Capability = "evidence"
	CapDecide   Capability = "decide"
)

type Authorizer interface {
	Authorize(context.Context, string, Capability) error
}
type MFAGate interface {
	Recent(context.Context, string, time.Time) bool
}
type Keyer interface {
	Key(string, string) (string, error)
}
type Summary struct {
	CaseID      string
	Kind        domain.Kind
	Status      domain.Status
	SummaryCode string
	CreatedAt   time.Time
}
type Evidence struct {
	Reference string
	Kind      domain.Kind
}
type Repository interface {
	List(context.Context, int) ([]domain.Case, error)
	Find(context.Context, string) (domain.Case, error)
	RecordEvidenceAccess(context.Context, string, string, string, time.Time) error
	Decide(context.Context, domain.Case, uint64, string) error
}
