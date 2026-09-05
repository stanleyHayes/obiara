package deferred_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/deferred"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/adapters/outbound/locale"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/screening/application"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

// routedReview records what the composed policy sent to a person.
type routedReview struct {
	cases []application.ReviewCase
}

func (review *routedReview) Route(_ context.Context, reviewCase application.ReviewCase) (string, error) {
	review.cases = append(review.cases, reviewCase)
	return reviewCase.ID, nil
}

type sequentialIDs struct{ n int }

func (ids *sequentialIDs) NewID() string {
	ids.n++
	sum := sha256.Sum256([]byte{byte(ids.n)})
	return hex.EncodeToString(sum[:])
}

type fixedMedia struct{}

func (fixedMedia) Inspect(context.Context, string) (application.MediaMetadata, error) {
	return application.MediaMetadata{MIME: "audio/ogg", Bytes: 2048, DurationMs: 45_000}, nil
}

// composed builds the screening adapter exactly as the composition root would
// under the owner's ruling: no machine opinion, no automated adjudication.
func composed(human application.HumanReview, reviewed ...locale.Reviewed) application.Adapter {
	return application.New(
		locale.NewSource("tw"),
		locale.NewCatalog(reviewed...),
		fixedMedia{},
		deferred.NewAdvisor(),
		deferred.NewAdjudicator(),
		human,
		&sequentialIDs{},
	)
}

func reviewedAt() time.Time {
	return time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
}

func TestEverySowReachesAPerson(t *testing.T) {
	// The ruling, expressed as a property of the composed policy rather than
	// as a comment: there is no input that gets past it. A sow either goes to
	// a reviewer or is refused for being structurally malformed; nothing is
	// delivered on a machine's say-so.
	human := &routedReview{}
	adapter := composed(human)

	for name, body := range map[string]string{
		"ordinary":        "I liked what you said about your grandmother.",
		"a phone number":  "call me on 024 000 0000",
		"asking for cash": "send me 50 cedis",
	} {
		// Routing to a person comes back as ErrHumanReviewRequired with the
		// reference. It is the sow context's own error value, which is what
		// lets the sow service recognise it and hold rather than fail.
		decision, err := adapter.Screen(context.Background(), body, []string{"asset-1"})
		if !errors.Is(err, sowapplication.ErrHumanReviewRequired) {
			t.Fatalf("%s: err = %v, want ErrHumanReviewRequired", name, err)
		}
		if decision.Approved {
			t.Fatalf("%s: a sow was approved with no person involved", name)
		}
		if decision.Reference == "" {
			t.Fatalf("%s: routed with no reference, so the sow could never be found again", name)
		}
	}
	if len(human.cases) != 3 {
		t.Fatalf("%d sows reached a person, want 3", len(human.cases))
	}
}

func TestAReviewedLanguageStillReachesAPerson(t *testing.T) {
	// Reviewing a language changes who can read a sow, never whether it is
	// delivered. With no machine opinion, a resolved locale still routes.
	human := &routedReview{}
	adapter := composed(human, locale.Reviewed{
		Tag: "tw", Version: 1, ReviewedAt: reviewedAt(),
	})

	decision, err := adapter.Screen(context.Background(), "wo ho te sɛn", nil)
	if !errors.Is(err, sowapplication.ErrHumanReviewRequired) {
		t.Fatalf("err = %v, want ErrHumanReviewRequired", err)
	}
	if decision.Approved {
		t.Fatal("a reviewed language was treated as a reason to deliver")
	}
	if len(human.cases) != 1 {
		t.Fatalf("%d sows reached a person, want 1", len(human.cases))
	}
	// The reason should now be the honest one — nobody had an opinion —
	// rather than the language being unreadable.
	if got := human.cases[0].Reason; got != application.ReasonUncertain {
		t.Fatalf("reason = %q, want uncertain", got)
	}
}

func TestAnAdjudicationIsNeverClaimedWithoutAHuman(t *testing.T) {
	if _, err := (deferred.Adjudicator{}).Decide(context.Background(),
		application.ScreeningInput{Text: "hello"},
		application.Advisory{Status: application.StatusApproved},
	); !errors.Is(err, deferred.ErrNoAdjudication) {
		t.Fatal("the adjudicator claimed a decision nobody made")
	}
	var _ sowapplication.Screening = composed(&routedReview{})
}
