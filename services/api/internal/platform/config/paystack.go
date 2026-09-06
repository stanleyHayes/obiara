package config

import "strings"

// Paystack addresses the payment processor that takes a member's money.
//
// The secret key has no default, for the same reason the object store's
// credentials have none: a default credential is a credential in source
// control. When it is absent the payment context is simply not composed and
// nothing in the product can be bought — which is honest, and is the state the
// product was in for its whole life until §73 (agent_plan.md §72).
//
// BaseURL does have a default, and that is not the same kind of thing: it is a
// named vendor's published API host, not somebody else's bucket. It stays
// configurable so tests and a sandbox can point elsewhere.
type Paystack struct {
	BaseURL string
	// SecretKey is used twice and both matter: as the API bearer token, and as
	// the HMAC key Paystack signs its webhooks with. Leaking it lets somebody
	// both spend and forge.
	SecretKey string
	// CallbackURL is where Paystack reports the outcome. It must be reachable
	// from outside, which is why it is configured rather than derived from a
	// request.
	CallbackURL string
}

const defaultPaystackBaseURL = "https://api.paystack.co"

// Configured reports whether enough is present to compose the payment context.
//
// The secret key alone. Without it nothing can be charged and no webhook can
// be verified, and half a configuration should read as "off" rather than as
// "on and broken".
func (paystack Paystack) Configured() bool { return paystack.SecretKey != "" }

// Live reports whether these are production keys. Paystack prefixes test keys
// with sk_test_ and live keys with sk_live_, which is the only warning a
// deployment gets that it is about to move real money.
func (paystack Paystack) Live() bool {
	return strings.HasPrefix(paystack.SecretKey, "sk_live_")
}

func loadPaystack(getenv func(string) string) Paystack {
	baseURL := strings.TrimRight(strings.TrimSpace(getenv("PAYSTACK_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = defaultPaystackBaseURL
	}
	return Paystack{
		BaseURL:     baseURL,
		SecretKey:   strings.TrimSpace(getenv("PAYSTACK_SECRET_KEY")),
		CallbackURL: strings.TrimSpace(getenv("PAYSTACK_CALLBACK_URL")),
	}
}
