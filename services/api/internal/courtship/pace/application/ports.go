package application

import (
	"context"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/pace/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Pace) error
	Find(context.Context, string) (domain.Pace, error)
	Save(context.Context, domain.Pace, uint64, string) error
}

type Keyer interface {
	Key(string, string) (string, error)
}
