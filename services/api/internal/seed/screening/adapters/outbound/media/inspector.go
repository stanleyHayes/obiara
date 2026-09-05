// Package media tells the screening policy how big and how long a sow's
// recordings are, so it can refuse the ones nobody should be asked to review.
package media

import (
	"context"
	"errors"

	mediadomain "github.com/stanleyHayes/obiara/services/api/internal/media/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
)

// ErrAssetUnavailable reports a recording the media context cannot describe.
var ErrAssetUnavailable = errors.New("recording could not be inspected")

// Assets reads back what storage recorded about an asset.
type Assets interface {
	FindByID(context.Context, string) (mediadomain.Asset, error)
}

// Inspector implements the screening context's MediaInspector port.
type Inspector struct{ assets Assets }

func NewInspector(assets Assets) Inspector { return Inspector{assets: assets} }

// Inspect returns the shape of a recording, never its bytes.
//
// The screening policy only needs to know what kind of thing this is and how
// long it runs; the words are already in the review's text and the audio
// itself is not something this port is entitled to hand around.
//
// An unreadable or deleted asset returns an error, which the policy turns
// into `ReasonUnsupportedMedia` and routes to a person. That is the right
// outcome: a sow referencing a recording nothing can describe is exactly what
// somebody should look at.
func (inspector Inspector) Inspect(ctx context.Context, ref string) (application.MediaMetadata, error) {
	if inspector.assets == nil {
		return application.MediaMetadata{}, ErrAssetUnavailable
	}
	asset, err := inspector.assets.FindByID(ctx, ref)
	if err != nil || asset.IsDeleted() {
		return application.MediaMetadata{}, ErrAssetUnavailable
	}
	return application.MediaMetadata{
		MIME:       asset.ContentType(),
		Bytes:      asset.Size(),
		DurationMs: asset.Duration().Milliseconds(),
	}, nil
}
