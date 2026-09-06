// Package media binds the Voice of Introduction to the media context.
//
// The introduction context asks for an upload grant, the size and type of
// what actually landed, and deletion. The media context already owns all
// three behind its own authorization boundary; this only translates between
// the two vocabularies so neither has to know the other's types.
package media

import (
	"context"
	"errors"
	"time"

	introapplication "github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	introdomain "github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
	mediaapplication "github.com/stanleyHayes/obiara/services/api/internal/media/application"
	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

// uploadGrantTTL is how long a member has to actually send the bytes. Long
// enough for a slow Accra connection to finish a two-minute clip, short
// enough that a leaked URL is not a standing write grant on the bucket.
const uploadGrantTTL = 15 * time.Minute

// Access is the subset of the media access service this needs.
type Access interface {
	RequestUpload(context.Context, mediaapplication.UploadRequest) (mediaapplication.SignedAccess, error)
	RequestRead(context.Context, mediaapplication.ReadRequest) (mediaapplication.SignedAccess, error)
}

// Assets is the media context's record of one recording.
//
// Register is here because nothing called it. The media context authorizes an
// upload grant against a known owner and object key, so the row has to exist
// before the grant does — and with nothing writing one, every begin-upload
// answered "not found" and no member could record anything at all
// (agent_plan.md §67).
type Assets interface {
	FindByID(context.Context, string) (mediadomain.Asset, error)
	Register(context.Context, mediadomain.Asset) error
}

// Remover erases the object, and reports whether the bytes are there.
//
// Deletion is the media context's job because it owns retention and legal
// hold; the introduction only asks. Stat is here for the other end of the
// same story: confirming an upload on a member's word alone would mark a
// recording ready when the bytes never left their phone, and every play of it
// would 404 at the bucket — which reads as a broken product rather than a
// failed upload.
type Remover interface {
	Delete(context.Context, string) error
	Stat(context.Context, string) (int64, error)
}

type Manager struct {
	access  Access
	assets  Assets
	remover Remover
	purpose string
	now     func() time.Time
}

func NewManager(access Access, assets Assets, remover Remover, purpose string, now func() time.Time) *Manager {
	if now == nil {
		now = time.Now
	}
	return &Manager{access: access, assets: assets, remover: remover, purpose: purpose, now: now}
}

// AuthorizeUpload asks the media context for a write grant on one asset.
//
// The asset row is registered before the grant is issued, not after the bytes
// land: the media context authorizes against a known owner and object key, so
// there has to be something to authorize against. An asset whose upload is
// never completed simply keeps a zero size and is swept by its own expiry.
func (manager *Manager) AuthorizeUpload(
	ctx context.Context,
	subjectID string,
	media introdomain.MediaRef,
) (introapplication.UploadAccess, error) {
	if manager.access == nil || manager.assets == nil {
		return introapplication.UploadAccess{}, introapplication.ErrDependencyUnavailable
	}
	checksum, err := mediadomain.NewChecksum("sha256", media.Checksum())
	if err != nil {
		return introapplication.UploadAccess{}, introdomain.ErrInvalidIntroduction
	}
	now := manager.now().UTC()
	asset, err := mediadomain.NewAsset(mediadomain.NewAssetParams{
		ID: media.AssetID(), ObjectKey: objectKey(subjectID, media.AssetID()),
		OwnerID: subjectID, ContentType: media.ContentType(),
		Size: media.Size(), Checksum: checksum, Duration: media.Duration(),
		CreatedAt: now,
	})
	if err != nil {
		return introapplication.UploadAccess{}, introdomain.ErrInvalidIntroduction
	}
	// Registered before the grant is issued, because the access service
	// authorizes against a known owner and object key. Re-registering the
	// same asset is a retried begin-upload and is not an error: the row is
	// already what this would write.
	if registerErr := manager.assets.Register(ctx, asset); registerErr != nil {
		existing, findErr := manager.assets.FindByID(ctx, media.AssetID())
		if findErr != nil || existing.OwnerID() != subjectID {
			return introapplication.UploadAccess{}, introapplication.ErrDependencyUnavailable
		}
		asset = existing
	}
	signed, err := manager.access.RequestUpload(ctx, mediaapplication.UploadRequest{
		SubjectID: subjectID,
		Purpose:   manager.purpose,
		Asset:     asset,
		TTL:       uploadGrantTTL,
	})
	if err != nil {
		return introapplication.UploadAccess{}, introapplication.ErrDependencyUnavailable
	}
	return introapplication.UploadAccess{URL: signed.URL, ExpiresAt: signed.ExpiresAt}, nil
}

// objectKey is where a recording lives in the bucket.
//
// Keyed by asset id and not by anything about the member: an object key ends
// up in logs, in bucket listings and in a signed URL's path, and a key that
// carried a member id would put them in all three.
func objectKey(_ string, assetID string) string {
	return "voices/" + assetID
}

// Inspect reports what storage actually accepted, which is the only account
// of the recording the introduction is allowed to trust. A client-declared
// size or duration would let a member claim a twenty-second answer they never
// gave, and the listening gate is built on that number.
func (manager *Manager) Inspect(ctx context.Context, assetID string) (introdomain.MediaRef, error) {
	if manager.assets == nil {
		return introdomain.MediaRef{}, introapplication.ErrDependencyUnavailable
	}
	asset, err := manager.assets.FindByID(ctx, assetID)
	if err != nil {
		if errors.Is(err, mediadomain.ErrAssetUnavailable) {
			return introdomain.MediaRef{}, introapplication.ErrNotFound
		}
		return introdomain.MediaRef{}, introapplication.ErrDependencyUnavailable
	}
	// The bytes are the point. The store signed the grant over this exact
	// length and digest, so if an object of this size is there, it is this
	// recording — but something has to ask whether it is there at all.
	if manager.remover != nil {
		stored, statErr := manager.remover.Stat(ctx, asset.ObjectKey())
		if statErr != nil {
			return introdomain.MediaRef{}, introapplication.ErrUploadNotArrived
		}
		if stored != asset.Size() {
			return introdomain.MediaRef{}, introapplication.ErrUploadNotArrived
		}
	}
	return introdomain.NewMediaRef(
		asset.ID(),
		asset.ContentType(),
		asset.Size(),
		asset.Duration(),
		asset.Checksum().Value(),
	)
}

// playbackGrantTTL is how long a play URL stays good. Long enough to listen
// to a two-minute answer twice over a bad connection, short enough that a URL
// pasted somewhere else has stopped working before it travels.
const playbackGrantTTL = 10 * time.Minute

// AuthorizePlayback issues a read grant for one recording.
//
// The media context decides whether this subject may hear this asset; that
// rule is owner-only today, and gains a second clause when introductions are
// delivered. Nothing about the decision is made here.
func (manager *Manager) AuthorizePlayback(
	ctx context.Context,
	subjectID, assetID string,
) (introapplication.UploadAccess, error) {
	if manager.access == nil {
		return introapplication.UploadAccess{}, introapplication.ErrDependencyUnavailable
	}
	signed, err := manager.access.RequestRead(ctx, mediaapplication.ReadRequest{
		SubjectID: subjectID,
		AssetID:   assetID,
		Purpose:   manager.purpose,
		TTL:       playbackGrantTTL,
	})
	if err != nil {
		if errors.Is(err, mediaapplication.ErrAccessDenied) {
			// Refused and absent are answered the same way upstream: a
			// distinct "you may not hear this" would confirm the recording
			// exists and belongs to somebody.
			return introapplication.UploadAccess{}, introapplication.ErrNotFound
		}
		return introapplication.UploadAccess{}, introapplication.ErrDependencyUnavailable
	}
	return introapplication.UploadAccess{URL: signed.URL, ExpiresAt: signed.ExpiresAt}, nil
}

func (manager *Manager) Delete(ctx context.Context, assetID string) error {
	if manager.remover == nil {
		return introapplication.ErrDependencyUnavailable
	}
	if err := manager.remover.Delete(ctx, assetID); err != nil {
		return introapplication.ErrDependencyUnavailable
	}
	return nil
}
