package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/anomaly/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type RuleCatalog interface {
	Current(context.Context, domain.Shape) (domain.Rule, error)
}
type ConsentVerifier interface {
	Revalidate(context.Context, string) error
}
type Authority interface {
	RequireOfflineEvaluator(context.Context, string) error
	RequireHumanRoute(context.Context, string, string) error
}
type CaseRouter interface {
	OpenHumanCase(context.Context, domain.Decision) error
}
