// Package locale answers which languages a sow's words have been read in.
//
// The screening policy does not reject an unreviewed locale — it sends the
// sow to a person (`ReasonUnsupportedLocale`). So this catalog decides who
// reads a sow, never whether one is delivered, and an empty catalog is a
// safe configuration rather than a broken one: everything goes to the queue.
//
// CL-REG-07 forbids machine-only translation for launch languages, so a tag
// belongs here only once somebody has actually reviewed the language.
package locale

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
)

// ErrUnknownLocale reports a tag nobody has reviewed. The caller routes to a
// human rather than refusing, so this is not an error the member ever sees.
var ErrUnknownLocale = errors.New("locale has not been language-reviewed")

// Reviewed is one language somebody has read and signed off.
type Reviewed struct {
	Tag        string
	Version    uint64
	ReviewedAt time.Time
}

// Source reports the locale a sow is being screened in.
//
// It is fixed rather than derived from the request: a member's device
// language is not evidence of what language they wrote in, and guessing wrong
// would route their words to somebody who cannot read them.
type Source struct{ tag string }

func NewSource(tag string) Source { return Source{tag: strings.TrimSpace(tag)} }

func (source Source) CurrentLocale(context.Context) (string, error) {
	if source.tag == "" {
		// "und" is the undetermined tag. The adapter treats an unresolvable
		// locale as a reason to involve a person, which is the right answer
		// when nobody has said what language to expect.
		return "und", nil
	}
	return source.tag, nil
}

// Catalog resolves a tag to its language review.
type Catalog struct{ reviewed map[string]Reviewed }

// NewCatalog builds the catalog from the languages that have been reviewed.
// An empty catalog sends every sow to a person, which is a safe default and
// the current policy anyway.
func NewCatalog(reviewed ...Reviewed) Catalog {
	index := make(map[string]Reviewed, len(reviewed))
	for _, entry := range reviewed {
		tag := strings.TrimSpace(entry.Tag)
		if tag == "" || entry.Version == 0 || entry.ReviewedAt.IsZero() {
			// A half-filled entry would claim a review that did not happen.
			continue
		}
		entry.Tag = tag
		index[tag] = entry
	}
	return Catalog{reviewed: index}
}

func (catalog Catalog) Resolve(_ context.Context, tag string) (application.LocaleReview, error) {
	entry, found := catalog.reviewed[strings.TrimSpace(tag)]
	if !found {
		return application.LocaleReview{}, ErrUnknownLocale
	}
	return application.LocaleReview{
		Tag: entry.Tag, Version: entry.Version,
		Reviewed: true, ReviewedAt: entry.ReviewedAt,
	}, nil
}
