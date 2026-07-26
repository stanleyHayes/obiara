// Package domain implements deterministic cold-start candidate projection
// from privacy-minimized inputs. It does not model or accept traits, model
// scores, popularity, game skill, raw graph paths, or vendor output.
package domain

import (
	"errors"
	"regexp"
	"slices"
)

const (
	MaxInputCandidates = 200
	MaxCandidates      = 20
	MaxReasons         = 4
	MaxTrustHops       = 4
)

type TrustReason string

const (
	TrustSharedCircle TrustReason = "shared_circle"
	TrustVouched      TrustReason = "vouched_connection"
	TrustKnown        TrustReason = "known_connection"
	TrustHost         TrustReason = "host_connection"
)

type ReasonCode string

const (
	ReasonReciprocalPreference ReasonCode = "reciprocal_preference"
	ReasonSharedCircle         ReasonCode = "shared_circle"
	ReasonVouchedConnection    ReasonCode = "vouched_connection"
	ReasonKnownConnection      ReasonCode = "known_connection"
	ReasonHostConnection       ReasonCode = "host_connection"
)

var (
	ErrInvalidInput = errors.New("invalid cold-start matching input")
	ErrInputBound   = errors.New("cold-start matching input exceeds bound")
)

var opaqueKeyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// ReciprocalPreference is a privacy-minimized summary from the explicit
// preference system. It intentionally contains no preference values or
// inferred traits.
type ReciprocalPreference struct {
	CandidateKey               string
	RequesterExplicit          bool
	CandidateExplicit          bool
	RequesterPreferenceVersion uint64
	CandidatePreferenceVersion uint64
}

// TrustSummary is already privacy-scoped by the trust visibility boundary.
// It contains no edge IDs, endpoints, intermediary identities, or raw path.
type TrustSummary struct {
	CandidateKey string
	Reason       TrustReason
	Hops         int
}

type Candidate struct {
	CandidateKey string
	Reasons      []ReasonCode
}

// Project intersects explicit reciprocal preferences with privacy-scoped
// trust summaries. Results are lexicographically ordered by opaque key; this
// is stable enumeration, not a rank or compatibility score.
func Project(requesterKey string, preferences []ReciprocalPreference, trust []TrustSummary, limit int) ([]Candidate, error) {
	if !opaqueKeyPattern.MatchString(requesterKey) || limit < 1 || limit > MaxCandidates {
		return nil, ErrInvalidInput
	}
	if len(preferences) > MaxInputCandidates || len(trust) > MaxInputCandidates {
		return nil, ErrInputBound
	}

	reciprocal := make(map[string]struct{}, len(preferences))
	seenPreferences := make(map[string]struct{}, len(preferences))
	for _, preference := range preferences {
		if !opaqueKeyPattern.MatchString(preference.CandidateKey) ||
			preference.CandidateKey == requesterKey ||
			preference.RequesterPreferenceVersion == 0 ||
			preference.CandidatePreferenceVersion == 0 {
			return nil, ErrInvalidInput
		}
		if _, duplicate := seenPreferences[preference.CandidateKey]; duplicate {
			return nil, ErrInvalidInput
		}
		seenPreferences[preference.CandidateKey] = struct{}{}
		if preference.RequesterExplicit && preference.CandidateExplicit {
			reciprocal[preference.CandidateKey] = struct{}{}
		}
	}

	reasons := make(map[string][]ReasonCode)
	for _, summary := range trust {
		if !opaqueKeyPattern.MatchString(summary.CandidateKey) ||
			summary.Hops < 1 || summary.Hops > MaxTrustHops {
			return nil, ErrInvalidInput
		}
		if _, allowed := reciprocal[summary.CandidateKey]; !allowed {
			continue
		}
		reason := reasonFor(summary.Reason)
		if reason == "" {
			return nil, ErrInvalidInput
		}
		if !slices.Contains(reasons[summary.CandidateKey], reason) {
			reasons[summary.CandidateKey] = append(reasons[summary.CandidateKey], reason)
		}
	}

	keys := make([]string, 0, len(reasons))
	for key := range reasons {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	if len(keys) > limit {
		keys = keys[:limit]
	}

	candidates := make([]Candidate, 0, len(keys))
	for _, key := range keys {
		candidateReasons := reasons[key]
		slices.SortFunc(candidateReasons, func(a, b ReasonCode) int {
			return reasonOrder(a) - reasonOrder(b)
		})
		candidateReasons = append([]ReasonCode{ReasonReciprocalPreference}, candidateReasons...)
		if len(candidateReasons) > MaxReasons {
			candidateReasons = candidateReasons[:MaxReasons]
		}
		candidates = append(candidates, Candidate{
			CandidateKey: key,
			Reasons:      append([]ReasonCode(nil), candidateReasons...),
		})
	}
	return candidates, nil
}

func reasonFor(reason TrustReason) ReasonCode {
	switch reason {
	case TrustSharedCircle:
		return ReasonSharedCircle
	case TrustVouched:
		return ReasonVouchedConnection
	case TrustKnown:
		return ReasonKnownConnection
	case TrustHost:
		return ReasonHostConnection
	default:
		return ""
	}
}

func reasonOrder(reason ReasonCode) int {
	switch reason {
	case ReasonSharedCircle:
		return 1
	case ReasonVouchedConnection:
		return 2
	case ReasonKnownConnection:
		return 3
	case ReasonHostConnection:
		return 4
	default:
		return 100
	}
}
