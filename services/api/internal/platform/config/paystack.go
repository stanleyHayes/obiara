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

// Affiliates is the commercial shape of the referral scheme.
//
// All three are money decisions rather than engineering ones, so none has a
// default that would quietly become policy. Absent means the scheme is not
// composed and nothing accrues, which is the honest state for a scheme nobody
// has set a rate for.
type Affiliates struct {
	// CommissionPesewas is the flat amount one qualified conversion earns.
	// Flat rather than a percentage so an affiliate's statement says nothing
	// about what any individual member paid.
	CommissionPesewas int64
	// MinimumPayoutPesewas is the floor a balance must clear before it can be
	// requested. It exists so the platform is not sending transfer fees to
	// move a cedi.
	MinimumPayoutPesewas int64
	// WithholdingBasisPoints is the tax withheld on commission, in hundredths
	// of a percent — 750 is 7.5%. Basis points because a rate multiplied by
	// money must not be floating point.
	//
	// There is deliberately no default. A wrong withholding rate is a filing
	// problem, not a bug, and guessing one on somebody's behalf would be the
	// worst kind of helpful.
	WithholdingBasisPoints int64
}

// Configured reports whether the affiliate scheme can run. All three, because
// a commission with no withholding rate cannot legally pay out and a scheme
// that accrues but can never pay is a liability that only grows.
func (affiliates Affiliates) Configured() bool {
	return affiliates.CommissionPesewas > 0 &&
		affiliates.WithholdingBasisPoints > 0 && affiliates.WithholdingBasisPoints < 10_000
}

func loadAffiliates(getenv func(string) string) Affiliates {
	return Affiliates{
		CommissionPesewas:      wholeNumber(getenv("AFFILIATE_COMMISSION_PESEWAS")),
		MinimumPayoutPesewas:   wholeNumber(getenv("AFFILIATE_MINIMUM_PAYOUT_PESEWAS")),
		WithholdingBasisPoints: wholeNumber(getenv("AFFILIATE_WITHHOLDING_BASIS_POINTS")),
	}
}

// wholeNumber reads a positive integer, answering zero for anything else.
// Zero is the "not set" value everywhere it is used, so a typo reads as absent
// rather than as some other number.
func wholeNumber(value string) int64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	var total int64
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0
		}
		total = total*10 + int64(character-'0')
		if total > 1<<40 {
			return 0
		}
	}
	return total
}
