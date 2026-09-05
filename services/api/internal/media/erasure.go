package media

import (
	"context"
	"fmt"

	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

// ObjectRemover erases stored bytes. Implemented by the object store adapter.
type ObjectRemover interface {
	Delete(ctx context.Context, objectKey string) error
}

// AssetStore is the metadata side of an erasure.
type AssetStore interface {
	FindByID(context.Context, string) (domain.Asset, error)
	Delete(ctx context.Context, id string) error
}

// Eraser removes an asset's bytes and then its row.
//
// The order is the whole point. Deleting the row first orphans the object:
// nothing is left that knows the key, so the bytes stay in the bucket
// permanently with no record that they are there. Bytes first means a failure
// leaves a row still pointing at them, which the next sweep retries.
//
// This is object deletion, not crypto-erase. Introduction audio is uploaded
// straight to the bucket and is not envelope-encrypted (NFR-301b is open), so
// there is no per-user key to destroy — only the object itself.
type Eraser struct {
	assets  AssetStore
	objects ObjectRemover
}

func NewEraser(assets AssetStore, objects ObjectRemover) Eraser {
	return Eraser{assets: assets, objects: objects}
}

// Erase removes the asset identified by id. An asset that is already gone is
// success: purges are retried, and a retry that fails because the first
// attempt worked would strand the caller forever.
func (eraser Eraser) Erase(ctx context.Context, assetID string) error {
	if eraser.assets == nil || eraser.objects == nil {
		return fmt.Errorf("media erasure is not configured")
	}
	asset, err := eraser.assets.FindByID(ctx, assetID)
	if err != nil {
		return nil
	}
	// A legal hold outranks retention. The runner that calls this is explicit
	// that holds live outside retention automation, and honouring that here
	// means a hold cannot be lost by a sweep that did not check.
	if asset.Retention().LegalHold() {
		return fmt.Errorf("asset %s is under legal hold", assetID)
	}
	if err := eraser.objects.Delete(ctx, asset.ObjectKey()); err != nil {
		return err
	}
	return eraser.assets.Delete(ctx, assetID)
}
