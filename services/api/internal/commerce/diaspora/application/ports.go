package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Authority interface {
	RequireDiasporaCheckout(context.Context, string) error
}
type Catalog interface {
	CurrentDiasporaQuote(context.Context, string, domain.Currency) (domain.Quote, error)
}

// ConfirmationVerifier is the isolated PCI/provider boundary. Implementations
// verify already-received facts; this service never performs a network charge.
type ConfirmationVerifier interface {
	Verify(context.Context, ProviderConfirmation) error
}
type ProviderConfirmation struct {
	CheckoutID, RequestRef, ProviderRef string
	Currency                            domain.Currency
	AmountMinor                         int64
	Succeeded                           bool
	OccurredAt                          time.Time
	Evidence                            string
}

// SettlementLedger exposes only a platform sale. It has no arbitrary line,
// GHS/MoMo, account identifier, payout or member-transfer primitive.
type SettlementLedger interface {
	RecordPlatformSale(context.Context, PlatformSale) (string, error)
}
type PlatformSale struct {
	CommandID, CheckoutRef, ProviderRef string
	Currency                            domain.Currency
	AmountMinor                         int64
}
type Repository interface {
	Create(context.Context, domain.Checkout) error
	Find(context.Context, string) (domain.Checkout, error)
	Save(context.Context, domain.Checkout, uint64, string) error
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
