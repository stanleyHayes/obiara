package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

var (
	ErrDoorwayQuestionMissing = errors.New("no doorway question set")
	ErrVaultItemConflict      = errors.New("vault position already taken")
)

// DoorwayRepository persists the one doorway question per member.
type DoorwayRepository interface {
	Upsert(context.Context, domain.DoorwayQuestion) error
	FindByMember(context.Context, string) (domain.DoorwayQuestion, error)
}

// VaultRepository persists vault item metadata.
type VaultRepository interface {
	Add(context.Context, domain.VaultItem) error
	Remove(context.Context, string) error
	ListByMember(context.Context, string) ([]domain.VaultItem, error)
	CountByMember(context.Context, string) (int, error)
}

// DoorwayService manages the doorway question.
type DoorwayService struct {
	questions DoorwayRepository
	now       func() time.Time
}

func NewDoorwayService(questions DoorwayRepository, now func() time.Time) DoorwayService {
	return DoorwayService{questions: questions, now: now}
}

// Set validates and replaces the member's doorway question.
func (service DoorwayService) Set(ctx context.Context, memberID, text string, custom bool) (domain.DoorwayQuestion, error) {
	question, err := domain.NewDoorwayQuestion(memberID, text, custom, service.now())
	if err != nil {
		return domain.DoorwayQuestion{}, err
	}
	if err := service.questions.Upsert(ctx, question); err != nil {
		return domain.DoorwayQuestion{}, err
	}
	return question, nil
}

// Get returns the member's doorway question for the sower-facing preview.
func (service DoorwayService) Get(ctx context.Context, memberID string) (domain.DoorwayQuestion, error) {
	return service.questions.FindByMember(ctx, memberID)
}

// VaultService manages vault items and veil presentation.
type VaultService struct {
	vault VaultRepository
	now   func() time.Time
	newID func() string
}

func NewVaultService(vault VaultRepository, now func() time.Time, newID func() string) VaultService {
	return VaultService{vault: vault, now: now, newID: newID}
}

// Add places an asset reference in the vault at a free position.
func (service VaultService) Add(ctx context.Context, memberID, assetID string, position int) (domain.VaultItem, error) {
	count, err := service.vault.CountByMember(ctx, memberID)
	if err != nil {
		return domain.VaultItem{}, err
	}
	if count >= domain.VaultCapacity {
		return domain.VaultItem{}, domain.ErrVaultFull
	}
	item, err := domain.NewVaultItem(service.newID(), memberID, assetID, position, service.now())
	if err != nil {
		return domain.VaultItem{}, err
	}
	if err := service.vault.Add(ctx, item); err != nil {
		return domain.VaultItem{}, err
	}
	return item, nil
}

// Remove deletes one item owned by the member.
func (service VaultService) Remove(ctx context.Context, itemID string) error {
	return service.vault.Remove(ctx, itemID)
}

// ViewFor renders the vault for a viewer with server-side veiling.
func (service VaultService) ViewFor(ctx context.Context, ownerID, viewerID string) ([]domain.VaultItemView, error) {
	items, err := service.vault.ListByMember(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	views := make([]domain.VaultItemView, 0, len(items))
	for _, item := range items {
		views = append(views, domain.Veil(item, viewerID))
	}
	return views, nil
}
