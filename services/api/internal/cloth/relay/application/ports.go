package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/relay/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Find(context.Context, string) (domain.Relay, error)
	Save(context.Context, domain.Relay, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type ReviewerAuthorization interface {
	Allowed(context.Context, string, string) (bool, error)
}
type ConsentRevalidator interface {
	Current(context.Context, string, string, string, []string) (bool, error)
}
