package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Posting) error
	FindByCommand(context.Context, string) (domain.Posting, error)
	ListLines(context.Context, string, domain.Currency) ([]domain.BookedLine, error)
}
type Authority interface {
	RequirePoster(context.Context, string, domain.Purpose) error
	RequireBalanceReader(context.Context, string, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
