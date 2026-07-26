package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/matching/evaluation/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Evaluation) error
	Find(context.Context, string) (domain.Evaluation, error)
	FindByCommand(context.Context, string) (domain.Evaluation, error)
	Append(context.Context, domain.Evaluation, uint64, string) error
}
type SnapshotVerifier interface {
	Revalidate(context.Context, domain.Snapshot) error
}
type SlicePolicy interface {
	RequireApproved(context.Context, []string) error
}
type Authority interface {
	RequireEvaluator(context.Context, string) error
	RequireHumanApprover(context.Context, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
