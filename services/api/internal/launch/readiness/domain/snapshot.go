package domain

import (
	"errors"
	"regexp"
	"slices"
	"time"
)

var ErrInvalid = errors.New("invalid launch readiness snapshot")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var token = regexp.MustCompile(`^[a-z][a-z0-9_.-]{1,63}$`)

const (
	MaxFamilies    = 1_000_000
	MaxHosts       = 100_000
	MaxMatchmakers = 100_000
)

type FamilyDensity struct {
	Market                                                                string
	ConsentedFamilies, TargetFamilies, DenseCircles, RequiredDenseCircles int
	EvidenceVersion                                                       uint64
	Complete                                                              bool
	ObservedAt                                                            time.Time
}
type HostCoverage struct {
	Market                                string
	Trained, Certified, Required          int
	TrainingVersion, CertificationVersion uint64
	CertifiedUntil                        time.Time
	Complete                              bool
	ObservedAt                            time.Time
}
type LicenseCoverage struct {
	Market, Jurisdiction string
	Licensed, Required   int
	LicenseVersion       uint64
	LicensedUntil        time.Time
	Complete             bool
	ObservedAt           time.Time
}
type Blocker string

const (
	FamilyEvidenceIncomplete  Blocker = "family_evidence_incomplete"
	FamilyTargetUnmet         Blocker = "family_target_unmet"
	DensityUnmet              Blocker = "density_unmet"
	HostEvidenceIncomplete    Blocker = "host_evidence_incomplete"
	HostCoverageUnmet         Blocker = "host_coverage_unmet"
	HostCertificationExpired  Blocker = "host_certification_expired"
	LicenseEvidenceIncomplete Blocker = "license_evidence_incomplete"
	LicenseCoverageUnmet      Blocker = "license_coverage_unmet"
	LicenseExpired            Blocker = "license_expired"
	JurisdictionMismatch      Blocker = "jurisdiction_mismatch"
)

type Audit struct {
	Sequence                  uint64
	Kind, CommandID, ActorKey string
	At                        time.Time
}
type State struct {
	ID, ActorKey, Market, Jurisdiction string
	Families                           FamilyDensity
	Hosts                              HostCoverage
	Licenses                           LicenseCoverage
	Ready                              bool
	Blockers                           []Blocker
	Revision                           uint64
	Audit                              []Audit
	AppliedIDs                         []string
}
type Snapshot struct{ state State }

func Project(id, actor, market, jurisdiction, command string, f FamilyDensity, h HostCoverage, l LicenseCoverage, now time.Time) (Snapshot, error) {
	if !opaque.MatchString(id) || !opaque.MatchString(actor) || !token.MatchString(command) || !validMarket(market) || !token.MatchString(jurisdiction) || now.IsZero() ||
		!validFamilies(f, market) || !validHosts(h, market) || !validLicenses(l, market) {
		return Snapshot{}, ErrInvalid
	}
	blockers := evaluate(jurisdiction, f, h, l, now)
	s := State{ID: id, ActorKey: actor, Market: market, Jurisdiction: jurisdiction, Families: f, Hosts: h, Licenses: l, Ready: len(blockers) == 0, Blockers: blockers, Revision: 1, AppliedIDs: []string{command}}
	s.Audit = []Audit{{1, "snapshot_projected", command, actor, now.UTC()}}
	return Snapshot{s}, nil
}
func Rehydrate(s State) (Snapshot, error) {
	s = clone(s)
	if !validState(s) {
		return Snapshot{}, ErrInvalid
	}
	return Snapshot{s}, nil
}
func (s Snapshot) State() State { return clone(s.state) }
func (s Snapshot) ID() string   { return s.state.ID }
func (s Snapshot) Ready() bool  { return s.state.Ready }

func evaluate(jurisdiction string, f FamilyDensity, h HostCoverage, l LicenseCoverage, now time.Time) []Blocker {
	var b []Blocker
	if !f.Complete {
		b = append(b, FamilyEvidenceIncomplete)
	}
	if f.ConsentedFamilies < f.TargetFamilies {
		b = append(b, FamilyTargetUnmet)
	}
	if f.DenseCircles < f.RequiredDenseCircles {
		b = append(b, DensityUnmet)
	}
	if !h.Complete {
		b = append(b, HostEvidenceIncomplete)
	}
	if h.Trained < h.Required || h.Certified < h.Required {
		b = append(b, HostCoverageUnmet)
	}
	if !h.CertifiedUntil.After(now) {
		b = append(b, HostCertificationExpired)
	}
	if !l.Complete {
		b = append(b, LicenseEvidenceIncomplete)
	}
	if l.Licensed < l.Required {
		b = append(b, LicenseCoverageUnmet)
	}
	if !l.LicensedUntil.After(now) {
		b = append(b, LicenseExpired)
	}
	if l.Jurisdiction != jurisdiction {
		b = append(b, JurisdictionMismatch)
	}
	slices.Sort(b)
	return b
}
func validMarket(v string) bool {
	return len(v) == 2 && v[0] >= 'A' && v[0] <= 'Z' && v[1] >= 'A' && v[1] <= 'Z'
}
func validFamilies(f FamilyDensity, m string) bool {
	return f.Market == m && f.ConsentedFamilies >= 0 && f.ConsentedFamilies <= MaxFamilies && f.TargetFamilies > 0 && f.TargetFamilies <= MaxFamilies && f.DenseCircles >= 0 && f.DenseCircles <= MaxFamilies && f.RequiredDenseCircles > 0 && f.RequiredDenseCircles <= MaxFamilies && f.EvidenceVersion > 0 && !f.ObservedAt.IsZero()
}
func validHosts(h HostCoverage, m string) bool {
	return h.Market == m && h.Trained >= 0 && h.Trained <= MaxHosts && h.Certified >= 0 && h.Certified <= MaxHosts && h.Required > 0 && h.Required <= MaxHosts && h.TrainingVersion > 0 && h.CertificationVersion > 0 && !h.CertifiedUntil.IsZero() && !h.ObservedAt.IsZero()
}
func validLicenses(l LicenseCoverage, m string) bool {
	return l.Market == m && token.MatchString(l.Jurisdiction) && l.Licensed >= 0 && l.Licensed <= MaxMatchmakers && l.Required > 0 && l.Required <= MaxMatchmakers && l.LicenseVersion > 0 && !l.LicensedUntil.IsZero() && !l.ObservedAt.IsZero()
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.ActorKey) || !validMarket(s.Market) || !token.MatchString(s.Jurisdiction) || !validFamilies(s.Families, s.Market) || !validHosts(s.Hosts, s.Market) || !validLicenses(s.Licenses, s.Market) || s.Revision != 1 || len(s.Audit) != 1 || len(s.AppliedIDs) != 1 {
		return false
	}
	a := s.Audit[0]
	if a.Sequence != 1 || a.Kind != "snapshot_projected" || a.CommandID != s.AppliedIDs[0] || !token.MatchString(a.CommandID) || a.ActorKey != s.ActorKey || a.At.IsZero() {
		return false
	}
	want := evaluate(s.Jurisdiction, s.Families, s.Hosts, s.Licenses, a.At)
	return s.Ready == (len(want) == 0) && slices.Equal(s.Blockers, want)
}
func clone(s State) State {
	s.Blockers = append([]Blocker(nil), s.Blockers...)
	s.Audit = append([]Audit(nil), s.Audit...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	return s
}
