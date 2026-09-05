package locale

import (
	"context"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
)

var reviewedAt = time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)

func TestAnUnreviewedLanguageGoesToAPersonRatherThanBeingRefused(t *testing.T) {
	// The screening adapter turns this into ReasonUnsupportedLocale and
	// routes, so an empty catalog is a safe configuration: every sow reaches
	// a reviewer. It is not a refusal and the member never sees it.
	catalog := NewCatalog()
	if _, err := catalog.Resolve(context.Background(), "ak"); err == nil {
		t.Fatal("an unreviewed language resolved")
	}
}

func TestOnlyACompleteReviewCounts(t *testing.T) {
	// A half-filled entry would claim a language review that did not happen,
	// which under CL-REG-07 is exactly the thing not to do.
	catalog := NewCatalog(
		Reviewed{Tag: "en-GH", Version: 0, ReviewedAt: reviewedAt},
		Reviewed{Tag: "ak", Version: 1},
		Reviewed{Tag: "  ", Version: 1, ReviewedAt: reviewedAt},
		Reviewed{Tag: "tw", Version: 2, ReviewedAt: reviewedAt},
	)
	for _, tag := range []string{"en-GH", "ak", ""} {
		if _, err := catalog.Resolve(context.Background(), tag); err == nil {
			t.Fatalf("%q counted as reviewed", tag)
		}
	}
	review, err := catalog.Resolve(context.Background(), "tw")
	if err != nil {
		t.Fatal(err)
	}
	if !review.Reviewed || review.Version != 2 || review.ReviewedAt.IsZero() {
		t.Fatalf("review = %#v", review)
	}
	var _ application.LocaleReview = review
}

func TestAnUnsetSourceSaysUndeterminedRatherThanGuessing(t *testing.T) {
	// Guessing a language from a device would route somebody's words to a
	// reviewer who cannot read them.
	tag, err := NewSource("").CurrentLocale(context.Background())
	if err != nil || tag != "und" {
		t.Fatalf("tag = %q, err = %v", tag, err)
	}
	if tag, _ := NewSource(" tw ").CurrentLocale(context.Background()); tag != "tw" {
		t.Fatalf("tag = %q", tag)
	}
}
