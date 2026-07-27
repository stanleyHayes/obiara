// Package stagingqualification composes repository-controlled synthetic
// staging evidence. It cannot produce a production approval.
package stagingqualification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/internal/platform/drrehearsal"
	"github.com/stanleyHayes/obiara/internal/quality/releasebundle"
)

const (
	SchemaVersion = "obiara.synthetic-staging-qualification.v1"
	MaxAge        = 24 * time.Hour
)

var shaPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

var requiredChecks = []string{
	"Lint, test, and build",
	"CodeQL (go)",
	"CodeQL (javascript-typescript)",
	"Dependency vulnerabilities",
}

type EvidenceRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Package struct {
	SchemaVersion string    `json:"schemaVersion"`
	Environment   string    `json:"environment"`
	SyntheticOnly bool      `json:"syntheticOnly"`
	CandidateSHA  string    `json:"candidateSha"`
	GeneratedAt   time.Time `json:"generatedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
	Inputs        struct {
		ReleaseEvidence EvidenceRef `json:"releaseEvidence"`
		ReleaseBundle   EvidenceRef `json:"releaseBundle"`
		DRRehearsal     EvidenceRef `json:"drRehearsal"`
	} `json:"inputs"`
	Limitations struct {
		ProductionEvidence bool `json:"productionEvidence"`
		LegalEvidence      bool `json:"legalEvidence"`
		ProviderEvidence   bool `json:"providerEvidence"`
		CohortEvidence     bool `json:"cohortEvidence"`
	} `json:"limitations"`
	Disposition string `json:"disposition"`
}

type ReleaseEvidence struct {
	SchemaVersion  string    `json:"schemaVersion"`
	Target         string    `json:"target"`
	SyntheticOnly  bool      `json:"syntheticOnly"`
	CommitSHA      string    `json:"commitSha"`
	GeneratedAt    time.Time `json:"generatedAt"`
	RequiredChecks []string  `json:"requiredChecks"`
	Disposition    string    `json:"disposition"`
}

type Sources struct {
	ReleaseEvidencePath string
	ReleaseBundlePath   string
	DRRehearsalPath     string
}

func Generate(candidateSHA string, generatedAt time.Time, sources Sources) (Package, error) {
	if !shaPattern.MatchString(candidateSHA) || generatedAt.IsZero() {
		return Package{}, errors.New("exact candidate SHA and generation time are required")
	}
	var result Package
	result.SchemaVersion = SchemaVersion
	result.Environment = "staging"
	result.SyntheticOnly = true
	result.CandidateSHA = candidateSHA
	result.GeneratedAt = generatedAt.UTC()
	result.ExpiresAt = generatedAt.UTC().Add(MaxAge)
	result.Disposition = "synthetic-staging-qualified"

	refs := []struct {
		path string
		dst  *EvidenceRef
	}{
		{sources.ReleaseEvidencePath, &result.Inputs.ReleaseEvidence},
		{sources.ReleaseBundlePath, &result.Inputs.ReleaseBundle},
		{sources.DRRehearsalPath, &result.Inputs.DRRehearsal},
	}
	for _, item := range refs {
		raw, err := os.ReadFile(item.path)
		if err != nil {
			return Package{}, err
		}
		sum := sha256.Sum256(raw)
		*item.dst = EvidenceRef{Path: logicalPath(item.path), SHA256: hex.EncodeToString(sum[:])}
	}
	if err := Validate(result, candidateSHA, generatedAt, sources); err != nil {
		return Package{}, err
	}
	return result, nil
}

func Load(path, expectedSHA string, now time.Time, sources Sources) (Package, error) {
	var qualification Package
	if err := decodeStrictFile(path, &qualification); err != nil {
		return Package{}, err
	}
	return qualification, Validate(qualification, expectedSHA, now, sources)
}

func Validate(q Package, expectedSHA string, now time.Time, sources Sources) error {
	if q.SchemaVersion != SchemaVersion || q.Environment != "staging" || !q.SyntheticOnly ||
		q.Disposition != "synthetic-staging-qualified" || !shaPattern.MatchString(expectedSHA) ||
		q.CandidateSHA != expectedSHA {
		return errors.New("qualification is not exact-SHA synthetic staging evidence")
	}
	if q.Limitations.ProductionEvidence || q.Limitations.LegalEvidence ||
		q.Limitations.ProviderEvidence || q.Limitations.CohortEvidence {
		return errors.New("synthetic qualification cannot satisfy external evidence")
	}
	if q.GeneratedAt.After(now) || now.Sub(q.GeneratedAt) > MaxAge ||
		!q.ExpiresAt.Equal(q.GeneratedAt.Add(MaxAge)) || !now.Before(q.ExpiresAt) {
		return errors.New("qualification evidence is stale")
	}
	if err := validateRef(q.Inputs.ReleaseEvidence, sources.ReleaseEvidencePath); err != nil {
		return fmt.Errorf("release evidence ref: %w", err)
	}
	if err := validateRef(q.Inputs.ReleaseBundle, sources.ReleaseBundlePath); err != nil {
		return fmt.Errorf("release bundle ref: %w", err)
	}
	if err := validateRef(q.Inputs.DRRehearsal, sources.DRRehearsalPath); err != nil {
		return fmt.Errorf("DR rehearsal ref: %w", err)
	}

	var release ReleaseEvidence
	if err := decodeStrictFile(sources.ReleaseEvidencePath, &release); err != nil {
		return err
	}
	if release.SchemaVersion != "obiara.release-evidence.v1" || release.Target != "staging" ||
		!release.SyntheticOnly || release.CommitSHA != expectedSHA ||
		release.Disposition != "qualified-non-production" || release.GeneratedAt.After(q.GeneratedAt) ||
		q.GeneratedAt.Sub(release.GeneratedAt) > MaxAge || !sameStrings(release.RequiredChecks, requiredChecks) {
		return errors.New("release evidence is incomplete, stale, or for another candidate")
	}

	bundle, err := releasebundle.Load(sources.ReleaseBundlePath, q.GeneratedAt)
	if err != nil {
		return err
	}
	if bundle.Environment != "staging" || bundle.Candidate.CommitSHA != expectedSHA ||
		bundle.Approvals.ProductionApproved || bundle.Disposition != "blocked" {
		return errors.New("release bundle is not a blocked synthetic staging bundle")
	}

	var dr drrehearsal.Evidence
	if err := decodeStrictFile(sources.DRRehearsalPath, &dr); err != nil {
		return err
	}
	if dr.SchemaVersion != 1 || dr.Environment != "staging" || !dr.SyntheticOnly ||
		!dr.PointInTime || dr.Result != "pass" || dr.RPOMinutesObserved > 5 ||
		dr.RTOMinutesObserved > 60 || dr.RestoreCompletedAt.After(q.GeneratedAt) ||
		q.GeneratedAt.Sub(dr.RestoreCompletedAt) > MaxAge ||
		!dr.SourceSnapshotAt.Before(dr.RestoreCompletedAt) ||
		dr.RestoreStartedAt.Before(dr.SourceSnapshotAt) ||
		!dr.RestoreCompletedAt.After(dr.RestoreStartedAt) ||
		!strings.HasPrefix(dr.Target, "isolated-") ||
		!digestPattern.MatchString(dr.ValidationDigest) ||
		!digestPattern.MatchString(dr.DataOwnerApprovalRef) ||
		!digestPattern.MatchString(dr.SecurityApprovalRef) ||
		dr.DataOwnerApprovalRef == dr.SecurityApprovalRef {
		return errors.New("DR evidence is incomplete, stale, or not synthetic staging")
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func validateRef(ref EvidenceRef, expectedPath string) error {
	if ref.Path != logicalPath(expectedPath) || len(ref.SHA256) != 64 {
		return errors.New("path or digest mismatch")
	}
	raw, err := os.ReadFile(expectedPath)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	if ref.SHA256 != hex.EncodeToString(sum[:]) {
		return errors.New("evidence digest mismatch")
	}
	return nil
}

func logicalPath(path string) string {
	clean := filepath.ToSlash(filepath.Clean(path))
	if index := strings.LastIndex(clean, "/deploy/"); index >= 0 {
		return clean[index+1:]
	}
	return clean
}

func decodeStrictFile(path string, dst any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("decode %s: trailing JSON", path)
	}
	return nil
}
