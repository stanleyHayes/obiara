// Package residencydecision validates metadata for an external residency and
// DPIA decision. It never supplies legal advice, signs, approves or activates
// production.
package residencydecision

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

const SchemaVersion = "obiara.residency-dpia-decision.v1"

var (
	sha40       = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64       = regexp.MustCompile(`^[0-9a-f]{64}$`)
	decisionRef = regexp.MustCompile(`^decision/[a-z0-9][a-z0-9-]{7,79}$`)
	roleRef     = regexp.MustCompile(`^role/[a-z][a-z0-9-]{2,63}$`)
	providerRef = regexp.MustCompile(`^provider/[a-z][a-z0-9-]{2,63}$`)
	regionRef   = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,63}$`)
	countryCode = regexp.MustCompile(`^[A-Z]{2}$`)
)

type Authority struct {
	ActorRef             string     `json:"actorRef"`
	SignatureEvidenceRef string     `json:"signatureEvidenceRef"`
	SignedAt             *time.Time `json:"signedAt"`
}

type AuthorityRefs struct {
	Founder             Authority `json:"founder"`
	DPOLegal            Authority `json:"dpoLegal"`
	IndependentReviewer Authority `json:"independentReviewer"`
}

type ScopeLocation struct {
	Boundary    string   `json:"boundary"`
	ProviderRef string   `json:"providerRef"`
	CountryCode string   `json:"countryCode"`
	RegionCode  string   `json:"regionCode"`
	DataClasses []string `json:"dataClasses"`
}

type Record struct {
	SchemaVersion         string          `json:"schemaVersion"`
	DecisionID            string          `json:"decisionId"`
	Environment           string          `json:"environment"`
	Market                string          `json:"market"`
	CandidateSHA          string          `json:"candidateSha"`
	Interpretation        string          `json:"interpretation"`
	IssuerKind            string          `json:"issuerKind"`
	IssuerRef             string          `json:"issuerRef"`
	AuthorityRefs         AuthorityRefs   `json:"authorityRefs"`
	ProcessingScopeRefs   []string        `json:"processingScopeRefs"`
	ScopeLocations        []ScopeLocation `json:"scopeLocations"`
	TransferAssessmentRef string          `json:"transferAssessmentRef"`
	DPIARef               string          `json:"dpiaRef"`
	ResidualRiskOutcome   string          `json:"residualRiskOutcome"`
	IssuedAt              *time.Time      `json:"issuedAt"`
	ExpiresAt             *time.Time      `json:"expiresAt"`
	Synthetic             bool            `json:"synthetic"`
	Outcome               string          `json:"outcome"`
}

type Decision struct {
	Eligible bool     `json:"eligible"`
	Blockers []string `json:"blockers"`
}

func ProcessingScopes() []string {
	return []string{
		"identity-and-authentication",
		"biometric-liveness",
		"profile-and-matching",
		"private-voice-and-media",
		"community-and-trust",
		"safety-care-and-evidence",
		"ai-counsel-and-language",
		"commerce-and-financial-ledger",
		"analytics-audit-and-operations",
	}
}

func LocationBoundaries() []string {
	return []string{
		"compute",
		"operational-database",
		"database-backups",
		"object-storage",
		"cdn-cache",
		"logs-observability",
		"live-media-network",
		"live-media-recording-egress",
		"transactional-email",
		"ai-vendor-processing",
		"identity-provider-processing",
		"provider-support-access",
	}
}

func Load(path string, now time.Time) (Record, Decision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Record{}, Decision{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var record Record
	if err := decoder.Decode(&record); err != nil {
		return Record{}, Decision{}, fmt.Errorf("parse residency decision: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Record{}, Decision{}, errors.New("residency decision contains trailing data")
	}
	decision, err := Evaluate(record, now)
	return record, decision, err
}

func Evaluate(record Record, now time.Time) (Decision, error) {
	if record.SchemaVersion != SchemaVersion ||
		!decisionRef.MatchString(record.DecisionID) ||
		record.Environment != "production" ||
		record.Market != "GH" ||
		!sha40.MatchString(record.CandidateSHA) ||
		(record.Interpretation != "ghana-only" && record.Interpretation != "africa-region") ||
		(record.IssuerKind != "external-authority" && record.IssuerKind != "repository") ||
		!safeRole(record.IssuerRef) ||
		(record.ResidualRiskOutcome != "pending" &&
			record.ResidualRiskOutcome != "accepted" &&
			record.ResidualRiskOutcome != "declined") ||
		(record.Outcome != "pending" && record.Outcome != "approved" && record.Outcome != "blocked") ||
		now.IsZero() {
		return Decision{}, errors.New("invalid residency decision metadata")
	}

	blockers := make([]string, 0)
	if record.IssuerKind != "external-authority" {
		blockers = append(blockers, "repository-issued")
	}
	if record.Synthetic {
		blockers = append(blockers, "synthetic")
	}
	if record.Outcome != "approved" {
		blockers = append(blockers, "outcome-"+record.Outcome)
	}
	if record.ResidualRiskOutcome != "accepted" {
		blockers = append(blockers, "residual-risk-"+record.ResidualRiskOutcome)
	}
	if !hex64.MatchString(record.TransferAssessmentRef) {
		blockers = append(blockers, "transfer-assessment-missing")
	}
	if !hex64.MatchString(record.DPIARef) {
		blockers = append(blockers, "dpia-missing")
	}
	if record.IssuedAt == nil || record.ExpiresAt == nil {
		blockers = append(blockers, "decision-unsigned-or-undated")
	} else if record.ExpiresAt.Before(*record.IssuedAt) ||
		record.ExpiresAt.Sub(*record.IssuedAt) > 90*24*time.Hour ||
		now.Before(record.IssuedAt.Add(-5*time.Minute)) ||
		now.After(*record.ExpiresAt) {
		blockers = append(blockers, "decision-stale")
	}

	authorityBlockers, err := validateAuthorities(record.AuthorityRefs, record.IssuerRef, record.IssuedAt, record.ExpiresAt)
	if err != nil {
		return Decision{}, err
	}
	blockers = append(blockers, authorityBlockers...)
	if !uniqueEvidenceRefs(record) {
		blockers = append(blockers, "evidence-reference-replayed")
	}
	blockers = append(blockers, validateExactSet(record.ProcessingScopeRefs, ProcessingScopes(), "processing-scope")...)

	locationBlockers, err := validateLocations(record.ScopeLocations, record.Interpretation)
	if err != nil {
		return Decision{}, err
	}
	blockers = append(blockers, locationBlockers...)
	return Decision{Eligible: len(blockers) == 0, Blockers: blockers}, nil
}

func validateAuthorities(refs AuthorityRefs, issuer string, issuedAt, expiresAt *time.Time) ([]string, error) {
	authorities := []struct {
		name string
		ref  Authority
	}{
		{"founder", refs.Founder},
		{"dpo-legal", refs.DPOLegal},
		{"independent-reviewer", refs.IndependentReviewer},
	}
	blockers := make([]string, 0)
	actors := map[string]bool{issuer: true}
	signatures := make(map[string]bool)
	for _, authority := range authorities {
		if authority.ref.ActorRef == "" && authority.ref.SignatureEvidenceRef == "" && authority.ref.SignedAt == nil {
			blockers = append(blockers, authority.name+"-unsigned")
			continue
		}
		if !safeRole(authority.ref.ActorRef) || !hex64.MatchString(authority.ref.SignatureEvidenceRef) {
			return nil, fmt.Errorf("%s authority metadata is invalid", authority.name)
		}
		if actors[authority.ref.ActorRef] {
			blockers = append(blockers, authority.name+"-self-approved")
		}
		if signatures[authority.ref.SignatureEvidenceRef] {
			blockers = append(blockers, authority.name+"-signature-replayed")
		}
		actors[authority.ref.ActorRef] = true
		signatures[authority.ref.SignatureEvidenceRef] = true
		if authority.ref.SignedAt == nil {
			blockers = append(blockers, authority.name+"-unsigned")
		} else if issuedAt == nil || expiresAt == nil ||
			authority.ref.SignedAt.Before(*issuedAt) ||
			authority.ref.SignedAt.After(*expiresAt) {
			blockers = append(blockers, authority.name+"-signature-outside-decision")
		}
	}
	return blockers, nil
}

func validateLocations(locations []ScopeLocation, interpretation string) ([]string, error) {
	required := LocationBoundaries()
	seen := make(map[string]bool, len(locations))
	blockers := make([]string, 0)
	for _, location := range locations {
		if seen[location.Boundary] {
			return nil, fmt.Errorf("duplicate location boundary %q", location.Boundary)
		}
		seen[location.Boundary] = true
		if !slices.Contains(required, location.Boundary) {
			return nil, fmt.Errorf("unknown location boundary %q", location.Boundary)
		}
		if !providerRef.MatchString(location.ProviderRef) ||
			!countryCode.MatchString(location.CountryCode) ||
			location.CountryCode == "ZZ" ||
			!regionRef.MatchString(location.RegionCode) ||
			containsUnsafeShape(location.ProviderRef) ||
			isAmbiguousRegion(location.RegionCode) {
			blockers = append(blockers, "location-invalid-"+location.Boundary)
		}
		if interpretation == "ghana-only" && location.CountryCode != "GH" {
			blockers = append(blockers, "ghana-only-location-mismatch-"+location.Boundary)
		}
		if interpretation == "africa-region" &&
			requiresAfricanResidency(location.Boundary) &&
			!africanCountryCodes[location.CountryCode] {
			blockers = append(blockers, "africa-region-location-mismatch-"+location.Boundary)
		}
		if len(location.DataClasses) == 0 {
			blockers = append(blockers, "data-classes-missing-"+location.Boundary)
		}
		classes := make(map[string]bool)
		for _, class := range location.DataClasses {
			if class != "C2" && class != "C3" && class != "C4" {
				return nil, fmt.Errorf("invalid data class %q", class)
			}
			if classes[class] {
				return nil, fmt.Errorf("duplicate data class %q", class)
			}
			classes[class] = true
		}
	}
	for _, boundary := range required {
		if !seen[boundary] {
			blockers = append(blockers, "location-missing-"+boundary)
		}
	}
	return blockers, nil
}

func validateExactSet(actual, required []string, prefix string) []string {
	seen := make(map[string]bool, len(actual))
	blockers := make([]string, 0)
	for _, value := range actual {
		if seen[value] || !slices.Contains(required, value) {
			blockers = append(blockers, prefix+"-invalid-"+value)
		}
		seen[value] = true
	}
	for _, value := range required {
		if !seen[value] {
			blockers = append(blockers, prefix+"-missing-"+value)
		}
	}
	return blockers
}

func safeRole(value string) bool {
	return roleRef.MatchString(value) && !containsUnsafeShape(value)
}

func containsUnsafeShape(value string) bool {
	lower := strings.ToLower(value)
	for _, unsafe := range []string{"://", "@", "secret", "token", "password", "private", "member", "email", "phone"} {
		if strings.Contains(lower, unsafe) {
			return true
		}
	}
	return false
}

func isAmbiguousRegion(value string) bool {
	switch strings.ToLower(value) {
	case "africa", "ghana", "global", "automatic", "auto", "nearest", "unknown", "undecided", "multi-region":
		return true
	default:
		return false
	}
}

func uniqueEvidenceRefs(record Record) bool {
	refs := []string{
		record.TransferAssessmentRef,
		record.DPIARef,
		record.AuthorityRefs.Founder.SignatureEvidenceRef,
		record.AuthorityRefs.DPOLegal.SignatureEvidenceRef,
		record.AuthorityRefs.IndependentReviewer.SignatureEvidenceRef,
	}
	seen := make(map[string]bool, len(refs))
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		if seen[ref] {
			return false
		}
		seen[ref] = true
	}
	return true
}

func requiresAfricanResidency(boundary string) bool {
	switch boundary {
	case "compute", "operational-database", "database-backups", "object-storage",
		"cdn-cache", "logs-observability", "live-media-network",
		"live-media-recording-egress":
		return true
	default:
		return false
	}
}

var africanCountryCodes = map[string]bool{
	"AO": true, "BF": true, "BI": true, "BJ": true, "BW": true, "CD": true,
	"CF": true, "CG": true, "CI": true, "CM": true, "CV": true, "DJ": true,
	"DZ": true, "EG": true, "ER": true, "ET": true, "GA": true, "GH": true,
	"GM": true, "GN": true, "GQ": true, "GW": true, "KE": true, "KM": true,
	"LR": true, "LS": true, "LY": true, "MA": true, "MG": true, "ML": true,
	"MR": true, "MU": true, "MW": true, "MZ": true, "NA": true, "NE": true,
	"NG": true, "RW": true, "SC": true, "SD": true, "SL": true, "SN": true,
	"SO": true, "SS": true, "ST": true, "SZ": true, "TD": true, "TG": true,
	"TN": true, "TZ": true, "UG": true, "ZA": true, "ZM": true, "ZW": true,
}
