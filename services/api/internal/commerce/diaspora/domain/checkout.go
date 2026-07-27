package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid diaspora checkout")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{2,127}$`)

const MaxMinor int64 = 100_000_000

type Currency string

const (
	CurrencyGBP Currency = "GBP"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

type Status string

const (
	AwaitingProvider  Status = "awaiting_provider"
	ProviderConfirmed Status = "provider_confirmed"
	Accounted         Status = "accounted"
	Failed            Status = "failed"
)

type Quote struct {
	SKUKey      string
	Version     uint64
	Currency    Currency
	AmountMinor int64
	ValidUntil  time.Time
}
type Event struct {
	Sequence                   uint64
	Kind, CommandID, Reference string
	At                         time.Time
}
type State struct {
	ID, MemberKey                      string
	Quote                              Quote
	Status                             Status
	RequestRef, ProviderRef, LedgerRef string
	Revision                           uint64
	Events                             []Event
	AppliedIDs                         []string
}
type Checkout struct{ state State }

func Create(id, member string, quote Quote, requestRef, command string, at time.Time) (Checkout, error) {
	if !opaque.MatchString(id) || !opaque.MatchString(member) || !validQuote(quote, at) || !opaque.MatchString(requestRef) || !token.MatchString(command) || at.IsZero() {
		return Checkout{}, ErrInvalid
	}
	s := State{ID: id, MemberKey: member, Quote: quote, Status: AwaitingProvider, RequestRef: requestRef, Revision: 1, AppliedIDs: []string{command}}
	s.Events = []Event{{1, "provider_request_prepared", command, requestRef, at.UTC()}}
	return Checkout{s}, nil
}
func (c Checkout) Confirm(providerRef, command string, success bool, at time.Time) (Checkout, error) {
	if c.state.Status != AwaitingProvider || !opaque.MatchString(providerRef) || !token.MatchString(command) || at.IsZero() || slices.Contains(c.state.AppliedIDs, command) {
		return Checkout{}, ErrInvalid
	}
	n := c.State()
	n.Revision++
	n.ProviderRef = providerRef
	n.Status = ProviderConfirmed
	kind := "provider_confirmed"
	if !success {
		n.Status = Failed
		kind = "provider_failed"
	}
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Events = append(n.Events, Event{n.Revision, kind, command, providerRef, at.UTC()})
	return Checkout{n}, nil
}
func (c Checkout) RecordLedger(ledgerRef, command string, at time.Time) (Checkout, error) {
	if c.state.Status != ProviderConfirmed || !opaque.MatchString(ledgerRef) || !token.MatchString(command) || at.IsZero() || slices.Contains(c.state.AppliedIDs, command) {
		return Checkout{}, ErrInvalid
	}
	n := c.State()
	n.Revision++
	n.Status = Accounted
	n.LedgerRef = ledgerRef
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Events = append(n.Events, Event{n.Revision, "platform_sale_accounted", command, ledgerRef, at.UTC()})
	return Checkout{n}, nil
}
func Rehydrate(s State) (Checkout, error) {
	s = clone(s)
	if !validState(s) {
		return Checkout{}, ErrInvalid
	}
	return Checkout{s}, nil
}
func (c Checkout) State() State              { return clone(c.state) }
func (c Checkout) ID() string                { return c.state.ID }
func (c Checkout) Revision() uint64          { return c.state.Revision }
func ValidOpaqueReference(value string) bool { return opaque.MatchString(value) }
func validCurrency(c Currency) bool          { return c == CurrencyGBP || c == CurrencyUSD || c == CurrencyEUR }
func validQuote(q Quote, at time.Time) bool {
	return opaque.MatchString(q.SKUKey) && q.Version > 0 && validCurrency(q.Currency) && q.AmountMinor > 0 && q.AmountMinor <= MaxMinor && q.ValidUntil.After(at)
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.MemberKey) || !opaque.MatchString(s.Quote.SKUKey) || s.Quote.Version == 0 || !validCurrency(s.Quote.Currency) || s.Quote.AmountMinor <= 0 || s.Quote.AmountMinor > MaxMinor || s.Quote.ValidUntil.IsZero() || !opaque.MatchString(s.RequestRef) || s.Revision == 0 || len(s.Events) != int(s.Revision) || len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	for i, e := range s.Events {
		if e.Sequence != uint64(i+1) || e.CommandID != s.AppliedIDs[i] || !token.MatchString(e.CommandID) || e.At.IsZero() || (i > 0 && slices.Contains(s.AppliedIDs[:i], e.CommandID)) {
			return false
		}
	}
	if s.Events[0].Kind != "provider_request_prepared" || s.Events[0].Reference != s.RequestRef {
		return false
	}
	switch s.Status {
	case AwaitingProvider:
		return s.Revision == 1 && s.ProviderRef == "" && s.LedgerRef == ""
	case ProviderConfirmed:
		return s.Revision == 2 && s.Events[1].Kind == "provider_confirmed" && s.Events[1].Reference == s.ProviderRef && opaque.MatchString(s.ProviderRef) && s.LedgerRef == ""
	case Accounted:
		return s.Revision == 3 && s.Events[1].Kind == "provider_confirmed" && s.Events[1].Reference == s.ProviderRef && s.Events[2].Kind == "platform_sale_accounted" && s.Events[2].Reference == s.LedgerRef && opaque.MatchString(s.ProviderRef) && opaque.MatchString(s.LedgerRef)
	case Failed:
		return s.Revision == 2 && s.Events[1].Kind == "provider_failed" && s.Events[1].Reference == s.ProviderRef && opaque.MatchString(s.ProviderRef) && s.LedgerRef == ""
	}
	return false
}
func clone(s State) State {
	s.Events = append([]Event(nil), s.Events...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	return s
}
