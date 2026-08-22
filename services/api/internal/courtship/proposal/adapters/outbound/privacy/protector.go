package privacy

import "context"

// Protector turns a caller-supplied detail reference into an opaque key.
//
// A proposal's detail — where to meet, what to talk about — is the most
// sensitive text in a courtship, and it must never be readable from the
// proposal store. The domain only ever compares detail references for
// equality, so keying them under a dedicated namespace preserves every
// behaviour the aggregate needs while leaving nothing legible at rest.
type Protector struct{ keyer *Keyer }

func NewProtector(keyer *Keyer) *Protector { return &Protector{keyer: keyer} }

func (protector *Protector) Protect(_ context.Context, detail string) (string, error) {
	return protector.keyer.Key("detail", detail)
}
