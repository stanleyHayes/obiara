package domain

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func opaque(char string) string { return strings.Repeat(char, 64) }

func validDefinition(t *testing.T) Definition {
	t.Helper()
	definition, err := NewDefinition(DefinitionSpec{
		ID: "women.safety.v1", Version: 3, MinimumCohort: 100,
		MinimumResponsePermille: 600,
		Dimensions: []DimensionRule{
			{Dimension: DimensionPrivacyControl, MinimumEvidence: 50},
			{Dimension: DimensionHarassmentResponse, MinimumEvidence: 50},
		},
		ReviewID: "review.panel.3", ReviewedByWomenPanelKey: opaque("a"),
		ReviewedAt: testNow.Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	return definition
}

func validAggregate(t *testing.T) Aggregate {
	t.Helper()
	aggregate, err := NewAggregate(AggregateSpec{
		ID: opaque("b"), Version: 7, CohortKey: opaque("c"),
		EligibleCount: 100, RespondentCount: 70,
		Evidence: []DimensionEvidence{
			{Dimension: DimensionHarassmentResponse, EvidenceCount: 60, Favorable: 42},
			{Dimension: DimensionPrivacyControl, EvidenceCount: 55, Favorable: 40},
		},
		WindowStartedAt: testNow.Add(-30 * 24 * time.Hour),
		WindowEndedAt:   testNow.Add(-time.Hour), RedactionVersion: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	return aggregate
}

func approvalFor(t *testing.T, aggregate Aggregate, dimensions []Dimension, gaps []GapCode) Approval {
	t.Helper()
	approval, err := NewApproval(ApprovalSpec{
		ID: opaque("d"), DefinitionID: "women.safety.v1", DefinitionVersion: 3,
		AggregateID: aggregate.Spec().ID, AggregateVersion: aggregate.Spec().Version,
		ReviewerKey: opaque("e"), Decision: DecisionApprove,
		ReviewedDimensions: dimensions, AcknowledgedGaps: gaps,
		ReviewedAt: testNow.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	return approval
}

func TestRepresentativeEvidenceProducesNeutralReadiness(t *testing.T) {
	aggregate := validAggregate(t)
	assessment, err := Evaluate(
		opaque("f"), validDefinition(t), aggregate,
		approvalFor(t, aggregate, []Dimension{DimensionPrivacyControl, DimensionHarassmentResponse}, nil),
		testNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !assessment.Representative || assessment.Outcome != OutcomeReadyForReleaseReview ||
		len(assessment.Gaps) != 0 {
		t.Fatalf("unexpected assessment: %+v", assessment)
	}
}

func TestInsufficientEvidenceCannotClaimRepresentativeness(t *testing.T) {
	aggregate, err := NewAggregate(AggregateSpec{
		ID: opaque("b"), Version: 7, CohortKey: opaque("c"),
		EligibleCount: 90, RespondentCount: 40,
		Evidence: []DimensionEvidence{
			{Dimension: DimensionHarassmentResponse, EvidenceCount: 30, Favorable: 20},
		},
		WindowStartedAt: testNow.Add(-30 * 24 * time.Hour),
		WindowEndedAt:   testNow.Add(-time.Hour), RedactionVersion: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	approval := approvalFor(t, aggregate,
		[]Dimension{DimensionPrivacyControl, DimensionHarassmentResponse},
		[]GapCode{GapCohortBelowMinimum, GapResponseBelowMinimum, GapDimensionMissing, GapEvidenceBelowMinimum},
	)
	assessment, err := Evaluate(opaque("f"), validDefinition(t), aggregate, approval, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Representative || assessment.Outcome != OutcomeEvidenceIncomplete {
		t.Fatalf("insufficient evidence was overstated: %+v", assessment)
	}
	if len(assessment.Gaps) < 3 {
		t.Fatalf("expected documented gaps, got %+v", assessment.Gaps)
	}
}

func TestIncompleteReviewCoverageIsNotReady(t *testing.T) {
	aggregate := validAggregate(t)
	assessment, err := Evaluate(
		opaque("f"), validDefinition(t), aggregate,
		approvalFor(t, aggregate, []Dimension{DimensionPrivacyControl}, nil),
		testNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	if assessment.Outcome != OutcomeEvidenceIncomplete || len(assessment.Gaps) != 1 ||
		assessment.Gaps[0].Code != GapReviewerCoverage {
		t.Fatalf("coverage gap not preserved: %+v", assessment)
	}
}

func TestMissingGapAcknowledgementIsDocumented(t *testing.T) {
	aggregate, err := NewAggregate(AggregateSpec{
		ID: opaque("b"), Version: 7, CohortKey: opaque("c"),
		EligibleCount: 100, RespondentCount: 50,
		Evidence: []DimensionEvidence{
			{Dimension: DimensionHarassmentResponse, EvidenceCount: 50, Favorable: 30},
			{Dimension: DimensionPrivacyControl, EvidenceCount: 50, Favorable: 30},
		},
		WindowStartedAt: testNow.Add(-time.Hour), WindowEndedAt: testNow,
		RedactionVersion: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	assessment, err := Evaluate(
		opaque("f"), validDefinition(t), aggregate,
		approvalFor(t, aggregate, []Dimension{DimensionPrivacyControl, DimensionHarassmentResponse}, nil),
		testNow.Add(time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, gap := range assessment.Gaps {
		found = found || gap.Code == GapAcknowledgement
	}
	if !found || assessment.Outcome != OutcomeEvidenceIncomplete {
		t.Fatalf("missing acknowledgement not retained: %+v", assessment)
	}
}

func TestDefinitionAggregateAndApprovalRejectUnsafeShapes(t *testing.T) {
	definition := validDefinition(t).Spec()
	definition.MinimumCohort = 2
	if _, err := NewDefinition(definition); err == nil {
		t.Fatal("accepted micro-cohort definition")
	}
	aggregate := validAggregate(t).Spec()
	aggregate.RespondentCount = aggregate.EligibleCount + 1
	if _, err := NewAggregate(aggregate); err == nil {
		t.Fatal("accepted impossible aggregate")
	}
	approval := approvalFor(t, validAggregate(t),
		[]Dimension{DimensionPrivacyControl, DimensionHarassmentResponse}, nil).Spec()
	approval.Decision = "token"
	if _, err := NewApproval(approval); err == nil {
		t.Fatal("accepted token review")
	}
}

func TestPublicJSONHasNoRawIdentityOrScoreFields(t *testing.T) {
	aggregate := validAggregate(t)
	assessment, err := Evaluate(
		opaque("f"), validDefinition(t), aggregate,
		approvalFor(t, aggregate, []Dimension{DimensionPrivacyControl, DimensionHarassmentResponse}, nil),
		testNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(assessment)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(payload))
	for _, forbidden := range []string{"message", "content", "email", "phone", "memberid", "score", "subgroup", "vendor", "model"} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("forbidden field %q in %s", forbidden, payload)
		}
	}
}

func FuzzInsufficientCohortNeverRepresentative(f *testing.F) {
	f.Add(uint32(0), uint32(0))
	f.Add(uint32(99), uint32(99))
	f.Fuzz(func(t *testing.T, eligible, respondents uint32) {
		eligible %= 100
		if eligible == 0 {
			eligible = 1
		}
		respondents %= eligible + 1
		aggregate, err := NewAggregate(AggregateSpec{
			ID: opaque("b"), Version: 7, CohortKey: opaque("c"),
			EligibleCount: eligible, RespondentCount: respondents,
			WindowStartedAt: testNow.Add(-time.Hour), WindowEndedAt: testNow,
			RedactionVersion: 1,
		})
		if err != nil {
			t.Skip()
		}
		approval := approvalFor(t, aggregate,
			[]Dimension{DimensionPrivacyControl, DimensionHarassmentResponse},
			[]GapCode{GapCohortBelowMinimum, GapResponseBelowMinimum, GapDimensionMissing},
		)
		assessment, err := Evaluate(opaque("f"), validDefinition(t), aggregate, approval, testNow.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		if assessment.Representative || assessment.Outcome != OutcomeEvidenceIncomplete {
			t.Fatalf("micro-cohort represented: eligible=%d respondents=%d", eligible, respondents)
		}
	})
}

func TestResponseRateArithmeticDoesNotOverflow(t *testing.T) {
	definition := validDefinition(t).Spec()
	definition.MinimumCohort = math.MaxUint32
	definition.Dimensions[0].MinimumEvidence = 10
	definition.Dimensions[1].MinimumEvidence = 10
	if _, err := NewDefinition(definition); err == nil {
		t.Fatal("definition cap should reject unbounded cohort")
	}
}
