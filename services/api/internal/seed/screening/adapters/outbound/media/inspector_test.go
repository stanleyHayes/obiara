package media

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

type assetStub struct {
	asset mediadomain.Asset
	err   error
}

func (stub assetStub) FindByID(context.Context, string) (mediadomain.Asset, error) {
	return stub.asset, stub.err
}

func recording(t *testing.T) mediadomain.Asset {
	t.Helper()
	checksum, err := mediadomain.NewChecksum("sha256", strings.Repeat("b", 64))
	if err != nil {
		t.Fatal(err)
	}
	asset, err := mediadomain.NewAsset(mediadomain.NewAssetParams{
		ID: "asset-1", ObjectKey: "objects/asset-1", OwnerID: "member-1",
		ContentType: "audio/ogg", Size: 2048, Checksum: checksum,
		Duration: 45 * time.Second, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return asset
}

func TestInspectReportsShapeAndNeverContent(t *testing.T) {
	metadata, err := NewInspector(assetStub{asset: recording(t)}).Inspect(context.Background(), "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if metadata.MIME != "audio/ogg" || metadata.Bytes != 2048 || metadata.DurationMs != 45_000 {
		t.Fatalf("metadata = %#v", metadata)
	}
}

func TestARecordingNothingCanDescribeGoesToAPerson(t *testing.T) {
	// The policy turns this into ReasonUnsupportedMedia and routes, which is
	// the right outcome: a sow referencing a recording nothing can describe
	// is exactly what somebody should look at.
	if _, err := NewInspector(assetStub{err: errors.New("gone")}).Inspect(context.Background(), "asset-1"); !errors.Is(err, ErrAssetUnavailable) {
		t.Fatal("an unreadable recording was described anyway")
	}
	if _, err := (Inspector{}).Inspect(context.Background(), "asset-1"); !errors.Is(err, ErrAssetUnavailable) {
		t.Fatal("an uncomposed inspector described something")
	}
}
