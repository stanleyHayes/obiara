package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/ledger/domain"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid ledger request")
	ErrNotFound = errors.New("ledger posting not found")
	ErrConflict = errors.New("ledger conflict")
	ErrApplied  = errors.New("ledger posting already applied")
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

type Line struct {
	AccountID string
	Class     domain.AccountClass
	Side      domain.Side
	Minor     int64
}
type PostCommand struct {
	Actor, CommandID, ReferenceID string
	Purpose                       domain.Purpose
	Currency                      domain.Currency
	Lines                         []Line
}

func (s Service) Post(ctx context.Context, q PostCommand) (domain.Posting, error) {
	if s.repo == nil || s.authority == nil || s.keyer == nil || s.ids == nil || s.now == nil {
		return domain.Posting{}, ErrInvalid
	}
	if e := s.authority.RequirePoster(ctx, q.Actor, q.Purpose); e != nil {
		return domain.Posting{}, e
	}
	reference, e := s.keyer.Key("ledger-reference:"+string(q.Purpose), q.ReferenceID)
	if e != nil {
		return domain.Posting{}, e
	}
	lines := make([]domain.Line, 0, len(q.Lines))
	for _, x := range q.Lines {
		account, xerr := s.keyer.Key("ledger-account:"+string(x.Class), x.AccountID)
		if xerr != nil {
			return domain.Posting{}, xerr
		}
		lines = append(lines, domain.Line{AccountKey: account, Class: x.Class, Side: x.Side, Minor: x.Minor})
	}
	p, e := domain.NewPosting(s.ids.NewID(), q.CommandID, reference, q.Purpose, q.Currency, lines, s.now().UTC())
	if e != nil {
		return domain.Posting{}, e
	}
	if e = s.repo.Create(ctx, p); errors.Is(e, ErrApplied) {
		existing, x := s.repo.FindByCommand(ctx, q.CommandID)
		if x != nil {
			return domain.Posting{}, x
		}
		if existing.Fingerprint() != p.Fingerprint() {
			return domain.Posting{}, ErrConflict
		}
		return existing, nil
	}
	return p, e
}

type BalanceRequest struct {
	Actor, AccountID string
	Class            domain.AccountClass
	Currency         domain.Currency
}

func (s Service) Balance(ctx context.Context, q BalanceRequest) (int64, error) {
	if s.repo == nil || s.authority == nil || s.keyer == nil {
		return 0, ErrInvalid
	}
	if e := s.authority.RequireBalanceReader(ctx, q.Actor, q.AccountID); e != nil {
		return 0, e
	}
	account, e := s.keyer.Key("ledger-account:"+string(q.Class), q.AccountID)
	if e != nil {
		return 0, e
	}
	lines, e := s.repo.ListLines(ctx, account, q.Currency)
	if e != nil {
		return 0, e
	}
	return domain.RecomputeBalance(account, q.Class, q.Currency, lines)
}
