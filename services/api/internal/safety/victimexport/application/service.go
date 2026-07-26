package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/safety/victimexport/domain"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid victim export request")
	ErrNotFound = errors.New("victim export not found")
	ErrConflict = errors.New("victim export conflict")
	ErrApplied  = errors.New("victim export command already applied")
)

type Service struct {
	repo      Repository
	authority Authority
	allowlist Allowlist
	redactor  Redactor
	keyer     Keyer
	ids       IDSource
	now       func() time.Time
}

func NewService(r Repository, a Authority, l Allowlist, d Redactor, k Keyer, i IDSource, n func() time.Time) Service {
	return Service{r, a, l, d, k, i, n}
}
func (s Service) ready() bool {
	return s.repo != nil && s.authority != nil && s.allowlist != nil && s.redactor != nil && s.keyer != nil && s.ids != nil && s.now != nil
}

type ReferenceRequest struct {
	Kind        domain.ReferenceKind
	ReferenceID string
}
type RequestCommand struct {
	Actor, Member, CommandID string
	Purpose                  domain.Purpose
	References               []ReferenceRequest
}

func (s Service) Request(ctx context.Context, q RequestCommand) (domain.Export, error) {
	if !s.ready() {
		return domain.Export{}, ErrInvalid
	}
	if e := s.authority.RequireMember(ctx, q.Actor, q.Member); e != nil {
		return domain.Export{}, e
	}
	member, e := s.keyer.Key("victim-export-member", q.Member)
	if e != nil {
		return domain.Export{}, e
	}
	refs := make([]domain.Reference, 0, len(q.References))
	for _, x := range q.References {
		if e = s.allowlist.Require(ctx, x.Kind, x.ReferenceID, q.Purpose); e != nil {
			return domain.Export{}, e
		}
		attestation, xerr := s.redactor.RedactReporterAndThirdParties(ctx, x.Kind, x.ReferenceID, q.Purpose)
		if xerr != nil {
			return domain.Export{}, xerr
		}
		refKey, xerr := s.keyer.Key("victim-export-reference:"+string(x.Kind), x.ReferenceID)
		if xerr != nil {
			return domain.Export{}, xerr
		}
		redactionKey, xerr := s.keyer.Key("victim-export-redaction", attestation)
		if xerr != nil {
			return domain.Export{}, xerr
		}
		refs = append(refs, domain.Reference{Kind: x.Kind, RefKey: refKey, RedactionKey: redactionKey})
	}
	out, e := domain.Request(s.ids.NewID(), member, q.Purpose, refs, domain.Command{ID: q.CommandID, At: s.now().UTC()})
	if e != nil {
		return domain.Export{}, e
	}
	if e = s.repo.Create(ctx, out); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return out, e
}

type Change struct {
	Actor, Member, ID, CommandID, Token string
	ExpectedRevision                    uint64
}

func (s Service) Authorize(ctx context.Context, q Change) (domain.Export, error) {
	return s.memberMutate(ctx, q, func(e domain.Export, c domain.Command) (domain.Export, error) {
		token, x := s.keyer.Key("victim-export-token", q.Token)
		if x != nil {
			return domain.Export{}, x
		}
		return e.Authorize(e.References(), token, c)
	})
}
func (s Service) Use(ctx context.Context, q Change) (domain.Export, error) {
	return s.memberMutate(ctx, q, func(e domain.Export, c domain.Command) (domain.Export, error) {
		token, x := s.keyer.Key("victim-export-token", q.Token)
		if x != nil {
			return domain.Export{}, x
		}
		return e.Use(token, c)
	})
}
func (s Service) Revoke(ctx context.Context, q Change) (domain.Export, error) {
	return s.memberMutate(ctx, q, func(e domain.Export, c domain.Command) (domain.Export, error) { return e.Revoke(c) })
}
func (s Service) memberMutate(ctx context.Context, q Change, fn func(domain.Export, domain.Command) (domain.Export, error)) (domain.Export, error) {
	if !s.ready() {
		return domain.Export{}, ErrInvalid
	}
	if e := s.authority.RequireMember(ctx, q.Actor, q.Member); e != nil {
		return domain.Export{}, e
	}
	out, e := s.repo.Find(ctx, q.ID)
	if e != nil {
		return domain.Export{}, e
	}
	member, e := s.keyer.Key("victim-export-member", q.Member)
	if e != nil || out.MemberKey() != member {
		return domain.Export{}, ErrInvalid
	}
	next, e := fn(out, domain.Command{ID: q.CommandID, ExpectedRevision: q.ExpectedRevision, At: s.now().UTC()})
	if e != nil {
		return domain.Export{}, e
	}
	if e = s.repo.Append(ctx, next, q.ExpectedRevision, q.CommandID); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return next, e
}
