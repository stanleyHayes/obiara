package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/safety/womensreview/domain"
)

var now = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func key(char string) string { return strings.Repeat(char, 64) }

func fixtures(t *testing.T) (domain.Definition, domain.Aggregate, domain.Approval) {
	t.Helper()
	definition, err := domain.NewDefinition(domain.DefinitionSpec{
		ID: "women.safety.v1", Version: 2, MinimumCohort: 50,
		MinimumResponsePermille: 500,
		Dimensions:              []domain.DimensionRule{{Dimension: domain.DimensionPrivacyControl, MinimumEvidence: 20}},
		ReviewID:                "women.panel.2", ReviewedByWomenPanelKey: key("a"), ReviewedAt: now.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	aggregate, err := domain.NewAggregate(domain.AggregateSpec{
		ID: key("b"), Version: 4, CohortKey: key("c"), EligibleCount: 60, RespondentCount: 40,
		Evidence:        []domain.DimensionEvidence{{Dimension: domain.DimensionPrivacyControl, EvidenceCount: 30, Favorable: 20}},
		WindowStartedAt: now.Add(-24 * time.Hour), WindowEndedAt: now.Add(-time.Hour), RedactionVersion: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	approval, err := domain.NewApproval(domain.ApprovalSpec{
		ID: key("d"), DefinitionID: "women.safety.v1", DefinitionVersion: 2,
		AggregateID: key("b"), AggregateVersion: 4, ReviewerKey: key("e"),
		Decision: domain.DecisionApprove, ReviewedDimensions: []domain.Dimension{domain.DimensionPrivacyControl},
		ReviewedAt: now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	return definition, aggregate, approval
}

func TestAssessRevalidatesDefinitionAndWomenReviewerBeforeRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	definitions := NewMockDefinitionCatalog(ctrl)
	aggregates := NewMockAggregateSource(ctrl)
	approvals := NewMockApprovalSource(ctrl)
	authority := NewMockReviewerAuthority(ctrl)
	sink := NewMockAssessmentSink(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	definition, aggregate, approval := fixtures(t)

	gomock.InOrder(
		authority.EXPECT().AuthorizeCurrentWomenReviewer(gomock.Any(), key("e")).Return(nil),
		definitions.EXPECT().Current(gomock.Any()).Return(definition, nil),
		aggregates.EXPECT().Current(gomock.Any(), key("c")).Return(aggregate, nil),
		approvals.EXPECT().Current(gomock.Any(), key("b"), key("e")).Return(approval, nil),
		ids.EXPECT().NewID().Return(key("f")),
		clock.EXPECT().Now().Return(now),
		definitions.EXPECT().Current(gomock.Any()).Return(definition, nil),
		authority.EXPECT().AuthorizeCurrentWomenReviewer(gomock.Any(), key("e")).Return(nil),
		sink.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, assessment domain.Assessment) error {
			if assessment.Outcome != domain.OutcomeReadyForReleaseReview {
				t.Fatalf("unexpected outcome: %+v", assessment)
			}
			return nil
		}),
	)
	service := New(definitions, aggregates, approvals, authority, sink, ids, clock)
	assessment, err := service.Assess(context.Background(), Request{
		CohortKey: key("c"), ReviewerKey: key("e"),
		ExpectedDefinitionVersion: 2, ExpectedAggregateVersion: 4,
	})
	if err != nil || assessment.ID != key("f") {
		t.Fatalf("Assess() = (%+v, %v)", assessment, err)
	}
}

func TestAssessFailsClosedWhenDefinitionChangesBeforeRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	definitions := NewMockDefinitionCatalog(ctrl)
	aggregates := NewMockAggregateSource(ctrl)
	approvals := NewMockApprovalSource(ctrl)
	authority := NewMockReviewerAuthority(ctrl)
	sink := NewMockAssessmentSink(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	definition, aggregate, approval := fixtures(t)
	changedSpec := definition.Spec()
	changedSpec.Version++
	changed, err := domain.NewDefinition(changedSpec)
	if err != nil {
		t.Fatal(err)
	}
	gomock.InOrder(
		authority.EXPECT().AuthorizeCurrentWomenReviewer(gomock.Any(), key("e")).Return(nil),
		definitions.EXPECT().Current(gomock.Any()).Return(definition, nil),
		aggregates.EXPECT().Current(gomock.Any(), key("c")).Return(aggregate, nil),
		approvals.EXPECT().Current(gomock.Any(), key("b"), key("e")).Return(approval, nil),
		ids.EXPECT().NewID().Return(key("f")),
		clock.EXPECT().Now().Return(now),
		definitions.EXPECT().Current(gomock.Any()).Return(changed, nil),
	)
	service := New(definitions, aggregates, approvals, authority, sink, ids, clock)
	_, err = service.Assess(context.Background(), Request{
		CohortKey: key("c"), ReviewerKey: key("e"),
		ExpectedDefinitionVersion: 2, ExpectedAggregateVersion: 4,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected fail closed, got %v", err)
	}
}

func TestAssessFailsClosedWhenReviewerAuthorityRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	definitions := NewMockDefinitionCatalog(ctrl)
	aggregates := NewMockAggregateSource(ctrl)
	approvals := NewMockApprovalSource(ctrl)
	authority := NewMockReviewerAuthority(ctrl)
	sink := NewMockAssessmentSink(ctrl)
	ids := NewMockIDSource(ctrl)
	clock := NewMockClock(ctrl)
	definition, aggregate, approval := fixtures(t)
	gomock.InOrder(
		authority.EXPECT().AuthorizeCurrentWomenReviewer(gomock.Any(), key("e")).Return(nil),
		definitions.EXPECT().Current(gomock.Any()).Return(definition, nil),
		aggregates.EXPECT().Current(gomock.Any(), key("c")).Return(aggregate, nil),
		approvals.EXPECT().Current(gomock.Any(), key("b"), key("e")).Return(approval, nil),
		ids.EXPECT().NewID().Return(key("f")),
		clock.EXPECT().Now().Return(now),
		definitions.EXPECT().Current(gomock.Any()).Return(definition, nil),
		authority.EXPECT().AuthorizeCurrentWomenReviewer(gomock.Any(), key("e")).Return(errors.New("revoked")),
	)
	service := New(definitions, aggregates, approvals, authority, sink, ids, clock)
	_, err := service.Assess(context.Background(), Request{
		CohortKey: key("c"), ReviewerKey: key("e"),
		ExpectedDefinitionVersion: 2, ExpectedAggregateVersion: 4,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected fail closed, got %v", err)
	}
}

func TestAssessRejectsStaleExpectedVersionsBeforeRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	definitions := NewMockDefinitionCatalog(ctrl)
	definition, _, _ := fixtures(t)
	authority := NewMockReviewerAuthority(ctrl)
	authority.EXPECT().AuthorizeCurrentWomenReviewer(gomock.Any(), key("e")).Return(nil)
	definitions.EXPECT().Current(gomock.Any()).Return(definition, nil)
	service := New(
		definitions, NewMockAggregateSource(ctrl), NewMockApprovalSource(ctrl),
		authority, NewMockAssessmentSink(ctrl), NewMockIDSource(ctrl), NewMockClock(ctrl),
	)
	_, err := service.Assess(context.Background(), Request{
		CohortKey: key("c"), ReviewerKey: key("e"),
		ExpectedDefinitionVersion: 99, ExpectedAggregateVersion: 4,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected stale version rejection, got %v", err)
	}
}
