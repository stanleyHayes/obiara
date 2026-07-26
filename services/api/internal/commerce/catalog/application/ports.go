package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.SKU) error
	Find(context.Context, string, uint64) (domain.SKU, error)
	FindLatest(context.Context, string) (domain.SKU, error)
	FindByCommand(context.Context, string) (domain.SKU, error)
	Append(context.Context, domain.SKU, uint64, string) error
}
type Authority interface {
	RequireCatalogEditor(context.Context, string) error
}
type Keyer interface {
	Key(string, string) (string, error)
}
type IDSource interface{ NewID() string }
