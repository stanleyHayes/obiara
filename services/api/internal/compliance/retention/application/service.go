package application

import (
	"context"
	"errors"
	"github.com/stanleyHayes/obiara/services/api/internal/compliance/retention/domain"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid retention request")
	ErrNotFound = errors.New("retention record not found")
	ErrConflict = errors.New("retention conflict")
	ErrApplied  = errors.New("retention command already applied")
)

type Service struct {
	repo      Repository
	policies  PolicyCatalog
	authority Authority
	verifier  ErasureVerifier
	keyer     Keyer
	ids       IDSource
	now       func() time.Time
}

func NewService(r Repository, p PolicyCatalog, a Authority, v ErasureVerifier, k Keyer, i IDSource, n func() time.Time) Service {
	return Service{r, p, a, v, k, i, n}
}
func (s Service) ready() bool {
	return s.repo != nil && s.policies != nil && s.authority != nil && s.verifier != nil && s.keyer != nil && s.ids != nil && s.now != nil
}

type CreateCommand struct {
	Actor, Subject, DataClass, Purpose, CommandID string
	PolicyVersion                                 uint64
}

func (s Service) Create(ctx context.Context, q CreateCommand) (domain.Record, error) {
	if !s.ready() {
		return domain.Record{}, ErrInvalid
	}
	if e := s.authority.RequireSubject(ctx, q.Actor, q.Subject); e != nil {
		return domain.Record{}, e
	}
	p, e := s.policies.Find(ctx, q.DataClass, q.Purpose, q.PolicyVersion)
	if e != nil {
		return domain.Record{}, e
	}
	subject, e := s.keyer.Key("retention-subject", q.Subject)
	if e != nil {
		return domain.Record{}, e
	}
	r, e := domain.Create(s.ids.NewID(), subject, p, domain.Command{ID: q.CommandID, At: s.now().UTC()})
	if e != nil {
		return domain.Record{}, e
	}
	if e = s.repo.Create(ctx, r); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, q.CommandID)
	}
	return r, e
}

type Change struct {
	Actor, Subject, ID, CommandID string
	ExpectedRevision              uint64
	CaseID                        string
}

func (s Service) PlaceHold(ctx context.Context, q Change) (domain.Record, error) {
	if !s.ready() {
		return domain.Record{}, ErrInvalid
	}
	if e := s.authority.RequireLegalHoldOfficer(ctx, q.Actor); e != nil {
		return domain.Record{}, e
	}
	caseKey, e := s.keyer.Key("retention-hold", q.CaseID)
	if e != nil {
		return domain.Record{}, e
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(r domain.Record, c domain.Command) (domain.Record, error) { return r.PlaceHold(caseKey, c) })
}
func (s Service) ReleaseHold(ctx context.Context, q Change) (domain.Record, error) {
	if !s.ready() {
		return domain.Record{}, ErrInvalid
	}
	if e := s.authority.RequireLegalHoldOfficer(ctx, q.Actor); e != nil {
		return domain.Record{}, e
	}
	caseKey, e := s.keyer.Key("retention-hold", q.CaseID)
	if e != nil {
		return domain.Record{}, e
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(r domain.Record, c domain.Command) (domain.Record, error) { return r.ReleaseHold(caseKey, c) })
}
func (s Service) RequestErasure(ctx context.Context, q Change) (domain.Record, error) {
	if !s.ready() {
		return domain.Record{}, ErrInvalid
	}
	if e := s.authority.RequireSubject(ctx, q.Actor, q.Subject); e != nil {
		return domain.Record{}, e
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(r domain.Record, c domain.Command) (domain.Record, error) { return r.RequestErasure(c) })
}
func (s Service) CompleteErasure(ctx context.Context, q Change) (domain.Record, error) {
	if !s.ready() {
		return domain.Record{}, ErrInvalid
	}
	if e := s.authority.RequireSubject(ctx, q.Actor, q.Subject); e != nil {
		return domain.Record{}, e
	}
	verification, e := s.verifier.Verify(ctx, q.Subject, q.ID, q.CommandID)
	if e != nil {
		return domain.Record{}, e
	}
	keyed, e := s.keyer.Key("retention-erasure-proof", verification)
	if e != nil {
		return domain.Record{}, e
	}
	return s.mutate(ctx, q.ID, q.CommandID, q.ExpectedRevision, func(r domain.Record, c domain.Command) (domain.Record, error) { return r.CompleteErasure(keyed, c) })
}
func (s Service) mutate(ctx context.Context, id, command string, revision uint64, fn func(domain.Record, domain.Command) (domain.Record, error)) (domain.Record, error) {
	r, e := s.repo.Find(ctx, id)
	if e != nil {
		return domain.Record{}, e
	}
	next, e := fn(r, domain.Command{ID: command, ExpectedRevision: revision, At: s.now().UTC()})
	if e != nil {
		return domain.Record{}, e
	}
	if e = s.repo.Append(ctx, next, revision, command); errors.Is(e, ErrApplied) {
		return s.repo.FindByCommand(ctx, command)
	}
	return next, e
}
