package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Repository interface {
	Create(context.Context, domain.Intent) error
	Find(context.Context, string) (domain.Intent, error)
	Save(context.Context, domain.Intent, uint64, string) error
}
type Provider interface {
	RequestCollection(context.Context, ProviderRequest) (string, error)
}

// ProviderRequest is what a processor needs in order to prompt somebody.
//
// Phone is the real number, not the stored digest. This is the whole reason
// Confirm now takes it: the intent stores only an HMAC of the number, which is
// right for a row that outlives the payment and useless for dialling. Handing
// the digest to a provider — which is what this did — asks a payment
// processor to charge a 64-character hex string (agent_plan.md §75).
//
// Email is here because a processor needs somewhere to send a receipt and
// somewhere to attach a dispute. It is a deliberate disclosure to the payment
// processor and to nobody else.
type ProviderRequest struct {
	RequestRef    string
	Phone         string
	Email         string
	Network       string
	AmountPesewas uint64
	Currency      string
}
type IDSource interface{ NewID() string }
type Clock interface{ Now() time.Time }
