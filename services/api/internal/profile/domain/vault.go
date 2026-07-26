package domain

import (
	"errors"
	"time"
)

// Vault holds the member's photos. Everything in the vault is veiled by
// default: strangers see only the veil; accepted viewers (sprouted seeds,
// E06) see the clear photo. Until acceptance exists, only the owner sees
// clear items — the veil is the default, not an exception (Doc 06 S-08).
const VaultCapacity = 12

var (
	ErrVaultItemInvalid = errors.New("invalid vault item")
	ErrVaultFull        = errors.New("photo vault is full")
	ErrVaultItemMissing = errors.New("vault item not found")
)

// VaultItem references one media asset held in the vault. Bytes and signed
// access stay in the media context; the vault stores ordering only.
type VaultItem struct {
	id        string
	memberID  string
	assetID   string
	position  int
	createdAt time.Time
}

func NewVaultItem(id, memberID, assetID string, position int, now time.Time) (VaultItem, error) {
	if id == "" || memberID == "" || assetID == "" || position < 0 || position >= VaultCapacity {
		return VaultItem{}, ErrVaultItemInvalid
	}
	return VaultItem{id: id, memberID: memberID, assetID: assetID, position: position, createdAt: now.UTC()}, nil
}

func (item VaultItem) ID() string           { return item.id }
func (item VaultItem) MemberID() string     { return item.memberID }
func (item VaultItem) AssetID() string      { return item.assetID }
func (item VaultItem) Position() int        { return item.position }
func (item VaultItem) CreatedAt() time.Time { return item.createdAt }

// VaultItemView is one vault entry as presented to a viewer.
type VaultItemView struct {
	AssetID  string
	Position int
	Veiled   bool
}

// Veil decides whether an item renders veiled to a viewer. Server-side and
// never client-overridable. Acceptance-based unveiling (sprouted seeds)
// wires in with the seed economy (E06); until then only the owner sees
// clear items.
func Veil(item VaultItem, viewerID string) VaultItemView {
	return VaultItemView{
		AssetID:  item.assetID,
		Position: item.position,
		Veiled:   viewerID != item.memberID,
	}
}
