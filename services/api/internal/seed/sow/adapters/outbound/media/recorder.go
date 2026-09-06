package media

import (
	"context"
	"errors"
	"time"

	mediaapplication "github.com/stanleyHayes/obiara/services/api/internal/media/application"
	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

// RecordingPurposeID is what a sow's recording is uploaded and played under.
//
// Distinct from the Voice of Introduction's purpose on purpose: the two are
// heard by different people under different rules, and a shared purpose would
// mean a rule written for one silently governed the other.
const RecordingPurposeID = "seed.sow.recording"

// uploadGrantTTL is how long a member has to send the bytes. Long enough for a
// ninety-second clip on a slow connection, short enough that a leaked URL is
// not a standing write grant on the bucket.
const uploadGrantTTL = 15 * time.Minute

// Access and Store are the parts of the media context a sow's recording needs.
type Access interface {
	RequestUpload(context.Context, mediaapplication.UploadRequest) (mediaapplication.SignedAccess, error)
}

type Store interface {
	Register(context.Context, mediadomain.Asset) error
	FindByID(context.Context, string) (mediadomain.Asset, error)
}

// IDs names a new recording.
type IDs interface{ NewID() string }

// Recorder opens a recording for a sow.
//
// A sow carries the member's answer in their own voice, and until this existed
// there was nowhere to put one: the only upload path in the product created a
// Voice of Introduction, which is a different thing — it answers one of three
// fixed questions and is offered to anyone who may hear the member. Sending a
// sow required a recording that could not be made (agent_plan.md §68).
type Recorder struct {
	access Access
	store  Store
	ids    IDs
	now    func() time.Time
}

func NewRecorder(access Access, store Store, ids IDs, now func() time.Time) Recorder {
	if now == nil {
		now = time.Now
	}
	return Recorder{access: access, store: store, ids: ids, now: now}
}

// Recording is what a member needs in order to send their answer's bytes.
type Recording struct {
	AssetID   string
	URL       string
	ExpiresAt time.Time
}

// ErrNotRecordable refuses a description a recording could not have.
var ErrNotRecordable = errors.New("that recording cannot be opened")

// Open registers the asset and returns a grant to upload it.
//
// The description comes in before the bytes because the grant is signed over
// the length and the digest, so the store itself refuses anything else. The
// length is the one claim, and it is bounded against the byte count rather
// than believed — nothing here can decode audio.
func (recorder Recorder) Open(
	ctx context.Context,
	ownerID, contentType string,
	sizeBytes int64,
	checksum string,
	duration time.Duration,
) (Recording, error) {
	if recorder.access == nil || recorder.store == nil || recorder.ids == nil {
		return Recording{}, errors.New("sow recording is not composed")
	}
	if !mediadomain.PlausibleDuration(contentType, sizeBytes, duration) {
		return Recording{}, ErrNotRecordable
	}
	digest, err := mediadomain.NewChecksum("sha256", checksum)
	if err != nil {
		return Recording{}, ErrNotRecordable
	}
	assetID := recorder.ids.NewID()
	asset, err := mediadomain.NewAsset(mediadomain.NewAssetParams{
		ID: assetID, ObjectKey: "sows/" + assetID, OwnerID: ownerID,
		ContentType: contentType, Size: sizeBytes, Checksum: digest,
		Duration: duration, CreatedAt: recorder.now().UTC(),
	})
	if err != nil {
		return Recording{}, ErrNotRecordable
	}
	if err := recorder.store.Register(ctx, asset); err != nil {
		return Recording{}, sowapplication.ErrUnavailable
	}
	signed, err := recorder.access.RequestUpload(ctx, mediaapplication.UploadRequest{
		SubjectID: ownerID,
		Purpose:   RecordingPurposeID,
		Asset:     asset,
		TTL:       uploadGrantTTL,
	})
	if err != nil {
		return Recording{}, sowapplication.ErrUnavailable
	}
	return Recording{AssetID: assetID, URL: signed.URL, ExpiresAt: signed.ExpiresAt}, nil
}
