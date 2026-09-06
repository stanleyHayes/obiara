package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/momo/domain"
)

var (
	ErrInvalid     = errors.New("invalid momo request")
	ErrNotFound    = errors.New("momo intent not found")
	ErrConflict    = errors.New("momo intent conflict")
	ErrApplied     = errors.New("momo command already applied")
	ErrUnavailable = errors.New("momo unavailable")
)

type Service struct {
	repo     Repository
	provider Provider
	ids      IDSource
	clock    Clock
	secret   []byte
}

func New(r Repository, p Provider, ids IDSource, c Clock, secret []byte) Service {
	return Service{r, p, ids, c, append([]byte(nil), secret...)}
}
func PhoneRef(secret []byte, phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if len(secret) < 32 || len(phone) < 8 || len(phone) > 20 {
		return "", ErrInvalid
	}
	for _, r := range phone {
		if (r < '0' || r > '9') && r != '+' {
			return "", ErrInvalid
		}
	}
	m := hmac.New(sha256.New, secret)
	m.Write([]byte("momo-phone:v1:" + phone))
	return hex.EncodeToString(m.Sum(nil)), nil
}
func (s Service) Create(ctx context.Context, memberKey, phoneRef string, amount uint64, command string) (domain.Intent, error) {
	i, e := domain.Create(s.ids.NewID(), memberKey, phoneRef, amount, command, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, i); e != nil {
		return domain.Intent{}, e
	}
	return i, nil
}

// Payer is what a processor needs to reach somebody, supplied by the caller
// at the moment of collection rather than read back from the intent.
//
// The intent stores an HMAC of the phone, and a digest cannot be dialled. So
// the raw number comes in here and is checked against what was stored: the
// caller cannot substitute somebody else's number, and no raw number is ever
// written down. It is the same shape as the mutual water's counterpart check
// (agent_plan.md §62), for the same reason.
type Payer struct {
	Phone, Email, Network string
}

func (s Service) Confirm(ctx context.Context, id, command string, payer Payer) (domain.Intent, error) {
	i, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Intent{}, e
	}
	// The number given has to be the number this intent was opened for.
	given, e := PhoneRef(s.secret, payer.Phone)
	if e != nil || given != i.State().PhoneRef {
		return domain.Intent{}, ErrInvalid
	}
	n, e := i.Confirm(command, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, i.Revision(), command); e != nil {
		return domain.Intent{}, e
	}
	// The intent's own id is the reference the processor is given, so the
	// webhook that comes back names a collection already on record and needs
	// no second lookup table. It is opaque, so a processor learns nothing
	// from it.
	ref := n.State().ID
	r, e := s.provider.RequestCollection(ctx, ProviderRequest{
		RequestRef: ref, Phone: payer.Phone, Email: payer.Email, Network: payer.Network,
		AmountPesewas: n.State().AmountPesewas, Currency: "GHS",
	})
	if e != nil || r != ref {
		return domain.Intent{}, ErrUnavailable
	}
	requested, e := n.MarkRequested(command+":provider", ref, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, requested, n.Revision(), command+":provider"); e != nil {
		return domain.Intent{}, e
	}
	return requested, nil
}

// Settle applies an outcome the caller has already authenticated.
//
// There is deliberately no signature check here. The processor signs the exact
// bytes of its webhook with an HMAC in a header, so the only place that can be
// verified is where those bytes still exist — the transport. Re-checking a
// signature over fields already parsed out of the body would be hashing
// something the processor never sent, and would break the first time it added
// a field or changed key order.
//
// What this still owns is replay: the callback id goes into the intent's
// applied commands, so a processor retrying — which they all do — settles once.
//
// The caller MUST have verified the webhook before reaching this. The only
// caller is the HTTP handler that does exactly that.
func (s Service) Settle(
	ctx context.Context, callbackID, intentID, providerRef string, success bool,
) (domain.Intent, error) {
	i, e := s.repo.Find(ctx, intentID)
	if e != nil {
		return domain.Intent{}, e
	}
	n, e := i.ApplyProvider(callbackID, providerRef, success, s.clock.Now())
	if e != nil {
		return domain.Intent{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, i.Revision(), callbackID); e != nil {
		return domain.Intent{}, e
	}
	return n, nil
}

// Find reads an intent back, so a caller can check that what a processor says
// arrived is what was actually asked for.
func (s Service) Find(ctx context.Context, id string) (domain.Intent, error) {
	return s.repo.Find(ctx, id)
}
