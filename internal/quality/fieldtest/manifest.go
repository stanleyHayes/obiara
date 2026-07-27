// Package fieldtest validates evidence captured by the Ghana Android and
// constrained-network field-test runbook. It performs no device or network I/O.
package fieldtest

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

const (
	SchemaVersion = "obiara.ghana-field-test.v1"
	MaxAge        = 7 * 24 * time.Hour
	MinSamples    = 30
)

var (
	sha40  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	digest = regexp.MustCompile(`^[0-9a-f]{64}$`)
	opaque = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{7,127}$`)
)

type Device struct {
	DeviceRef    string `json:"deviceRef"`
	Android      string `json:"android"`
	APILevel     int    `json:"apiLevel"`
	RAMMiB       int    `json:"ramMiB"`
	Physical     bool   `json:"physical"`
	CleanInstall bool   `json:"cleanInstall"`
}

type Network struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	DownKbps        int    `json:"downKbps"`
	UpKbps          int    `json:"upKbps"`
	LatencyMS       int    `json:"latencyMs"`
	JitterMS        int    `json:"jitterMs"`
	PacketLossBPS   int    `json:"packetLossBps"`
	ExternalTraffic bool   `json:"externalTraffic"`
}

type Measurement struct {
	Path        string  `json:"path"`
	Category    string  `json:"category"`
	Metric      string  `json:"metric"`
	NetworkID   string  `json:"networkId"`
	Unit        string  `json:"unit"`
	Samples     []int64 `json:"samples"`
	P50         int64   `json:"p50"`
	P90         int64   `json:"p90"`
	P95         int64   `json:"p95"`
	P90Budget   int64   `json:"p90Budget"`
	EvidenceRef string  `json:"evidenceRef"`
}

type Manifest struct {
	SchemaVersion string        `json:"schemaVersion"`
	Environment   string        `json:"environment"`
	SyntheticOnly bool          `json:"syntheticOnly"`
	CandidateSHA  string        `json:"candidateSha"`
	Device        Device        `json:"device"`
	Networks      []Network     `json:"networks"`
	Measurements  []Measurement `json:"measurements"`
	Safety        struct {
		ProductionData bool `json:"productionData"`
		MemberData     bool `json:"memberData"`
		ExternalSpend  bool `json:"externalSpend"`
	} `json:"safety"`
	OperatorRef string    `json:"operatorRef"`
	ReviewerRef string    `json:"reviewerRef"`
	GeneratedAt time.Time `json:"generatedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Disposition string    `json:"disposition"`
	Blockers    []string  `json:"blockers"`
}

var requiredNetworks = map[string]Network{
	"ghana-3g": {
		ID: "ghana-3g", Kind: "representative-3g", DownKbps: 1024, UpKbps: 384,
		LatencyMS: 300, JitterMS: 50, PacketLossBPS: 200,
	},
	"fixed-local": {
		ID: "fixed-local", Kind: "fixed", DownKbps: 10000, UpKbps: 10000,
		LatencyMS: 30, JitterMS: 5, PacketLossBPS: 0,
	},
}

type pathContract struct {
	category  string
	metric    string
	networkID string
	unit      string
	budget    int64
}

var requiredPaths = map[string]pathContract{
	"cold-launch":          {"compute", "cold-start", "ghana-3g", "ms", 3000},
	"warm-launch":          {"compute", "warm-start", "fixed-local", "ms", 1500},
	"offline-reconcile":    {"database", "offline-reconcile", "ghana-3g", "ms", 3000},
	"voice-upload-60s":     {"media", "voice-upload-60s", "ghana-3g", "ms", 3000},
	"progressive-playback": {"media", "progressive-playback", "ghana-3g", "ms", 2000},
	"fire-listen-only":     {"media", "listen-only-transition", "ghana-3g", "ms", 5000},
}

func Load(path, expectedSHA string, now time.Time) (Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode field-test manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Manifest{}, errors.New("field-test manifest has trailing JSON")
	}
	return manifest, Validate(manifest, expectedSHA, now)
}

func Validate(m Manifest, expectedSHA string, now time.Time) error {
	if m.SchemaVersion != SchemaVersion || m.Environment != "staging" ||
		!sha40.MatchString(expectedSHA) || m.CandidateSHA != expectedSHA {
		return errors.New("manifest is not exact-SHA staging evidence")
	}
	if m.Safety.ProductionData || m.Safety.MemberData || m.Safety.ExternalSpend {
		return errors.New("field test must use local synthetic data without external spend")
	}
	if !opaque.MatchString(m.OperatorRef) || !opaque.MatchString(m.ReviewerRef) ||
		m.OperatorRef == m.ReviewerRef {
		return errors.New("distinct opaque operator and reviewer are required")
	}
	if m.GeneratedAt.After(now) || now.Sub(m.GeneratedAt) > MaxAge ||
		!m.ExpiresAt.Equal(m.GeneratedAt.Add(MaxAge)) || !now.Before(m.ExpiresAt) {
		return errors.New("field-test evidence is stale")
	}
	if err := validateDevice(m.Device, m.SyntheticOnly); err != nil {
		return err
	}
	if err := validateNetworks(m.Networks); err != nil {
		return err
	}
	budgetMisses, err := validateMeasurements(m.Measurements)
	if err != nil {
		return err
	}
	expectedBlockers := append([]string(nil), budgetMisses...)
	if m.SyntheticOnly {
		expectedBlockers = append(expectedBlockers, "synthetic-evidence")
	}
	if !m.Device.Physical {
		expectedBlockers = append(expectedBlockers, "physical-api26-device-required")
	}
	slices.Sort(expectedBlockers)
	actualBlockers := append([]string(nil), m.Blockers...)
	slices.Sort(actualBlockers)
	if !slices.Equal(actualBlockers, expectedBlockers) {
		return errors.New("blockers do not match measured facts")
	}
	if len(expectedBlockers) > 0 {
		if m.Disposition != "blocked" {
			return errors.New("incomplete or synthetic evidence must remain blocked")
		}
	} else if m.Disposition != "qualified-field-evidence" {
		return errors.New("complete field evidence has invalid disposition")
	}
	return nil
}

func validateDevice(d Device, synthetic bool) error {
	if !opaque.MatchString(d.DeviceRef) || d.Android != "8.0" || d.APILevel != 26 ||
		d.RAMMiB != 2048 || !d.CleanInstall {
		return errors.New("exact Android 8 API 26 2 GiB clean-install floor is required")
	}
	if !synthetic && !d.Physical {
		return errors.New("non-synthetic evidence requires physical hardware")
	}
	return nil
}

func validateNetworks(networks []Network) error {
	if len(networks) != len(requiredNetworks) {
		return errors.New("exact representative 3G and fixed network profiles are required")
	}
	seen := make(map[string]bool, len(networks))
	for _, network := range networks {
		expected, ok := requiredNetworks[network.ID]
		if !ok || seen[network.ID] || network.ExternalTraffic ||
			network.Kind != expected.Kind || network.DownKbps != expected.DownKbps ||
			network.UpKbps != expected.UpKbps || network.LatencyMS != expected.LatencyMS ||
			network.JitterMS != expected.JitterMS || network.PacketLossBPS != expected.PacketLossBPS {
			return errors.New("network profile drift or external traffic")
		}
		seen[network.ID] = true
	}
	return nil
}

func validateMeasurements(measurements []Measurement) ([]string, error) {
	if len(measurements) != len(requiredPaths) {
		return nil, errors.New("fixed field-test paths are incomplete")
	}
	seen := make(map[string]bool, len(measurements))
	var misses []string
	for _, measurement := range measurements {
		contract, ok := requiredPaths[measurement.Path]
		if !ok || seen[measurement.Path] || measurement.Category != contract.category ||
			measurement.Metric != contract.metric || measurement.NetworkID != contract.networkID ||
			measurement.Unit != contract.unit || measurement.P90Budget != contract.budget ||
			!digest.MatchString(measurement.EvidenceRef) || len(measurement.Samples) < MinSamples {
			return nil, errors.New("measurement contract is incomplete")
		}
		seen[measurement.Path] = true
		sorted := append([]int64(nil), measurement.Samples...)
		for _, sample := range sorted {
			if sample < 0 {
				return nil, errors.New("measurement sample cannot be negative")
			}
		}
		slices.Sort(sorted)
		if measurement.P50 != percentile(sorted, 50) ||
			measurement.P90 != percentile(sorted, 90) ||
			measurement.P95 != percentile(sorted, 95) {
			return nil, errors.New("reported percentile does not match samples")
		}
		if measurement.P90 > measurement.P90Budget {
			misses = append(misses, "budget:"+measurement.Path)
		}
	}
	return misses, nil
}

func percentile(sorted []int64, percent int) int64 {
	index := (percent*len(sorted)+99)/100 - 1
	return sorted[index]
}
