package securityclosure

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	policySchema   = "obiara.security-closure.v1"
	evidenceSchema = "obiara.penetration-evidence.v1"
	localTarget    = "http://127.0.0.1:18080"
)

var opaqueRef = regexp.MustCompile(`^[a-z][a-z0-9-]*:[a-zA-Z0-9][a-zA-Z0-9:._-]*$`)

type Policy struct {
	SchemaVersion string `yaml:"schemaVersion"`
	DAST          struct {
		Target                 string `yaml:"target"`
		Mode                   string `yaml:"mode"`
		MaximumMinutes         int    `yaml:"maximumMinutes"`
		Image                  string `yaml:"image"`
		ImageVersion           string `yaml:"imageVersion"`
		AllowExternalTargets   bool   `yaml:"allowExternalTargets"`
		AllowProductionTargets bool   `yaml:"allowProductionTargets"`
	} `yaml:"dast"`
	Penetration struct {
		MaximumEvidenceAgeDays                  int      `yaml:"maximumEvidenceAgeDays"`
		MaximumAcceptanceDays                   int      `yaml:"maximumAcceptanceDays"`
		ProductionRequiresIndependentAssessment bool     `yaml:"productionRequiresIndependentAssessment"`
		ProductionRequiresAllFindingsClosed     bool     `yaml:"productionRequiresAllFindingsClosed"`
		ForbiddenEvidenceFields                 []string `yaml:"forbiddenEvidenceFields"`
	} `yaml:"penetration"`
}

type Evidence struct {
	SchemaVersion  string    `json:"schemaVersion"`
	AssessmentID   string    `json:"assessmentId"`
	AssessmentKind string    `json:"assessmentKind"`
	ScopeRefs      []string  `json:"scopeRefs"`
	AssessorRef    string    `json:"assessorRef"`
	StartedAt      time.Time `json:"startedAt"`
	CompletedAt    time.Time `json:"completedAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
	Findings       []Finding `json:"findings"`
}

type Finding struct {
	FindingID      string    `json:"findingId"`
	Severity       string    `json:"severity"`
	Status         string    `json:"status"`
	OwnerRef       string    `json:"ownerRef"`
	RemediationRef string    `json:"remediationRef"`
	RetestRef      string    `json:"retestRef"`
	ClosedAt       time.Time `json:"closedAt"`
	VerifierRef    string    `json:"verifierRef"`
}

type Decision struct {
	ProductionEligible bool
	Blockers           []string
}

func LoadPolicy(path string) (Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	var policy Policy
	if err := yaml.Unmarshal(raw, &policy); err != nil {
		return Policy{}, fmt.Errorf("parse security policy: %w", err)
	}
	return policy, ValidatePolicy(policy)
}

func LoadEvidence(path string, policy Policy, now time.Time) (Evidence, Decision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Evidence{}, Decision{}, err
	}
	for _, forbidden := range policy.Penetration.ForbiddenEvidenceFields {
		if strings.Contains(strings.ToLower(string(raw)), `"`+strings.ToLower(forbidden)+`"`) {
			return Evidence{}, Decision{}, fmt.Errorf("forbidden evidence field %q", forbidden)
		}
	}
	var evidence Evidence
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, Decision{}, fmt.Errorf("parse penetration evidence: %w", err)
	}
	decision, err := Evaluate(policy, evidence, now)
	return evidence, decision, err
}

func ValidatePolicy(policy Policy) error {
	if policy.SchemaVersion != policySchema ||
		policy.DAST.Target != localTarget ||
		policy.DAST.Mode != "passive-baseline" ||
		policy.DAST.MaximumMinutes < 1 || policy.DAST.MaximumMinutes > 3 ||
		policy.DAST.AllowExternalTargets || policy.DAST.AllowProductionTargets {
		return errors.New("DAST must be bounded to the passive localhost fixture")
	}
	if !strings.HasPrefix(policy.DAST.Image, "ghcr.io/zaproxy/zaproxy@sha256:") ||
		len(strings.TrimPrefix(policy.DAST.Image, "ghcr.io/zaproxy/zaproxy@sha256:")) != 64 ||
		policy.DAST.ImageVersion == "" {
		return errors.New("ZAP image must have a version and immutable digest")
	}
	if policy.Penetration.MaximumEvidenceAgeDays < 1 ||
		policy.Penetration.MaximumEvidenceAgeDays > 90 ||
		policy.Penetration.MaximumAcceptanceDays < 1 ||
		policy.Penetration.MaximumAcceptanceDays > 30 ||
		!policy.Penetration.ProductionRequiresIndependentAssessment ||
		!policy.Penetration.ProductionRequiresAllFindingsClosed ||
		len(policy.Penetration.ForbiddenEvidenceFields) < 6 {
		return errors.New("penetration closure policy is incomplete")
	}
	return nil
}

func Evaluate(policy Policy, evidence Evidence, now time.Time) (Decision, error) {
	if err := ValidatePolicy(policy); err != nil {
		return Decision{}, err
	}
	if evidence.SchemaVersion != evidenceSchema ||
		!validRef(evidence.AssessmentID) || !validRef(evidence.AssessorRef) {
		return Decision{}, errors.New("invalid assessment identity")
	}
	if evidence.AssessmentKind != "synthetic-ci-exercise" &&
		evidence.AssessmentKind != "independent-penetration-test" {
		return Decision{}, errors.New("invalid assessment kind")
	}
	if evidence.StartedAt.IsZero() || evidence.CompletedAt.Before(evidence.StartedAt) ||
		evidence.ExpiresAt.Before(evidence.CompletedAt) ||
		evidence.ExpiresAt.After(evidence.CompletedAt.AddDate(0, 0, policy.Penetration.MaximumEvidenceAgeDays)) {
		return Decision{}, errors.New("invalid assessment time window")
	}
	if len(evidence.ScopeRefs) == 0 {
		return Decision{}, errors.New("assessment scope is empty")
	}
	scope := make(map[string]struct{}, len(evidence.ScopeRefs))
	for _, ref := range evidence.ScopeRefs {
		if !validRef(ref) {
			return Decision{}, fmt.Errorf("invalid scope reference %q", ref)
		}
		if _, exists := scope[ref]; exists {
			return Decision{}, fmt.Errorf("duplicate scope reference %q", ref)
		}
		scope[ref] = struct{}{}
	}

	blockers := make([]string, 0)
	if evidence.AssessmentKind != "independent-penetration-test" {
		blockers = append(blockers, "independent penetration assessment is required")
	}
	if now.After(evidence.ExpiresAt) {
		blockers = append(blockers, "assessment evidence is expired")
	}
	for _, required := range []string{"service:api", "service:web", "service:admin", "service:mobile"} {
		if _, exists := scope[required]; !exists {
			blockers = append(blockers, "missing production scope "+required)
		}
	}

	seen := make(map[string]struct{}, len(evidence.Findings))
	for _, finding := range evidence.Findings {
		if err := validateFinding(finding, evidence); err != nil {
			return Decision{}, err
		}
		if _, exists := seen[finding.FindingID]; exists {
			return Decision{}, fmt.Errorf("duplicate finding %q", finding.FindingID)
		}
		seen[finding.FindingID] = struct{}{}
		if finding.Status != "closed" {
			blockers = append(blockers, "unclosed finding "+finding.FindingID)
		}
	}
	return Decision{ProductionEligible: len(blockers) == 0, Blockers: blockers}, nil
}

func validateFinding(finding Finding, evidence Evidence) error {
	if !validRef(finding.FindingID) || !validRef(finding.OwnerRef) {
		return errors.New("finding identity or owner is invalid")
	}
	switch finding.Severity {
	case "critical", "high", "medium", "low":
	default:
		return fmt.Errorf("finding %q has invalid severity", finding.FindingID)
	}
	switch finding.Status {
	case "open", "in-progress":
		if finding.RemediationRef != "" || finding.RetestRef != "" ||
			!finding.ClosedAt.IsZero() || finding.VerifierRef != "" {
			return fmt.Errorf("unclosed finding %q claims closure evidence", finding.FindingID)
		}
	case "closed":
		if !validRef(finding.RemediationRef) || !validRef(finding.RetestRef) ||
			!validRef(finding.VerifierRef) || finding.ClosedAt.Before(evidence.CompletedAt) ||
			finding.VerifierRef == evidence.AssessorRef ||
			finding.VerifierRef == finding.OwnerRef ||
			evidence.AssessorRef == finding.OwnerRef {
			return fmt.Errorf("finding %q lacks distinct verified closure", finding.FindingID)
		}
	default:
		return fmt.Errorf("finding %q has invalid status", finding.FindingID)
	}
	return nil
}

func validRef(value string) bool {
	return opaqueRef.MatchString(value) && len(value) <= 160
}
