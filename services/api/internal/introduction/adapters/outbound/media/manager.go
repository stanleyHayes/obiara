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

// Assets reads back what storage recorded once the bytes arrived.
type Assets interface {
	FindByID(context.Context, string) (mediadomain.Asset, error)
}

// Remover erases the object. Deletion is the media context's job because it
// owns retention and legal hold; the introduction only asks.
type Remover interface {
	Delete(context.Context, string) error
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
	subjectID, assetID, contentType string,
) (introapplication.UploadAccess, error) {
	if manager.access == nil || manager.assets == nil {
		return introapplication.UploadAccess{}, introapplication.ErrDependencyUnavailable
	}
	asset, err := manager.assets.FindByID(ctx, assetID)
	if err != nil {
		return introapplication.UploadAccess{}, introapplication.ErrNotFound
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
	_ = contentType // the asset's own recorded type is authoritative
	return introapplication.UploadAccess{URL: signed.URL, ExpiresAt: signed.ExpiresAt}, nil
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
