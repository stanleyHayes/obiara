// Package application governs market packs (E16-S06): drafting, four-eyes
// publishing, retirement, and an immutable configuration change audit
// (plan §8 configuration_changes; §4: all privileged actions auditable).
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/marketpack/domain"
)

var ErrPackNotFound = errors.New("market pack not found")

// PackRepository persists market packs.
type PackRepository interface {
	Create(context.Context, domain.MarketPack) error
	FindByID(context.Context, string) (domain.MarketPack, error)
	Update(context.Context, domain.MarketPack) error
	ListPublished(context.Context) ([]domain.MarketPack, error)
}

// ConfigAudit is the immutable configuration change log.
type ConfigAudit interface {
	Append(ctx context.Context, actorID, action, packID string, at time.Time) error
}

// MarketPackService governs packs.
type MarketPackService struct {
	packs PackRepository
	audit ConfigAudit
	now   func() time.Time
	newID func() string
}

func NewMarketPackService(packs PackRepository, audit ConfigAudit, now func() time.Time, newID func() string) MarketPackService {
	return MarketPackService{packs: packs, audit: audit, now: now, newID: newID}
}

// Draft creates a pack draft and audits the proposal.
func (service MarketPackService) Draft(ctx context.Context, market domain.Market, terminologyRef string, features map[string]bool, proposerID string) (domain.MarketPack, error) {
	pack, err := domain.NewPack(service.newID(), market, terminologyRef, features, proposerID, service.now())
	if err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.packs.Create(ctx, pack); err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.audit.Append(ctx, proposerID, "marketpack.draft", pack.ID(), service.now().UTC()); err != nil {
		return domain.MarketPack{}, err
	}
	return pack, nil
}

// Publish releases a draft under four-eyes approval.
func (service MarketPackService) Publish(ctx context.Context, packID, approverID string) (domain.MarketPack, error) {
	pack, err := service.packs.FindByID(ctx, packID)
	if err != nil {
		return domain.MarketPack{}, err
	}
	if err := pack.Publish(approverID, service.now()); err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.packs.Update(ctx, pack); err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.audit.Append(ctx, approverID, "marketpack.publish", pack.ID(), service.now().UTC()); err != nil {
		return domain.MarketPack{}, err
	}
	return pack, nil
}

// Retire withdraws a published pack.
func (service MarketPackService) Retire(ctx context.Context, packID, actorID string) (domain.MarketPack, error) {
	pack, err := service.packs.FindByID(ctx, packID)
	if err != nil {
		return domain.MarketPack{}, err
	}
	if err := pack.Retire(actorID); err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.packs.Update(ctx, pack); err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.audit.Append(ctx, actorID, "marketpack.retire", pack.ID(), service.now().UTC()); err != nil {
		return domain.MarketPack{}, err
	}
	return pack, nil
}

// Published lists live packs for runtime consumers.
func (service MarketPackService) Published(ctx context.Context) ([]domain.MarketPack, error) {
	return service.packs.ListPublished(ctx)
}
