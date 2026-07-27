package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/escrow/domain"
)

var (
	ErrInvalid     = errors.New("invalid escrow request")
	ErrNotFound    = errors.New("escrow not found")
	ErrConflict    = errors.New("escrow conflict")
	ErrApplied     = errors.New("escrow command applied")
	ErrUnavailable = errors.New("escrow unavailable")
)

type Service struct {
	repo   Repository
	ledger Ledger
	ids    IDSource
	clock  Clock
}

func New(r Repository, l Ledger, ids IDSource, c Clock) Service { return Service{r, l, ids, c} }
func (s Service) Fund(ctx context.Context, fundingRef string, amount uint64, terms domain.Terms, command string) (domain.Escrow, error) {
	x, e := domain.Fund(s.ids.NewID(), fundingRef, amount, terms, command, s.clock.Now())
	if e != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if e = s.repo.Create(ctx, x); e != nil {
		return domain.Escrow{}, e
	}
	return x, nil
}
func (s Service) Mutate(ctx context.Context, id, command string, fn func(domain.Escrow) (domain.Escrow, error)) (domain.Escrow, error) {
	x, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Escrow{}, e
	}
	n, e := fn(x)
	if e != nil {
		return domain.Escrow{}, ErrInvalid
	}
	if e = s.repo.Save(ctx, n, x.Revision(), command); e != nil {
		return domain.Escrow{}, e
	}
	return n, nil
}
func (s Service) Settle(ctx context.Context, id, milestone, command string) (domain.Escrow, domain.Statement, error) {
	x, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Escrow{}, domain.Statement{}, e
	}
	n, statement, e := x.Settle(milestone, s.ids.NewID(), command, s.clock.Now())
	if e != nil {
		return domain.Escrow{}, domain.Statement{}, ErrInvalid
	}
	if e = s.ledger.RecordSettlement(ctx, command, statement); e != nil {
		return domain.Escrow{}, domain.Statement{}, ErrUnavailable
	}
	if e = s.repo.Save(ctx, n, x.Revision(), command); e != nil {
		return domain.Escrow{}, domain.Statement{}, e
	}
	return n, statement, nil
}
