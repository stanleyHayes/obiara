// Package launchgates evaluates metadata-only production evidence. It never
// provisions, approves, submits, deploys or activates a production resource.
package launchgates

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

const SchemaVersion = "obiara.launch-gates.v1"

var (
	sha40   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	roleRef = regexp.MustCompile(`^role/[a-z][a-z0-9-]{2,63}$`)
)

type EvidenceKind string

const (
	Repository       EvidenceKind = "repository"
	ExternalDecision EvidenceKind = "external-decision"
	Provider         EvidenceKind = "provider"
	Credential       EvidenceKind = "credential"
	Cohort           EvidenceKind = "cohort"
	Store            EvidenceKind = "store"
	ProductionAction EvidenceKind = "production-action"
)

type Provenance string

const (
	RepositoryControl    Provenance = "repository-control"
	ExternalAuthority    Provenance = "external-authority"
	ProviderControlPlane Provenance = "provider-control-plane"
	CredentialCustodian  Provenance = "credential-custodian"
	CohortReview         Provenance = "cohort-review"
	StoreConsole         Provenance = "store-console"
	ChangeAuthority      Provenance = "change-authority"
)

type Outcome string

const (
	Satisfied Outcome = "satisfied"
	Pending   Outcome = "pending"
	Blocked   Outcome = "blocked"
)

type Evidence struct {
	GateID       string       `json:"gateId"`
	EvidenceRef  string       `json:"evidenceRef"`
	Kind         EvidenceKind `json:"kind"`
	Provenance   Provenance   `json:"provenance"`
	Environment  string       `json:"environment"`
	Market       string       `json:"market"`
	CandidateSHA string       `json:"candidateSha"`
	IssuerRef    string       `json:"issuerRef"`
	ReviewerRef  string       `json:"reviewerRef"`
	Outcome      Outcome      `json:"outcome"`
	Synthetic    bool         `json:"synthetic"`
	CollectedAt  time.Time    `json:"collectedAt"`
	ExpiresAt    time.Time    `json:"expiresAt"`
}
type Registry struct {
	SchemaVersion string     `json:"schemaVersion"`
	Environment   string     `json:"environment"`
	Market        string     `json:"market"`
	CandidateSHA  string     `json:"candidateSha"`
	GeneratedAt   time.Time  `json:"generatedAt"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	Evidence      []Evidence `json:"evidence"`
}
type Requirement struct {
	ID         string
	Kind       EvidenceKind
	Provenance Provenance
	MaxAge     time.Duration
	DependsOn  []string
}
type GateResult struct {
	ID        string       `json:"id"`
	Kind      EvidenceKind `json:"kind"`
	Satisfied bool         `json:"satisfied"`
	Blockers  []string     `json:"blockers"`
}
type Decision struct {
	Ready        bool         `json:"ready"`
	Environment  string       `json:"environment"`
	CandidateSHA string       `json:"candidateSha"`
	Gates        []GateResult `json:"gates"`
	Blockers     []string     `json:"blockers"`
}

func Requirements() []Requirement {
	day := 24 * time.Hour
	return []Requirement{
		{"repository.engineering", Repository, RepositoryControl, day, nil},
		{"repository.security", Repository, RepositoryControl, day, nil},
		{"repository.recovery", Repository, RepositoryControl, 7 * day, nil},
		{"repository.release", Repository, RepositoryControl, day, []string{"repository.engineering", "repository.security", "repository.recovery"}},
		{"external.residency", ExternalDecision, ExternalAuthority, 90 * day, nil},
		{"external.production-topology", ExternalDecision, ExternalAuthority, 30 * day, []string{"external.residency"}},
		{"external.ghana-device-network", ExternalDecision, ExternalAuthority, 30 * day, nil},
		{"provider.atlas", Provider, ProviderControlPlane, 30 * day, []string{"external.residency", "external.production-topology"}},
		{"provider.storage-cdn", Provider, ProviderControlPlane, 30 * day, []string{"external.residency", "external.production-topology"}},
		{"provider.live-media", Provider, ProviderControlPlane, 30 * day, []string{"external.residency", "external.production-topology"}},
		{"provider.email", Provider, ProviderControlPlane, 30 * day, []string{"external.residency", "external.production-topology"}},
		{"credential.production-runtime", Credential, CredentialCustodian, day, []string{"provider.atlas", "provider.storage-cdn", "provider.live-media", "provider.email"}},
		{"cohort.uat", Cohort, CohortReview, 7 * day, []string{"repository.engineering", "repository.security", "external.ghana-device-network"}},
		{"cohort.launch-operations", Cohort, CohortReview, 7 * day, []string{"repository.recovery", "cohort.uat"}},
		{"store.mobile-release", Store, StoreConsole, day, []string{"credential.production-runtime", "cohort.uat"}},
		{"external.founder-decision", ExternalDecision, ExternalAuthority, day, []string{"repository.release", "credential.production-runtime", "cohort.launch-operations", "store.mobile-release"}},
		{"production.activation", ProductionAction, ChangeAuthority, 4 * time.Hour, []string{"external.founder-decision", "repository.release", "credential.production-runtime", "cohort.launch-operations", "store.mobile-release"}},
	}
}

func Load(path string, now time.Time) (Registry, Decision, error) {
	raw, e := os.ReadFile(path)
	if e != nil {
		return Registry{}, Decision{}, e
	}
	var registry Registry
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if e = decoder.Decode(&registry); e != nil {
		return Registry{}, Decision{}, fmt.Errorf("parse launch gate registry: %w", e)
	}
	if e = ensureEOF(decoder); e != nil {
		return Registry{}, Decision{}, e
	}
	decision, e := Evaluate(registry, now)
	return registry, decision, e
}
func Evaluate(registry Registry, now time.Time) (Decision, error) {
	decision := Decision{Environment: registry.Environment, CandidateSHA: registry.CandidateSHA}
	if registry.SchemaVersion != SchemaVersion || registry.Environment != "production" || registry.Market != "GH" || !sha40.MatchString(registry.CandidateSHA) || now.IsZero() ||
		registry.GeneratedAt.IsZero() || !registry.ExpiresAt.After(registry.GeneratedAt) || registry.ExpiresAt.Sub(registry.GeneratedAt) > 24*time.Hour || now.Before(registry.GeneratedAt.Add(-5*time.Minute)) || now.After(registry.ExpiresAt) {
		return decision, errors.New("invalid or stale production registry")
	}
	requirements := Requirements()
	byGate := make(map[string]Evidence, len(registry.Evidence))
	seenRefs := make(map[string]bool, len(registry.Evidence))
	for _, evidence := range registry.Evidence {
		if _, exists := byGate[evidence.GateID]; exists {
			return decision, fmt.Errorf("duplicate evidence for %q", evidence.GateID)
		}
		if seenRefs[evidence.EvidenceRef] {
			return decision, errors.New("duplicate evidence reference")
		}
		if !validEvidenceShape(evidence, registry) {
			return decision, fmt.Errorf("invalid evidence metadata for %q", evidence.GateID)
		}
		seenRefs[evidence.EvidenceRef] = true
		byGate[evidence.GateID] = evidence
	}
	known := map[string]bool{}
	for _, requirement := range requirements {
		known[requirement.ID] = true
	}
	for id := range byGate {
		if !known[id] {
			return decision, fmt.Errorf("unknown production gate %q", id)
		}
	}
	satisfied := map[string]bool{}
	for _, requirement := range requirements {
		result := GateResult{ID: requirement.ID, Kind: requirement.Kind}
		evidence, exists := byGate[requirement.ID]
		if !exists {
			result.Blockers = append(result.Blockers, "evidence-missing")
		} else {
			if evidence.Kind != requirement.Kind {
				result.Blockers = append(result.Blockers, "evidence-kind-mismatch")
			}
			if evidence.Provenance != requirement.Provenance {
				result.Blockers = append(result.Blockers, "provenance-mismatch")
			}
			if evidence.Synthetic {
				result.Blockers = append(result.Blockers, "synthetic-evidence")
			}
			if evidence.Outcome != Satisfied {
				result.Blockers = append(result.Blockers, "outcome-"+string(evidence.Outcome))
			}
			if evidence.CollectedAt.After(now.Add(5*time.Minute)) || now.Sub(evidence.CollectedAt) > requirement.MaxAge || evidence.ExpiresAt.Sub(evidence.CollectedAt) > requirement.MaxAge || now.After(evidence.ExpiresAt) {
				result.Blockers = append(result.Blockers, "evidence-stale")
			}
		}
		for _, dependency := range requirement.DependsOn {
			if !satisfied[dependency] {
				result.Blockers = append(result.Blockers, "dependency-"+dependency)
			}
		}
		result.Satisfied = len(result.Blockers) == 0
		satisfied[requirement.ID] = result.Satisfied
		if !result.Satisfied {
			decision.Blockers = append(decision.Blockers, requirement.ID)
		}
		decision.Gates = append(decision.Gates, result)
	}
	decision.Ready = len(decision.Blockers) == 0
	return decision, nil
}
func validEvidenceShape(e Evidence, r Registry) bool {
	if e.GateID == "" || !hex64.MatchString(e.EvidenceRef) || e.Environment != "production" || e.Market != "GH" || e.CandidateSHA != r.CandidateSHA ||
		!roleRef.MatchString(e.IssuerRef) || !roleRef.MatchString(e.ReviewerRef) || e.IssuerRef == e.ReviewerRef || e.CollectedAt.IsZero() || !e.ExpiresAt.After(e.CollectedAt) ||
		(e.Outcome != Satisfied && e.Outcome != Pending && e.Outcome != Blocked) {
		return false
	}
	expected := map[EvidenceKind]Provenance{Repository: RepositoryControl, ExternalDecision: ExternalAuthority, Provider: ProviderControlPlane, Credential: CredentialCustodian, Cohort: CohortReview, Store: StoreConsole, ProductionAction: ChangeAuthority}[e.Kind]
	return expected != "" && e.Provenance == expected && !containsSecretShape(e.IssuerRef) && !containsSecretShape(e.ReviewerRef)
}
func containsSecretShape(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "://") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "password") || strings.Contains(lower, "private")
}
func ensureEOF(decoder *json.Decoder) error {
	var extra any
	e := decoder.Decode(&extra)
	if !errors.Is(e, io.EOF) {
		return errors.New("launch registry contains trailing data")
	}
	return nil
}
func SortedGateIDs() []string {
	ids := make([]string, 0, len(Requirements()))
	for _, r := range Requirements() {
		ids = append(ids, r.ID)
	}
	slices.Sort(ids)
	return ids
}
