package media

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

type fakeAssets struct {
	asset     domain.Asset
	findErr   error
	deleted   []string
	deleteErr error
}

func (f *fakeAssets) FindByID(context.Context, string) (domain.Asset, error) {
	return f.asset, f.findErr
}

func (f *fakeAssets) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return f.deleteErr
}

type fakeObjects struct {
	removed []string
	err     error
}

func (f *fakeObjects) Delete(_ context.Context, key string) error {
	f.removed = append(f.removed, key)
	return f.err
}

func testAsset(t *testing.T, legalHold bool) domain.Asset {
	t.Helper()
	checksum, err := domain.NewChecksum("sha256", strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)
	asset, err := domain.NewAsset(domain.NewAssetParams{
		ID: "intro_asset_1", ObjectKey: "voice/introductions/a.opus",
		OwnerID: "member-1", ContentType: "audio/ogg", Size: 1024,
		Checksum: checksum, Duration: 42 * time.Second, CreatedAt: now,
		Retention: domain.NewRetention(now.Add(180*24*time.Hour), legalHold),
	})
	if err != nil {
		t.Fatal(err)
	}
	return asset
}

func TestErasureRemovesTheBytesBeforeTheRow(t *testing.T) {
	// A row deleted first orphans the object: nothing is left that knows the
	// key, so the bytes stay in the bucket with no record that they exist.
	assets := &fakeAssets{asset: testAsset(t, false)}
	objects := &fakeObjects{}

	if err := NewEraser(assets, objects).Erase(context.Background(), "intro_asset_1"); err != nil {
		t.Fatalf("Erase = %v", err)
	}
	if len(objects.removed) != 1 || objects.removed[0] != "voice/introductions/a.opus" {
		t.Fatalf("objects removed = %v", objects.removed)
	}
	if len(assets.deleted) != 1 {
		t.Fatalf("rows deleted = %v", assets.deleted)
	}
}

func TestAFailedObjectDeleteLeavesTheRowPointingAtIt(t *testing.T) {
	// The row is what the next sweep uses to find the bytes again. Deleting it
	// anyway would make the failure permanent and invisible.
	assets := &fakeAssets{asset: testAsset(t, false)}
	objects := &fakeObjects{err: errors.New("storage refused")}

	if err := NewEraser(assets, objects).Erase(context.Background(), "intro_asset_1"); err == nil {
		t.Fatal("a failed object delete was reported as a successful erasure")
	}
	if len(assets.deleted) != 0 {
		t.Fatal("the row was deleted despite the bytes still being there")
	}
}

func TestALegalHoldIsNeverSwept(t *testing.T) {
	assets := &fakeAssets{asset: testAsset(t, true)}
	objects := &fakeObjects{}

	if err := NewEraser(assets, objects).Erase(context.Background(), "intro_asset_1"); err == nil {
		t.Fatal("a held asset was erased")
	}
	if len(objects.removed) != 0 || len(assets.deleted) != 0 {
		t.Fatal("a held asset was touched")
	}
}

func TestAnAssetThatIsAlreadyGoneIsSuccess(t *testing.T) {
	// Purges are retried; a retry must not fail because the first attempt
	// worked, or the aggregate never leaves purge_pending.
	assets := &fakeAssets{findErr: domain.ErrAssetUnavailable}
	objects := &fakeObjects{}

	if err := NewEraser(assets, objects).Erase(context.Background(), "gone"); err != nil {
		t.Fatalf("Erase on a missing asset = %v, want nil", err)
	}
}

// Stat is what confirming an upload asks. This fake reports a byte for an
// object it still holds, and nothing for one it has removed.
func (f *fakeObjects) Stat(_ context.Context, objectKey string) (int64, error) {
	for _, removed := range f.removed {
		if removed == objectKey {
			return 0, errors.New("gone")
		}
	}
	return 1, nil
}

func TestDeletingARecordingRemovesTheBytesAndNotOnlyTheRow(t *testing.T) {
	// The introduction context asks its remover to Delete when a member
	// withdraws a recording. It was given the asset row repository, whose
	// Delete marks the row and leaves the audio in the bucket — so
	// withdrawing removed the record of the recording and kept the recording.
	assets := &fakeAssets{asset: testAsset(t, false)}
	objects := &fakeObjects{}
	if err := NewEraser(assets, objects).Delete(context.Background(), "asset-1"); err != nil {
		t.Fatal(err)
	}
	if len(objects.removed) != 1 {
		t.Fatal("the bytes are still in the bucket")
	}
	if len(assets.deleted) != 1 {
		t.Fatal("the row was left pointing at bytes that are gone")
	}
}
