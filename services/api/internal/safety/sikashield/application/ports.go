package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/sikashield/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Catalog interface {
	Current(context.Context, string) (domain.Pattern, error)
}
type MetricsGate interface {
	Current(context.Context, string, uint64) (domain.Metrics, error)
}
type EvidenceVerifier interface {
	Revalidate(context.Context, string, domain.Source) error
}
type CaseRouter interface {
	OpenHumanCase(context.Context, domain.Decision) error
}
type Authority interface {
	RequireOfflineEvaluator(context.Context, string) error
}
