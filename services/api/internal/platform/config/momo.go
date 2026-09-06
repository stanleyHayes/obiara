package config

import "strings"

// MobileMoney addresses the collection provider that takes a member's money.
//
// Every field is read from the environment with no default, for the same
// reason the object store has none: a default endpoint is somebody else's
// endpoint, and a default credential is a credential in source control. When
// this is absent the payment context is simply not composed and nothing in
// the product can be bought — which is honest, and is the state the product
// has been in all along (agent_plan.md §72).
type MobileMoney struct {
	// BaseURL is the provider's API root. MTN's sandbox and production hosts
	// differ, so it is configured rather than chosen here.
	BaseURL string
	// SubscriptionKey identifies the API product; APIUser and APIKey are the
	// credential pair issued for it.
	SubscriptionKey string
	APIUser         string
	APIKey          string
	// TargetEnvironment is "sandbox" or the production environment name the
	// provider assigned. Sending the wrong one silently charges nobody.
	TargetEnvironment string
	// CallbackURL is where the provider reports the outcome. It has to be
	// reachable from outside, which is why it is configured rather than
	// derived from the request.
	CallbackURL string
}

// Configured reports whether enough is present to compose the payment
// context. The credential fields are checked together: half a credential is a
// misconfiguration that should read as "off", not as "on and broken".
func (money MobileMoney) Configured() bool {
	return money.BaseURL != "" && money.SubscriptionKey != "" &&
		money.APIUser != "" && money.APIKey != "" && money.TargetEnvironment != ""
}

func loadMobileMoney(getenv func(string) string) MobileMoney {
	return MobileMoney{
		BaseURL:           strings.TrimRight(strings.TrimSpace(getenv("MOMO_BASE_URL")), "/"),
		SubscriptionKey:   strings.TrimSpace(getenv("MOMO_SUBSCRIPTION_KEY")),
		APIUser:           strings.TrimSpace(getenv("MOMO_API_USER")),
		APIKey:            strings.TrimSpace(getenv("MOMO_API_KEY")),
		TargetEnvironment: strings.TrimSpace(getenv("MOMO_TARGET_ENVIRONMENT")),
		CallbackURL:       strings.TrimSpace(getenv("MOMO_CALLBACK_URL")),
	}
}
