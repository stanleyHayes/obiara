// Package readinesshandoff validates external-readiness metadata. It has no
// capability to handle secrets, submit stores, contact cohorts, or mutate staffing.
package readinesshandoff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"time"
)

const SchemaVersion = "obiara.readiness-handoff.v1"

var (
	hex64   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	roleRef = regexp.MustCompile(`^role/[a-z][a-z0-9-]{2,63}$`)
)

type Evidence struct {
	RequirementID string    `json:"requirementId"`
	Kind          string    `json:"kind"`
	Provenance    string    `json:"provenance"`
	EvidenceRef   string    `json:"evidenceRef"`
	IssuerRef     string    `json:"issuerRef"`
	ReviewerRef   string    `json:"reviewerRef"`
	Outcome       string    `json:"outcome"`
	Synthetic     bool      `json:"synthetic"`
	CollectedAt   time.Time `json:"collectedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
}

type Registry struct {
	SchemaVersion string     `json:"schemaVersion"`
	GeneratedAt   time.Time  `json:"generatedAt"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	Environment   string     `json:"environment"`
	Market        string     `json:"market"`
	Evidence      []Evidence `json:"evidence"`
}

type Requirement struct {
	ID, Kind, Provenance string
}

type Result struct {
	ID       string   `json:"id"`
	Ready    bool     `json:"ready"`
	Blockers []string `json:"blockers"`
}

type Decision struct {
	Ready    bool     `json:"ready"`
	Results  []Result `json:"results"`
	Blockers []string `json:"blockers"`
}

func Requirements() []Requirement {
	return []Requirement{
		{"credential.runtime-account-custody", "credential", "credential-custodian"},
		{"store.play-account-custody", "store", "store-console"},
		{"store.app-account-custody", "store", "store-console"},
		{"store.signing-ceremony", "store", "signing-witness"},
		{"cohort.uat-consent", "cohort", "cohort-review"},
		{"cohort.uat-training", "cohort", "cohort-review"},
		{"cohort.uat-completion", "cohort", "cohort-review"},
		{"operations.circle-coverage", "operations", "operations-review"},
		{"operations.host-readiness", "operations", "operations-review"},
		{"operations.support-coverage", "operations", "operations-review"},
		{"operations.trust-safety-coverage", "operations", "operations-review"},
	}
}

func Load(path string, now time.Time) (Registry, Decision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Registry{}, Decision{}, err
	}
	var registry Registry
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&registry); err != nil {
		return Registry{}, Decision{}, fmt.Errorf("parse readiness handoff: %w", err)
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Registry{}, Decision{}, errors.New("readiness handoff contains trailing data")
	}
	decision, err := Evaluate(registry, now)
	return registry, decision, err
}

func Evaluate(registry Registry, now time.Time) (Decision, error) {
	decision := Decision{}
	if registry.SchemaVersion != SchemaVersion || registry.Environment != "production" || registry.Market != "GH" ||
		now.IsZero() || registry.GeneratedAt.IsZero() || !registry.ExpiresAt.After(registry.GeneratedAt) ||
		registry.ExpiresAt.Sub(registry.GeneratedAt) > 24*time.Hour || now.Before(registry.GeneratedAt.Add(-5*time.Minute)) || now.After(registry.ExpiresAt) {
		return decision, errors.New("invalid or stale readiness handoff")
	}
	known := map[string]Requirement{}
	for _, requirement := range Requirements() {
		known[requirement.ID] = requirement
	}
	byID, refs := map[string]Evidence{}, map[string]bool{}
	for _, evidence := range registry.Evidence {
		requirement, exists := known[evidence.RequirementID]
		if !exists {
			return decision, fmt.Errorf("unknown requirement %q", evidence.RequirementID)
		}
		if _, duplicate := byID[evidence.RequirementID]; duplicate {
			return decision, fmt.Errorf("duplicate requirement %q", evidence.RequirementID)
		}
		if evidence.Kind != requirement.Kind || evidence.Provenance != requirement.Provenance ||
			!hex64.MatchString(evidence.EvidenceRef) || refs[evidence.EvidenceRef] ||
			!roleRef.MatchString(evidence.IssuerRef) || !roleRef.MatchString(evidence.ReviewerRef) || evidence.IssuerRef == evidence.ReviewerRef {
			return decision, fmt.Errorf("invalid evidence metadata for %q", evidence.RequirementID)
		}
		refs[evidence.EvidenceRef], byID[evidence.RequirementID] = true, evidence
	}
	for _, requirement := range Requirements() {
		result := Result{ID: requirement.ID}
		evidence, exists := byID[requirement.ID]
		if !exists {
			result.Blockers = append(result.Blockers, "evidence-missing")
		} else {
			if evidence.Synthetic {
				result.Blockers = append(result.Blockers, "synthetic-evidence")
			}
			if evidence.Outcome != "satisfied" {
				result.Blockers = append(result.Blockers, "outcome-"+evidence.Outcome)
			}
			if evidence.CollectedAt.IsZero() || !evidence.ExpiresAt.After(evidence.CollectedAt) ||
				evidence.ExpiresAt.Sub(evidence.CollectedAt) > 30*24*time.Hour ||
				evidence.CollectedAt.After(now.Add(5*time.Minute)) || now.After(evidence.ExpiresAt) {
				result.Blockers = append(result.Blockers, "evidence-stale")
			}
		}
		result.Ready = len(result.Blockers) == 0
		if !result.Ready {
			decision.Blockers = append(decision.Blockers, requirement.ID)
		}
		decision.Results = append(decision.Results, result)
	}
	decision.Ready = len(decision.Blockers) == 0
	return decision, nil
}

func RequirementIDs() []string {
	ids := make([]string, 0, len(Requirements()))
	for _, requirement := range Requirements() {
		ids = append(ids, requirement.ID)
	}
	slices.Sort(ids)
	return ids
}
