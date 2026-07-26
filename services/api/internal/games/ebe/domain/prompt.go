// Package domain implements the pure reviewed-content and private-duel
// rules for E10-S05 Ɛbɛ. It has no publishing, matching, rating, popularity,
// persistence, or cultural-authority surface.
package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	MaxPromptRunes   = 500
	MaxAnswerRunes   = 280
	MaxAcceptedForms = 20
)

type SourceKind string

const (
	SourceBook                 SourceKind = "book"
	SourceOralArchive          SourceKind = "oral_archive"
	SourceInstitutionalArchive SourceKind = "institutional_archive"
)

type ReviewDecision string

const DecisionApproved ReviewDecision = "approved"

var (
	ErrPromptInvalid    = errors.New("invalid reviewed proverb prompt")
	ErrPromptUnapproved = errors.New("proverb prompt is not reviewer-approved")
	ErrAnswerInvalid    = errors.New("invalid proverb answer")
)

var (
	opaqueKeyPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
	idPattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
	languagePattern  = regexp.MustCompile(`^[a-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$`)
)

// Source is a citation, not a claim that Obiara or a reviewer is a cultural
// authority. Citation is human-readable and Locator may be an HTTPS archive
// or catalogue URL.
type Source struct {
	Kind     SourceKind
	Citation string
	Locator  string
}

// Review records the decision provenance for exactly one content version.
// ReviewerKey is an opaque internal reference.
type Review struct {
	ID          string
	ReviewerKey string
	Decision    ReviewDecision
	ReviewedAt  time.Time
}

type PromptSpec struct {
	ID              string
	Version         uint64
	Language        string
	Cue             string
	AcceptedAnswers []string
	Source          Source
	Review          Review
}

// Prompt is an immutable, versioned, reviewer-approved prompt snapshot.
type Prompt struct {
	spec            PromptSpec
	normalizedForms []string
	digest          string
}

func NewApprovedPrompt(spec PromptSpec) (Prompt, error) {
	spec = clonePromptSpec(spec)
	if !idPattern.MatchString(spec.ID) || spec.Version == 0 ||
		!languagePattern.MatchString(spec.Language) ||
		!boundedText(spec.Cue, MaxPromptRunes) ||
		len(spec.AcceptedAnswers) == 0 || len(spec.AcceptedAnswers) > MaxAcceptedForms ||
		!validSource(spec.Source) || !validReview(spec.Review) {
		return Prompt{}, ErrPromptInvalid
	}
	if spec.Review.Decision != DecisionApproved {
		return Prompt{}, ErrPromptUnapproved
	}
	forms := make([]string, 0, len(spec.AcceptedAnswers))
	for index, answer := range spec.AcceptedAnswers {
		if !boundedText(answer, MaxAnswerRunes) {
			return Prompt{}, ErrAnswerInvalid
		}
		spec.AcceptedAnswers[index] = strings.TrimSpace(answer)
		normalized := normalizeAnswer(answer)
		if normalized == "" || slices.Contains(forms, normalized) {
			return Prompt{}, ErrAnswerInvalid
		}
		forms = append(forms, normalized)
	}
	slices.Sort(forms)
	return Prompt{
		spec:            spec,
		normalizedForms: forms,
		digest:          promptDigest(spec, forms),
	}, nil
}

func (prompt Prompt) Spec() PromptSpec {
	return clonePromptSpec(prompt.spec)
}

func (prompt Prompt) Digest() string { return prompt.digest }

func (prompt Prompt) Accepts(answer string) (bool, error) {
	if !boundedText(answer, MaxAnswerRunes) {
		return false, ErrAnswerInvalid
	}
	return slices.Contains(prompt.normalizedForms, normalizeAnswer(answer)), nil
}

func validSource(source Source) bool {
	if source.Kind != SourceBook && source.Kind != SourceOralArchive && source.Kind != SourceInstitutionalArchive {
		return false
	}
	if !boundedText(source.Citation, 500) || len(source.Locator) > 1000 {
		return false
	}
	if source.Locator == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(source.Locator)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func validReview(review Review) bool {
	return idPattern.MatchString(review.ID) &&
		opaqueKeyPattern.MatchString(review.ReviewerKey) &&
		!review.ReviewedAt.IsZero()
}

func boundedText(value string, limit int) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed != "" && utf8.ValidString(trimmed) && utf8.RuneCountInString(trimmed) <= limit
}

func normalizeAnswer(value string) string {
	return strings.Join(strings.Fields(norm.NFKC.String(strings.ToLower(strings.TrimSpace(value)))), " ")
}

func promptDigest(spec PromptSpec, forms []string) string {
	payload := strings.Join([]string{
		spec.ID,
		strconv.FormatUint(spec.Version, 10),
		spec.Language,
		spec.Cue,
		string(spec.Source.Kind),
		spec.Source.Citation,
		spec.Source.Locator,
		spec.Review.ID,
		spec.Review.ReviewerKey,
		spec.Review.ReviewedAt.UTC().Format(time.RFC3339Nano),
		strings.Join(forms, "\x1e"),
	}, "\x1f")
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func clonePromptSpec(spec PromptSpec) PromptSpec {
	spec.AcceptedAnswers = append([]string(nil), spec.AcceptedAnswers...)
	spec.Cue = strings.TrimSpace(spec.Cue)
	spec.Source.Citation = strings.TrimSpace(spec.Source.Citation)
	spec.Review.ReviewedAt = spec.Review.ReviewedAt.UTC()
	return spec
}
