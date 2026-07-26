// Package application revalidates the current definition and women-reviewer
// authority before recording a neutral readiness assessment.
package application

import (
	"context"
	"errors"
	"strings"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/womensreview/domain"
)

var (
	ErrInvalidRequest = errors.New("invalid women's-safety evidence request")
	ErrUnavailable    = errors.New("women's-safety evidence review unavailable")
)

type Request struct {
	CohortKey                 string
	ReviewerKey               string
	ExpectedDefinitionVersion uint64
	ExpectedAggregateVersion  uint64
}

type Service struct {
	definitions DefinitionCatalog
	aggregates  AggregateSource
	approvals   ApprovalSource
	authority   ReviewerAuthority
	sink        AssessmentSink
	ids         IDSource
	clock       Clock
}

func New(definitions DefinitionCatalog, aggregates AggregateSource, approvals ApprovalSource, authority ReviewerAuthority, sink AssessmentSink, ids IDSource, clock Clock) Service {
	return Service{definitions: definitions, aggregates: aggregates, approvals: approvals, authority: authority, sink: sink, ids: ids, clock: clock}
}

func (service Service) Assess(ctx context.Context, request Request) (domain.Assessment, error) {
	request.CohortKey = strings.TrimSpace(request.CohortKey)
	request.ReviewerKey = strings.TrimSpace(request.ReviewerKey)
	if service.definitions == nil || service.aggregates == nil || service.approvals == nil ||
		service.authority == nil || service.sink == nil || service.ids == nil || service.clock == nil ||
		!opaque(request.CohortKey) || !opaque(request.ReviewerKey) ||
		request.ExpectedDefinitionVersion == 0 || request.ExpectedAggregateVersion == 0 {
		return domain.Assessment{}, ErrInvalidRequest
	}
	if err := service.authority.AuthorizeCurrentWomenReviewer(ctx, request.ReviewerKey); err != nil {
		return domain.Assessment{}, ErrUnavailable
	}
	definition, err := service.definitions.Current(ctx)
	if err != nil || definition.Spec().Version != request.ExpectedDefinitionVersion {
		return domain.Assessment{}, ErrUnavailable
	}
	aggregate, err := service.aggregates.Current(ctx, request.CohortKey)
	if err != nil || aggregate.Spec().Version != request.ExpectedAggregateVersion ||
		aggregate.Spec().CohortKey != request.CohortKey {
		return domain.Assessment{}, ErrUnavailable
	}
	approval, err := service.approvals.Current(ctx, aggregate.Spec().ID, request.ReviewerKey)
	if err != nil || approval.Spec().ReviewerKey != request.ReviewerKey {
		return domain.Assessment{}, ErrUnavailable
	}
	assessment, err := domain.Evaluate(service.ids.NewID(), definition, aggregate, approval, service.clock.Now())
	if err != nil {
		return domain.Assessment{}, ErrUnavailable
	}

	// Re-read the definition and revalidate reviewer authority immediately
	// before the only write. No release or enforcement action is exposed.
	current, err := service.definitions.Current(ctx)
	if err != nil || current.Spec().ID != definition.Spec().ID ||
		current.Spec().Version != definition.Spec().Version {
		return domain.Assessment{}, ErrUnavailable
	}
	if err := service.authority.AuthorizeCurrentWomenReviewer(ctx, request.ReviewerKey); err != nil {
		return domain.Assessment{}, ErrUnavailable
	}
	if err := service.sink.Record(ctx, assessment); err != nil {
		return domain.Assessment{}, ErrUnavailable
	}
	return assessment, nil
}

func opaque(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' && character < 'a' || character > 'f' {
			return false
		}
	}
	return true
}
