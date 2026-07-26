// Package domain evaluates redacted, cohort-level women's-safety evidence.
// Raw content, identities, member scores, subgroup microdata, and enforcement
// actions are deliberately not representable.
package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	MaxDimensions = 12
	MaxGaps       = 24
)

var (
	ErrInvalidDefinition = errors.New("invalid reviewed evidence definition")
	ErrInvalidAggregate  = errors.New("invalid redacted evidence aggregate")
	ErrInvalidApproval   = errors.New("invalid women-reviewer approval")
	ErrInvalidAssessment = errors.New("invalid evidence assessment")
)

var opaquePattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var tokenPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)

type Dimension string

const (
	DimensionHarassmentResponse Dimension = "harassment_response"
	DimensionCoercionResistance Dimension = "coercion_resistance"
	DimensionPrivacyControl     Dimension = "privacy_control"
	DimensionReportingAccess    Dimension = "reporting_access"
	DimensionCareAccess         Dimension = "care_access"
)

type DimensionRule struct {
	Dimension       Dimension
	MinimumEvidence uint32
}

type DefinitionSpec struct {
	ID                      string
	Version                 uint64
	MinimumCohort           uint32
	MinimumResponsePermille uint16
	Dimensions              []DimensionRule
	ReviewID                string
	ReviewedByWomenPanelKey string
	ReviewedAt              time.Time
}

type Definition struct{ spec DefinitionSpec }

func NewDefinition(spec DefinitionSpec) (Definition, error) {
	spec.Dimensions = append([]DimensionRule(nil), spec.Dimensions...)
	spec.ReviewedAt = spec.ReviewedAt.UTC()
	if !tokenPattern.MatchString(spec.ID) || spec.Version == 0 ||
		spec.MinimumCohort < 20 || spec.MinimumCohort > 1_000_000 ||
		spec.MinimumResponsePermille < 100 || spec.MinimumResponsePermille > 1000 ||
		len(spec.Dimensions) == 0 || len(spec.Dimensions) > MaxDimensions ||
		!tokenPattern.MatchString(spec.ReviewID) ||
		!opaquePattern.MatchString(spec.ReviewedByWomenPanelKey) ||
		spec.ReviewedAt.IsZero() {
		return Definition{}, ErrInvalidDefinition
	}
	seen := map[Dimension]struct{}{}
	for _, rule := range spec.Dimensions {
		if !validDimension(rule.Dimension) || rule.MinimumEvidence < 10 ||
			rule.MinimumEvidence > spec.MinimumCohort {
			return Definition{}, ErrInvalidDefinition
		}
		if _, exists := seen[rule.Dimension]; exists {
			return Definition{}, ErrInvalidDefinition
		}
		seen[rule.Dimension] = struct{}{}
	}
	slices.SortFunc(spec.Dimensions, func(a, b DimensionRule) int {
		return strings.Compare(string(a.Dimension), string(b.Dimension))
	})
	return Definition{spec: spec}, nil
}

func (definition Definition) Spec() DefinitionSpec {
	spec := definition.spec
	spec.Dimensions = append([]DimensionRule(nil), spec.Dimensions...)
	return spec
}

type DimensionEvidence struct {
	Dimension     Dimension
	EvidenceCount uint32
	Favorable     uint32
}

// Aggregate is a bounded, redacted cohort summary. It cannot carry rows,
// subgroup labels, free text, identities, or content.
type AggregateSpec struct {
	ID               string
	Version          uint64
	CohortKey        string
	EligibleCount    uint32
	RespondentCount  uint32
	Evidence         []DimensionEvidence
	WindowStartedAt  time.Time
	WindowEndedAt    time.Time
	RedactionVersion uint64
}

type Aggregate struct{ spec AggregateSpec }

func NewAggregate(spec AggregateSpec) (Aggregate, error) {
	spec.Evidence = append([]DimensionEvidence(nil), spec.Evidence...)
	spec.WindowStartedAt = spec.WindowStartedAt.UTC()
	spec.WindowEndedAt = spec.WindowEndedAt.UTC()
	if !opaquePattern.MatchString(spec.ID) || spec.Version == 0 ||
		!opaquePattern.MatchString(spec.CohortKey) || spec.EligibleCount == 0 ||
		spec.RespondentCount > spec.EligibleCount || len(spec.Evidence) > MaxDimensions ||
		spec.WindowStartedAt.IsZero() || !spec.WindowEndedAt.After(spec.WindowStartedAt) ||
		spec.RedactionVersion == 0 {
		return Aggregate{}, ErrInvalidAggregate
	}
	seen := map[Dimension]struct{}{}
	for _, evidence := range spec.Evidence {
		if !validDimension(evidence.Dimension) ||
			evidence.EvidenceCount > spec.RespondentCount ||
			evidence.Favorable > evidence.EvidenceCount {
			return Aggregate{}, ErrInvalidAggregate
		}
		if _, exists := seen[evidence.Dimension]; exists {
			return Aggregate{}, ErrInvalidAggregate
		}
		seen[evidence.Dimension] = struct{}{}
	}
	slices.SortFunc(spec.Evidence, func(a, b DimensionEvidence) int {
		return strings.Compare(string(a.Dimension), string(b.Dimension))
	})
	return Aggregate{spec: spec}, nil
}

func (aggregate Aggregate) Spec() AggregateSpec {
	spec := aggregate.spec
	spec.Evidence = append([]DimensionEvidence(nil), aggregate.spec.Evidence...)
	return spec
}

type ReviewDecision string

const DecisionApprove ReviewDecision = "approve"

type ApprovalSpec struct {
	ID                 string
	DefinitionID       string
	DefinitionVersion  uint64
	AggregateID        string
	AggregateVersion   uint64
	ReviewerKey        string
	Decision           ReviewDecision
	ReviewedDimensions []Dimension
	AcknowledgedGaps   []GapCode
	ReviewedAt         time.Time
}

type Approval struct{ spec ApprovalSpec }

func NewApproval(spec ApprovalSpec) (Approval, error) {
	spec.ReviewedDimensions = append([]Dimension(nil), spec.ReviewedDimensions...)
	spec.AcknowledgedGaps = append([]GapCode(nil), spec.AcknowledgedGaps...)
	spec.ReviewedAt = spec.ReviewedAt.UTC()
	if !opaquePattern.MatchString(spec.ID) || !tokenPattern.MatchString(spec.DefinitionID) ||
		spec.DefinitionVersion == 0 || !opaquePattern.MatchString(spec.AggregateID) ||
		spec.AggregateVersion == 0 || !opaquePattern.MatchString(spec.ReviewerKey) ||
		spec.Decision != DecisionApprove || spec.ReviewedAt.IsZero() ||
		len(spec.ReviewedDimensions) == 0 || len(spec.ReviewedDimensions) > MaxDimensions ||
		len(spec.AcknowledgedGaps) > MaxGaps {
		return Approval{}, ErrInvalidApproval
	}
	if !uniqueDimensions(spec.ReviewedDimensions) || !uniqueGaps(spec.AcknowledgedGaps) {
		return Approval{}, ErrInvalidApproval
	}
	slices.Sort(spec.ReviewedDimensions)
	slices.Sort(spec.AcknowledgedGaps)
	return Approval{spec: spec}, nil
}

func (approval Approval) Spec() ApprovalSpec {
	spec := approval.spec
	spec.ReviewedDimensions = append([]Dimension(nil), approval.spec.ReviewedDimensions...)
	spec.AcknowledgedGaps = append([]GapCode(nil), approval.spec.AcknowledgedGaps...)
	return spec
}

type GapCode string

const (
	GapCohortBelowMinimum   GapCode = "cohort_below_minimum"
	GapResponseBelowMinimum GapCode = "response_below_minimum"
	GapDimensionMissing     GapCode = "dimension_missing"
	GapEvidenceBelowMinimum GapCode = "evidence_below_minimum"
	GapReviewerCoverage     GapCode = "reviewer_coverage_incomplete"
	GapAcknowledgement      GapCode = "gap_acknowledgement_incomplete"
)

type Gap struct {
	Code      GapCode
	Dimension Dimension
}

type Outcome string

const (
	OutcomeEvidenceIncomplete    Outcome = "evidence_incomplete"
	OutcomeReadyForReleaseReview Outcome = "ready_for_release_review"
)

type Assessment struct {
	ID                string
	CohortKey         string
	DefinitionID      string
	DefinitionVersion uint64
	AggregateID       string
	AggregateVersion  uint64
	ApprovalID        string
	Outcome           Outcome
	Representative    bool
	Gaps              []Gap
	EvaluatedAt       time.Time
}

func Evaluate(id string, definition Definition, aggregate Aggregate, approval Approval, at time.Time) (Assessment, error) {
	if !opaquePattern.MatchString(id) || at.IsZero() {
		return Assessment{}, ErrInvalidAssessment
	}
	def := definition.Spec()
	agg := aggregate.Spec()
	app := approval.Spec()
	if app.DefinitionID != def.ID || app.DefinitionVersion != def.Version ||
		app.AggregateID != agg.ID || app.AggregateVersion != agg.Version ||
		app.ReviewedAt.After(at.UTC()) {
		return Assessment{}, ErrInvalidAssessment
	}

	gaps := evidenceGaps(def, agg)
	reviewed := make(map[Dimension]struct{}, len(app.ReviewedDimensions))
	for _, dimension := range app.ReviewedDimensions {
		reviewed[dimension] = struct{}{}
	}
	for _, rule := range def.Dimensions {
		if _, ok := reviewed[rule.Dimension]; !ok {
			gaps = append(gaps, Gap{Code: GapReviewerCoverage, Dimension: rule.Dimension})
		}
	}
	requiredGapCodes := map[GapCode]struct{}{}
	for _, gap := range gaps {
		if gap.Code != GapReviewerCoverage && gap.Code != GapAcknowledgement {
			requiredGapCodes[gap.Code] = struct{}{}
		}
	}
	acknowledged := map[GapCode]struct{}{}
	for _, code := range app.AcknowledgedGaps {
		acknowledged[code] = struct{}{}
	}
	for code := range requiredGapCodes {
		if _, ok := acknowledged[code]; !ok {
			gaps = append(gaps, Gap{Code: GapAcknowledgement})
			break
		}
	}
	slices.SortFunc(gaps, func(a, b Gap) int {
		if comparison := strings.Compare(string(a.Code), string(b.Code)); comparison != 0 {
			return comparison
		}
		return strings.Compare(string(a.Dimension), string(b.Dimension))
	})
	representative := len(evidenceGaps(def, agg)) == 0
	outcome := OutcomeEvidenceIncomplete
	if representative && len(gaps) == 0 {
		outcome = OutcomeReadyForReleaseReview
	}
	return Assessment{
		ID: id, CohortKey: agg.CohortKey, DefinitionID: def.ID,
		DefinitionVersion: def.Version, AggregateID: agg.ID,
		AggregateVersion: agg.Version, ApprovalID: app.ID, Outcome: outcome,
		Representative: representative, Gaps: gaps, EvaluatedAt: at.UTC(),
	}, nil
}

func evidenceGaps(def DefinitionSpec, agg AggregateSpec) []Gap {
	var gaps []Gap
	if agg.EligibleCount < def.MinimumCohort {
		gaps = append(gaps, Gap{Code: GapCohortBelowMinimum})
	}
	responsePermille := uint32(0)
	if agg.EligibleCount > 0 {
		responsePermille = uint32(agg.RespondentCount) * 1000 / uint32(agg.EligibleCount)
	}
	if responsePermille < uint32(def.MinimumResponsePermille) {
		gaps = append(gaps, Gap{Code: GapResponseBelowMinimum})
	}
	byDimension := map[Dimension]DimensionEvidence{}
	for _, evidence := range agg.Evidence {
		byDimension[evidence.Dimension] = evidence
	}
	for _, rule := range def.Dimensions {
		evidence, ok := byDimension[rule.Dimension]
		if !ok {
			gaps = append(gaps, Gap{Code: GapDimensionMissing, Dimension: rule.Dimension})
		} else if evidence.EvidenceCount < rule.MinimumEvidence {
			gaps = append(gaps, Gap{Code: GapEvidenceBelowMinimum, Dimension: rule.Dimension})
		}
	}
	return gaps
}

func validDimension(d Dimension) bool {
	return slices.Contains([]Dimension{
		DimensionHarassmentResponse, DimensionCoercionResistance,
		DimensionPrivacyControl, DimensionReportingAccess, DimensionCareAccess,
	}, d)
}

func uniqueDimensions(values []Dimension) bool {
	seen := map[Dimension]struct{}{}
	for _, value := range values {
		if !validDimension(value) {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func uniqueGaps(values []GapCode) bool {
	seen := map[GapCode]struct{}{}
	for _, value := range values {
		switch value {
		case GapCohortBelowMinimum, GapResponseBelowMinimum, GapDimensionMissing, GapEvidenceBelowMinimum:
		default:
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
