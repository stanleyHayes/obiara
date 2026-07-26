package application

import (
	"context"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type Repository interface {
	ListQueued(context.Context, int) ([]CaseSummary, error)
	Detail(context.Context, string) (CaseDetail, error)
	AccessEvidence(context.Context, EvidenceAccess) (Evidence, error)
	Decide(context.Context, DecisionCommand) (DecisionResult, error)
}

type Keyer interface {
	Key(namespace, value string) (string, error)
}

type EvidenceAccess struct {
	CaseID        string
	ActorID       string
	Purpose       string
	Reason        string
	CorrelationID string
	OccurredAt    time.Time
}

type DecisionCommand struct {
	CaseID          string
	ActorID         string
	Outcome         Outcome
	Reason          string
	IdempotencyKey  string
	ExpectedVersion int64
	CorrelationID   string
	OccurredAt      time.Time
}
