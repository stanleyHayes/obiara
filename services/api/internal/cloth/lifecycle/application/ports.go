package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/lifecycle/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Find(context.Context, string) (domain.Lifecycle, error)
	Save(context.Context, domain.Lifecycle, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type LegalHold interface {
	DeletionAllowed(context.Context, string) (bool, error)
}
