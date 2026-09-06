// Package invariants holds the tests that only fail when real pieces are put
// together. The media context, its policy and its callers each pass their own
// tests; what they do to each other is what this is for.
package invariants

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	introduction "github.com/stanleyHayes/obiara/services/api/internal/introduction"
	"github.com/stanleyHayes/obiara/services/api/internal/media/adapters/outbound/sharingpolicy"
	mediaapplication "github.com/stanleyHayes/obiara/services/api/internal/media/application"
	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
	poddomain "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
)

// oneAsset is a repository holding a single recording, owned by whoever the
// test says owns it.
type oneAsset struct{ asset mediadomain.Asset }

func (r oneAsset) FindByID(context.Context, string) (mediadomain.Asset, error) {
	return r.asset, nil
}

// anySigner signs whatever it is given. Signing is not what is under test.
type anySigner struct{}

func (anySigner) SignUpload(context.Context, mediaapplication.UploadSigningRequest) (mediaapplication.SignedAccess, error) {
	return mediaapplication.SignedAccess{URL: "https://example.invalid/upload"}, nil
}
func (anySigner) SignRead(context.Context, mediaapplication.ReadSigningRequest) (mediaapplication.SignedAccess, error) {
	return mediaapplication.SignedAccess{URL: "https://example.invalid/read"}, nil
}

func recording(t *testing.T, ownerID string) mediadomain.Asset {
	t.Helper()
	checksum, err := mediadomain.NewChecksum("sha256", strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	asset, err := mediadomain.NewAsset(mediadomain.NewAssetParams{
		ID: "asset-1", ObjectKey: "voices/asset-1", OwnerID: ownerID,
		ContentType: "audio/mpeg", Size: 1024, Checksum: checksum,
		Duration: 90 * time.Second, CreatedAt: time.Now().Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	return asset
}

// entitled is an entitlement that answers the same way every time. The real
// bridges are in main.go; what this file is about is whether the policy the
// module composes can admit a listener at all.
type entitled bool

func (e entitled) MayHear(context.Context, string, string, string) (bool, error) {
	return bool(e), nil
}

func accessService(t *testing.T, ownerID string, entitlement sharingpolicy.Entitlement) mediaapplication.AccessService {
	t.Helper()
	return mediaapplication.NewAccessService(
		oneAsset{recording(t, ownerID)},
		// The policy the API actually composes, with the purposes main.go
		// actually passes it.
		sharingpolicy.New(
			[]string{introduction.ConsentPurposeID, poddomain.PlaybackPurposeID},
			map[string]sharingpolicy.Entitlement{
				introduction.ConsentPurposeID: entitlement,
				poddomain.PlaybackPurposeID:   entitlement,
			},
		),
		anySigner{},
		time.Now,
	)
}

func TestARecipientCanHearWhatWasLeftForThem(t *testing.T) {
	// The last step of the delivery chain: a sow is screened, released,
	// placed in a pod, and rests at the recipient's house front. Opening it
	// asks the media context for a grant naming the recipient — who is not
	// the owner of the recording, because the sower recorded it.
	access := accessService(t, "the-sower", entitled(true))
	granted, err := access.RequestRead(context.Background(), mediaapplication.ReadRequest{
		SubjectID: "the-recipient",
		AssetID:   "asset-1",
		Purpose:   poddomain.PlaybackPurposeID,
		TTL:       5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("a recipient could not hear the pod left for them: %v", err)
	}
	if granted.URL == "" {
		t.Fatal("no playback url")
	}
}

func TestAMemberCanHearAnotherMembersVoiceOfIntroduction(t *testing.T) {
	// FR-202 arms a sow only after twenty seconds of the other person's
	// Voice of Introduction. If nobody can obtain a grant to hear somebody
	// else's recording, the gate can never be satisfied and no sow is ever
	// possible — in a product where people meet through their voices.
	access := accessService(t, "the-other-member", entitled(true))
	if _, err := access.RequestRead(context.Background(), mediaapplication.ReadRequest{
		SubjectID: "the-listener",
		AssetID:   "asset-1",
		Purpose:   introduction.ConsentPurposeID,
		TTL:       5 * time.Minute,
	}); err != nil {
		t.Fatalf("a member could not hear another member's voice: %v", err)
	}
}

func TestAStrangerStillCannotHelpThemselves(t *testing.T) {
	// The direction that must not change: opening delivery to a recipient
	// must not open every recording to everybody. This is here so that
	// whatever fixes the two tests above is checked against it.
	access := accessService(t, "the-owner", entitled(true))
	if _, err := access.RequestRead(context.Background(), mediaapplication.ReadRequest{
		SubjectID: "a-stranger",
		AssetID:   "asset-1",
		Purpose:   "some.other.purpose",
		TTL:       5 * time.Minute,
	}); !errors.Is(err, mediaapplication.ErrAccessDenied) {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
}

func TestAnUnentitledListenerIsStillRefused(t *testing.T) {
	// The other half of the same change. Opening delivery to a recipient
	// must not open every recording to everybody, so the entitlement is what
	// admits somebody and not the purpose alone.
	access := accessService(t, "the-sower", entitled(false))
	if _, err := access.RequestRead(context.Background(), mediaapplication.ReadRequest{
		SubjectID: "somebody-with-no-pod",
		AssetID:   "asset-1",
		Purpose:   poddomain.PlaybackPurposeID,
		TTL:       5 * time.Minute,
	}); !errors.Is(err, mediaapplication.ErrAccessDenied) {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
}

func TestTheOwnerIsAdmittedWithNoEntitlementAtAll(t *testing.T) {
	// Hearing your own recording back must not depend on a bridge being
	// reachable — it is how a member checks what they recorded.
	access := accessService(t, "the-member", entitled(false))
	if _, err := access.RequestRead(context.Background(), mediaapplication.ReadRequest{
		SubjectID: "the-member",
		AssetID:   "asset-1",
		Purpose:   introduction.ConsentPurposeID,
		TTL:       5 * time.Minute,
	}); err != nil {
		t.Fatalf("a member could not hear their own recording: %v", err)
	}
}
