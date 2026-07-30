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
	CreateWithAudit(context.Context, domain.MarketPack, string, string, time.Time) error
	FindByID(context.Context, string) (domain.MarketPack, error)
	UpdateWithAudit(context.Context, domain.MarketPack, string, string, time.Time) error
	ListAll(context.Context, int) ([]domain.MarketPack, error)
	ListPublished(context.Context) ([]domain.MarketPack, error)
}

// MarketPackService governs packs.
type MarketPackService struct {
	packs PackRepository
	now   func() time.Time
	newID func() string
}

func NewMarketPackService(packs PackRepository, now func() time.Time, newID func() string) MarketPackService {
	return MarketPackService{packs: packs, now: now, newID: newID}
}

// Draft creates a pack draft and audits the proposal.
func (service MarketPackService) Draft(ctx context.Context, market domain.Market, terminologyRef string, features map[string]bool, proposerID string) (domain.MarketPack, error) {
	pack, err := domain.NewPack(service.newID(), market, terminologyRef, features, proposerID, service.now())
	if err != nil {
		return domain.MarketPack{}, err
	}
	if err := service.packs.CreateWithAudit(ctx, pack, proposerID, "marketpack.draft", service.now().UTC()); err != nil {
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
	if err := service.packs.UpdateWithAudit(ctx, pack, approverID, "marketpack.publish", service.now().UTC()); err != nil {
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
	if err := service.packs.UpdateWithAudit(ctx, pack, actorID, "marketpack.retire", service.now().UTC()); err != nil {
		return domain.MarketPack{}, err
	}
	return pack, nil
}

// Published lists live packs for runtime consumers.
func (service MarketPackService) Published(ctx context.Context) ([]domain.MarketPack, error) {
	return service.packs.ListPublished(ctx)
}

// All returns the bounded governance register for authenticated operators.
func (service MarketPackService) All(ctx context.Context, limit int) ([]domain.MarketPack, error) {
	return service.packs.ListAll(ctx, limit)
}
