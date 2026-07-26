package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/honesty/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Find(context.Context, string) (domain.Ribbon, error)
	Save(context.Context, domain.Ribbon, uint64, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
