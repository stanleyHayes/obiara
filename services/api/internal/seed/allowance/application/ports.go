package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/allowance/domain"
)

var (
	ErrNotFound         = errors.New("allowance ledger not found")
	ErrConcurrentChange = errors.New("allowance ledger changed concurrently")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Ledger) error
	Find(context.Context, string) (domain.Ledger, error)
	Save(context.Context, domain.Ledger, int64) error
}

type SubjectKeyer interface{ Key(string) (string, error) }
