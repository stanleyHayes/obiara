package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/reconciliation/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	AppendFact(context.Context, domain.StatementFact) error
	FindFactByEvent(context.Context, string) (domain.StatementFact, error)
	AppendAudit(context.Context, domain.Audit) error
	ListFactsForDay(context.Context, string) ([]domain.StatementFact, error)
	AppendCheckpoint(context.Context, domain.Checkpoint) error
	FindCheckpoint(context.Context, string) (domain.Checkpoint, error)
}
type SignatureVerifier interface {
	Verify(context.Context, SignedEnvelope) error
}
type Ledger interface {
	Proof(context.Context, string) (domain.LedgerProof, bool, error)
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }

type SignedEnvelope struct {
	Provider, EventID, Reference, LedgerCommand, Currency, Status string
	Minor, OccurredUnix                                           int64
	Signature                                                     string
}
