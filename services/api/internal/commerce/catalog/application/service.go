package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/catalog/domain"
	"time"
)

var (
	ErrInvalid     = errors.New("invalid catalog request")
	ErrNotFound    = errors.New("catalog sku not found")
	ErrConflict    = errors.New("catalog conflict")
	ErrApplied     = errors.New("catalog command already applied")
	ErrUnavailable = errors.New("catalog sku unavailable")
)

type Service struct {
	repo      Repository
	authority Authority
	keyer     Keyer
	ids       IDSource
	now       func() time.Time
}

func NewService(r Repository, a Authority, k Keyer, i IDSource, n func() time.Time) Service {
	return Service{r, a, k, i, n}
}
func (s Service) ready() bool {
	return s.repo != nil && s.authority != nil && s.keyer != nil && s.ids != nil && s.now != nil
}

type CreateCommand struct {
	Actor, SKUKey, Title, CommandID string
	Kind                            domain.Kind
	Price                           domain.Price
}

func (s Service) Create(ctx context.Context, q CreateCommand) (domain.SKU, error) {
	if !s.ready() {
		return domain.SKU{}, ErrInvalid
	}
	if e := s.authority.RequireCatalogEditor(ctx, q.Actor); e != nil {
		return domain.SKU{}, e
	}
	title, e := s.keyer.Key("catalog-title", q.Title)
	if e != nil {
		return domain.SKU{}, e
	}
	x, e := domain.Create(s.ids.NewID(), q.SKUKey, title, 1, q.Kind, q.Price, domain.Command{ID: q.CommandID, At: s.now().UTC()})
	if e != nil {
		return domain.SKU{}, e
	}
	if e = s.repo.Create(ctx, x); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return x, e
}

type NextCommand struct {
	Actor, SKUKey, CommandID string
	Price                    domain.Price
}

func (s Service) CreateNext(ctx context.Context, q NextCommand) (domain.SKU, error) {
	if !s.ready() {
		return domain.SKU{}, ErrInvalid
	}
	if e := s.authority.RequireCatalogEditor(ctx, q.Actor); e != nil {
		return domain.SKU{}, e
	}
	prior, e := s.repo.FindLatest(ctx, q.SKUKey)
	if e != nil {
		return domain.SKU{}, e
	}
	x, e := domain.NextVersion(prior, s.ids.NewID(), q.Price, domain.Command{ID: q.CommandID, At: s.now().UTC()})
	if e != nil {
		return domain.SKU{}, e
	}
	if e = s.repo.Create(ctx, x); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return x, e
}

type Change struct {
	Actor, ID, CommandID      string
	Version, ExpectedRevision uint64
}

func (s Service) Publish(ctx context.Context, q Change) (domain.SKU, error) {
	return s.mutate(ctx, q, func(x domain.SKU, c domain.Command) (domain.SKU, error) { return x.Publish(c) })
}
func (s Service) Retire(ctx context.Context, q Change) (domain.SKU, error) {
	return s.mutate(ctx, q, func(x domain.SKU, c domain.Command) (domain.SKU, error) { return x.Retire(c) })
}
func (s Service) mutate(ctx context.Context, q Change, fn func(domain.SKU, domain.Command) (domain.SKU, error)) (domain.SKU, error) {
	if !s.ready() {
		return domain.SKU{}, ErrInvalid
	}
	if e := s.authority.RequireCatalogEditor(ctx, q.Actor); e != nil {
		return domain.SKU{}, e
	}
	x, e := s.repo.Find(ctx, q.ID, q.Version)
	if e != nil {
		return domain.SKU{}, e
	}
	next, e := fn(x, domain.Command{ID: q.CommandID, ExpectedRevision: q.ExpectedRevision, At: s.now().UTC()})
	if e != nil {
		return domain.SKU{}, e
	}
	if e = s.repo.Append(ctx, next, q.ExpectedRevision, q.CommandID); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return next, e
}
func (s Service) ReadPublished(ctx context.Context, sku string, version uint64) (domain.SKU, error) {
	if !s.ready() {
		return domain.SKU{}, ErrInvalid
	}
	x, e := s.repo.Find(ctx, sku, version)
	if e != nil {
		return domain.SKU{}, e
	}
	if x.Status() != domain.StatusPublished {
		return domain.SKU{}, ErrUnavailable
	}
	return x, nil
}
