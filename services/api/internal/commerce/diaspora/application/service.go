package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/diaspora/domain"
)

var (
	ErrInvalid     = errors.New("invalid diaspora payment request")
	ErrUnavailable = errors.New("diaspora payment unavailable")
	ErrNotFound    = errors.New("diaspora checkout not found")
	ErrConflict    = errors.New("diaspora checkout conflict")
	ErrApplied     = errors.New("diaspora command applied")
)

type Service struct {
	authority Authority
	catalog   Catalog
	verifier  ConfirmationVerifier
	ledger    SettlementLedger
	repo      Repository
	ids       IDSource
	clock     Clock
}

func New(a Authority, c Catalog, v ConfirmationVerifier, l SettlementLedger, r Repository, i IDSource, k Clock) Service {
	return Service{a, c, v, l, r, i, k}
}

type Prepared struct {
	Checkout    domain.Checkout
	RequestRef  string
	Currency    domain.Currency
	AmountMinor int64
}

func (s Service) Prepare(ctx context.Context, member, sku string, currency domain.Currency, command string) (Prepared, error) {
	if s.authority.RequireDiasporaCheckout(ctx, member) != nil {
		return Prepared{}, ErrUnavailable
	}
	quote, e := s.catalog.CurrentDiasporaQuote(ctx, sku, currency)
	if e != nil {
		return Prepared{}, ErrUnavailable
	}
	requestRef := s.ids.NewID()
	checkout, e := domain.Create(s.ids.NewID(), member, quote, requestRef, command, s.clock.Now())
	if e != nil {
		return Prepared{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, checkout); e != nil {
		return Prepared{}, e
	}
	return Prepared{checkout, requestRef, quote.Currency, quote.AmountMinor}, nil
}
func (s Service) Confirm(ctx context.Context, confirmation ProviderConfirmation, command string) (domain.Checkout, error) {
	checkout, e := s.repo.Find(ctx, confirmation.CheckoutID)
	if e != nil {
		return domain.Checkout{}, e
	}
	state := checkout.State()
	if confirmation.RequestRef != state.RequestRef || confirmation.Currency != state.Quote.Currency || confirmation.AmountMinor != state.Quote.AmountMinor || confirmation.OccurredAt.IsZero() || !domain.ValidOpaqueReference(confirmation.Evidence) {
		return domain.Checkout{}, ErrInvalid
	}
	if e = s.verifier.Verify(ctx, confirmation); e != nil {
		return domain.Checkout{}, ErrUnavailable
	}
	next, e := checkout.Confirm(confirmation.ProviderRef, command, confirmation.Succeeded, s.clock.Now())
	if e != nil {
		return domain.Checkout{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, next, checkout.Revision(), command); e != nil {
		return domain.Checkout{}, e
	}
	return next, nil
}
func (s Service) Account(ctx context.Context, id, command string) (domain.Checkout, error) {
	checkout, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Checkout{}, e
	}
	state := checkout.State()
	if state.Status != domain.ProviderConfirmed {
		return domain.Checkout{}, ErrInvalid
	}
	quote, e := s.catalog.CurrentDiasporaQuote(ctx, state.Quote.SKUKey, state.Quote.Currency)
	if e != nil || quote.Version != state.Quote.Version || quote.AmountMinor != state.Quote.AmountMinor {
		return domain.Checkout{}, ErrUnavailable
	}
	ledgerRef, e := s.ledger.RecordPlatformSale(ctx, PlatformSale{command, state.ID, state.ProviderRef, state.Quote.Currency, state.Quote.AmountMinor})
	if e != nil {
		return domain.Checkout{}, ErrUnavailable
	}
	next, e := checkout.RecordLedger(ledgerRef, command, s.clock.Now())
	if e != nil {
		return domain.Checkout{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, next, checkout.Revision(), command); e != nil {
		return domain.Checkout{}, e
	}
	return next, nil
}
