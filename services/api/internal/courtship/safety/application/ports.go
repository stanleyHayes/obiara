package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/safety/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Find(context.Context, string) (domain.Safety, error)
	Save(context.Context, domain.Safety, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
