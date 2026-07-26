package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/lifecycle/domain"
	"time"
)

type Service struct {
	repo  Repository
	keys  Keyer
	holds LegalHold
	now   func() time.Time
}

func NewService(r Repository, k Keyer, h LegalHold, n func() time.Time) Service {
	return Service{r, k, h, n}
}
func (s Service) load(ctx context.Context, pairID, memberID string) (domain.Lifecycle, string, error) {
	pk, e := s.keys.Key("cloth_pair", pairID)
	if e != nil {
		return domain.Lifecycle{}, "", e
	}
	mk, e := s.keys.Key("cloth_member", memberID)
	if e != nil {
		return domain.Lifecycle{}, "", e
	}
	v, e := s.repo.Find(ctx, pk)
	return v, mk, e
}
func (s Service) Archive(ctx context.Context, pairID, memberID, commandID, archiveRef string, expected uint64) (domain.Lifecycle, error) {
	v, m, e := s.load(ctx, pairID, memberID)
	if e != nil {
		return domain.Lifecycle{}, e
	}
	n, e := v.Archive(domain.Command{ID: commandID, ActorKey: m, ArchiveRef: archiveRef, ExpectedRevision: expected, At: s.now()})
	if e != nil {
		return domain.Lifecycle{}, e
	}
	if n.Revision() != v.Revision() {
		e = s.repo.Save(ctx, n, v.Revision(), commandID)
	}
	return n, e
}
func (s Service) Export(ctx context.Context, pairID, memberID, archiveRef string) (domain.Export, error) {
	v, m, e := s.load(ctx, pairID, memberID)
	if e != nil {
		return domain.Export{}, e
	}
	return v.Export(m, archiveRef)
}
func (s Service) Delete(ctx context.Context, pairID, memberID, commandID string, expected uint64) (domain.Lifecycle, error) {
	v, m, e := s.load(ctx, pairID, memberID)
	if e != nil {
		return domain.Lifecycle{}, e
	}
	allowed, e := s.holds.DeletionAllowed(ctx, v.ID())
	if e != nil || !allowed {
		return domain.Lifecycle{}, domain.ErrDenied
	}
	receipt, e := s.keys.Key("cloth_receipt", v.ID()+"\x00"+commandID)
	if e != nil {
		return domain.Lifecycle{}, e
	}
	n, e := v.Delete(domain.Command{ID: commandID, ActorKey: m, ReceiptKey: receipt, ExpectedRevision: expected, At: s.now()})
	if e != nil {
		return domain.Lifecycle{}, e
	}
	if n.Revision() != v.Revision() {
		e = s.repo.Save(ctx, n, v.Revision(), commandID)
	}
	return n, e
}
