package invariants

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	intromedia "github.com/stanleyHayes/obiara/services/api/internal/introduction/adapters/outbound/media"
	introapplication "github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	introdomain "github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
	mediaapplication "github.com/stanleyHayes/obiara/services/api/internal/media/application"
	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

// refusingAssets is a store that cannot be written to and holds nothing —
// which is how a member finds out their recording never got a row, rather
// than being handed a grant that authorizes against nothing.
type refusingAssets struct{}

func (refusingAssets) FindByID(context.Context, string) (mediadomain.Asset, error) {
	return mediadomain.Asset{}, mediadomain.ErrAssetUnavailable
}

func (refusingAssets) Register(context.Context, mediadomain.Asset) error {
	return errors.New("asset store down")
}

// realAccess is the media access service the API composes.
type realAccess struct {
	service mediaapplication.AccessService
}

func (a realAccess) RequestUpload(
	ctx context.Context, request mediaapplication.UploadRequest,
) (mediaapplication.SignedAccess, error) {
	return a.service.RequestUpload(ctx, request)
}

func (a realAccess) RequestRead(
	ctx context.Context, request mediaapplication.ReadRequest,
) (mediaapplication.SignedAccess, error) {
	return a.service.RequestRead(ctx, request)
}

// registered is the asset store as it behaves once something writes to it.
type registered struct{ asset mediadomain.Asset }

func (r *registered) Register(_ context.Context, asset mediadomain.Asset) error {
	r.asset = asset
	return nil
}

func (r *registered) FindByID(context.Context, string) (mediadomain.Asset, error) {
	if r.asset.ID() == "" {
		return mediadomain.Asset{}, mediadomain.ErrAssetUnavailable
	}
	return r.asset, nil
}

func TestAMemberCanBeGrantedAnUploadForTheirOwnRecording(t *testing.T) {
	// The floor the whole product stands on. Without an upload grant no
	// member records a Voice of Introduction, so nobody has a voice, so
	// nobody can hear anybody, so no sow is ever possible.
	store := &registered{}
	manager := intromedia.NewManager(
		realAccess{accessService(t, "the-member", entitled(false))},
		store,
		nil,
		"voice.introduction",
		time.Now,
	)
	media, err := introdomain.NewMediaRef(
		"asset-1", "audio/ogg", 240_000, 60*time.Second, strings.Repeat("a", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AuthorizeUpload(context.Background(), "the-member", media); err != nil {
		t.Fatalf("a member could not be granted an upload: %v", err)
	}
	// The row the grant is authorized against, written by the thing that
	// needs it rather than by nobody.
	if store.asset.ID() != "asset-1" || store.asset.OwnerID() != "the-member" {
		t.Fatalf("registered %#v", store.asset)
	}
	if store.asset.ObjectKey() == "" || strings.Contains(store.asset.ObjectKey(), "the-member") {
		// An object key reaches logs, bucket listings and a signed URL's
		// path. One carrying a member id puts them in all three.
		t.Fatalf("object key = %q", store.asset.ObjectKey())
	}
}

func TestAnUploadGrantNeedsAnAssetRowThatSomethingHasToWrite(t *testing.T) {
	// The media context authorizes against a known owner and object key, so
	// the row has to exist before the grant. When it cannot be written, the
	// grant is refused rather than issued against nothing — which is what
	// used to happen every time, because nothing wrote one at all.
	manager := intromedia.NewManager(
		realAccess{accessService(t, "the-member", entitled(false))},
		refusingAssets{},
		nil,
		"voice.introduction",
		time.Now,
	)
	media, err := introdomain.NewMediaRef(
		"asset-1", "audio/ogg", 240_000, 60*time.Second, strings.Repeat("a", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.AuthorizeUpload(
		context.Background(), "the-member", media,
	); !errors.Is(err, introapplication.ErrDependencyUnavailable) {
		t.Fatalf("err = %v, want ErrDependencyUnavailable", err)
	}
}
