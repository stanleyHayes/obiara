// Package application provides the stateless E11-S09 adapter behind the
// existing sow Screening port. It does not accept, deliver, persist, or spend.
package application

import (
	"context"

	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

type Adapter struct {
	locales     LocaleSource
	catalog     LocaleCatalog
	media       MediaInspector
	advisor     Advisor
	adjudicator Adjudicator
	human       HumanReview
	ids         IDSource
}

func New(
	locales LocaleSource,
	catalog LocaleCatalog,
	media MediaInspector,
	advisor Advisor,
	adjudicator Adjudicator,
	human HumanReview,
	ids IDSource,
) Adapter {
	return Adapter{locales: locales, catalog: catalog, media: media, advisor: advisor, adjudicator: adjudicator, human: human, ids: ids}
}

// Screen implements seed/sow/application.Screening.
func (adapter Adapter) Screen(ctx context.Context, body string, mediaRefs []string) (sowapplication.ScreeningDecision, error) {
	if !adapter.ready() || len(mediaRefs) > MaxMedia {
		return sowapplication.ScreeningDecision{}, ErrInvalidInput
	}
	text, err := normalizeText(body)
	if err != nil {
		return adapter.rejectStructural()
	}
	tag, err := adapter.locales.CurrentLocale(ctx)
	if err != nil {
		return adapter.route(ctx, ScreeningInput{Text: text, LocaleTag: "und"}, ReasonUnsupportedLocale, nil)
	}
	review, err := adapter.catalog.Resolve(ctx, tag)
	if err != nil || !validLocale(review) {
		return adapter.route(ctx, ScreeningInput{Text: text, LocaleTag: tag}, ReasonUnsupportedLocale, nil)
	}
	input := ScreeningInput{Text: text, LocaleTag: review.Tag, LocaleVersion: review.Version}
	for _, ref := range mediaRefs {
		metadata, inspectErr := adapter.media.Inspect(ctx, ref)
		if inspectErr != nil || !validMedia(metadata) {
			return adapter.route(ctx, input, ReasonUnsupportedMedia, nil)
		}
		input.Media = append(input.Media, metadata)
	}
	advisory, err := adapter.advisor.Screen(ctx, cloneInput(input))
	if err != nil {
		return adapter.route(ctx, input, ReasonProviderFailure, nil)
	}
	if !validAdvisory(advisory) || advisory.Status == StatusUncertain {
		return adapter.route(ctx, input, ReasonUncertain, &advisory)
	}
	adjudication, err := adapter.adjudicator.Decide(ctx, cloneInput(input), cloneAdvisory(advisory))
	if err != nil || !validAdjudication(adjudication) {
		return adapter.route(ctx, input, ReasonUncertain, &advisory)
	}
	return sowapplication.ScreeningDecision{
		Approved:  adjudication.Status == StatusApproved,
		Reference: adjudication.Reference,
	}, nil
}

func (adapter Adapter) rejectStructural() (sowapplication.ScreeningDecision, error) {
	reference := adapter.ids.NewID()
	if !opaquePattern.MatchString(reference) {
		return sowapplication.ScreeningDecision{}, ErrInvalidInput
	}
	return sowapplication.ScreeningDecision{Approved: false, Reference: reference}, nil
}

func (adapter Adapter) route(
	ctx context.Context,
	input ScreeningInput,
	reason ReasonCode,
	advisory *Advisory,
) (sowapplication.ScreeningDecision, error) {
	id := adapter.ids.NewID()
	if !opaquePattern.MatchString(id) {
		return sowapplication.ScreeningDecision{}, ErrHumanReviewRequired
	}
	reference, err := adapter.human.Route(ctx, ReviewCase{
		ID: id, Input: cloneInput(input), Reason: reason, Advisory: cloneAdvisoryPointer(advisory),
	})
	if err != nil || !opaquePattern.MatchString(reference) {
		return sowapplication.ScreeningDecision{}, ErrHumanReviewRequired
	}
	return sowapplication.ScreeningDecision{Approved: false, Reference: reference}, ErrHumanReviewRequired
}

func (adapter Adapter) ready() bool {
	return adapter.locales != nil && adapter.catalog != nil && adapter.media != nil &&
		adapter.advisor != nil && adapter.adjudicator != nil && adapter.human != nil && adapter.ids != nil
}

func cloneInput(input ScreeningInput) ScreeningInput {
	input.Media = append([]MediaMetadata(nil), input.Media...)
	return input
}

func cloneAdvisory(advisory Advisory) Advisory {
	advisory.Reasons = append([]ReasonCode(nil), advisory.Reasons...)
	return advisory
}

func cloneAdvisoryPointer(advisory *Advisory) *Advisory {
	if advisory == nil {
		return nil
	}
	copy := cloneAdvisory(*advisory)
	return &copy
}
