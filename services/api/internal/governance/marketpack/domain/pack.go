package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrInvalid = errors.New("invalid market-pack governance state")
var opaque = regexp.MustCompile(`^[a-f0-9]{64}$`)
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)
var localePattern = regexp.MustCompile(`^[a-z]{2,3}-[A-Z]{2}$`)
var placeholderPattern = regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]*)\}`)

const (
	MaxEntries   = 2000
	MaxTextRunes = 512
)

type MasterEntry struct{ Key, Text string }
type Term struct {
	Value          string
	DoNotTranslate bool
}
type MasterSpec struct {
	ID      string
	Version uint64
	Entries []MasterEntry
	Terms   []Term
}
type Master struct{ spec MasterSpec }

func NewMaster(spec MasterSpec) (Master, error) {
	spec.Entries = append([]MasterEntry(nil), spec.Entries...)
	spec.Terms = append([]Term(nil), spec.Terms...)
	if !keyPattern.MatchString(spec.ID) || spec.Version == 0 || len(spec.Entries) == 0 || len(spec.Entries) > MaxEntries {
		return Master{}, ErrInvalid
	}
	seen := map[string]bool{}
	for _, entry := range spec.Entries {
		if !keyPattern.MatchString(entry.Key) || !validText(entry.Text) || seen[entry.Key] {
			return Master{}, ErrInvalid
		}
		seen[entry.Key] = true
	}
	termSeen := map[string]bool{}
	for _, term := range spec.Terms {
		if strings.TrimSpace(term.Value) == "" || utf8.RuneCountInString(term.Value) > 64 || termSeen[term.Value] {
			return Master{}, ErrInvalid
		}
		termSeen[term.Value] = true
	}
	slices.SortFunc(spec.Entries, func(a, b MasterEntry) int { return strings.Compare(a.Key, b.Key) })
	slices.SortFunc(spec.Terms, func(a, b Term) int { return strings.Compare(a.Value, b.Value) })
	return Master{spec}, nil
}
func (m Master) Spec() MasterSpec {
	s := m.spec
	s.Entries = append([]MasterEntry(nil), m.spec.Entries...)
	s.Terms = append([]Term(nil), m.spec.Terms...)
	return s
}

type Translation struct{ Key, Text string }
type ReviewStage string

const (
	StageProfessional ReviewStage = "professional_translation"
	StageCommunity    ReviewStage = "community_language"
	StageInContext    ReviewStage = "in_context"
)

type Check string

const (
	CheckMeaning     Check = "meaning_checked"
	CheckVoice       Check = "voice_checked"
	CheckCulturalFit Check = "cultural_fit_checked"
	CheckDignity     Check = "dignity_checked"
	CheckScreenshot  Check = "screenshot_checked"
	CheckTruncation  Check = "truncation_checked"
)

type Review struct {
	Stage       ReviewStage
	ReviewerKey string
	Checks      []Check
	EvidenceRef string
	ReviewedAt  time.Time
}
type Approval struct {
	ApproverKey, ReasoningRef string
	ApprovedAt                time.Time
}
type Audit struct {
	Sequence                             uint64
	Kind, CommandID, ActorKey, Reference string
	At                                   time.Time
}
type State struct {
	ID, Market, Locale, MasterID, AuthorKey string
	ContentDigest                           string
	Version, MasterVersion                  uint64
	Translations                            []Translation
	Reviews                                 []Review
	Approval                                *Approval
	Revision                                uint64
	Audit                                   []Audit
	AppliedIDs                              []string
}
type Pack struct{ state State }

func Propose(id, market, locale, author string, version uint64, master Master, translations []Translation, command string, at time.Time) (Pack, error) {
	translations = append([]Translation(nil), translations...)
	if !opaque.MatchString(id) || !opaque.MatchString(author) || !localePattern.MatchString(locale) || len(market) != 2 || market != strings.ToUpper(market) || version == 0 || !keyPattern.MatchString(command) || at.IsZero() || !parity(master, translations) {
		return Pack{}, ErrInvalid
	}
	slices.SortFunc(translations, func(a, b Translation) int { return strings.Compare(a.Key, b.Key) })
	spec := master.Spec()
	s := State{ID: id, Market: market, Locale: locale, MasterID: spec.ID, MasterVersion: spec.Version, AuthorKey: author, Version: version, Translations: translations, ContentDigest: translationDigest(translations), Revision: 1, AppliedIDs: []string{command}}
	s.Audit = []Audit{{1, "proposed", command, author, spec.ID, at.UTC()}}
	return Pack{s}, nil
}
func Rehydrate(s State) (Pack, error) {
	s = clone(s)
	if !validState(s) {
		return Pack{}, ErrInvalid
	}
	return Pack{s}, nil
}
func (p Pack) AddReview(review Review, command string) (Pack, error) {
	review.Checks = append([]Check(nil), review.Checks...)
	if p.state.Approval != nil || !validReview(review) || review.ReviewerKey == p.state.AuthorKey || !keyPattern.MatchString(command) || slices.Contains(p.state.AppliedIDs, command) {
		return Pack{}, ErrInvalid
	}
	for _, existing := range p.state.Reviews {
		if existing.Stage == review.Stage || existing.ReviewerKey == review.ReviewerKey {
			return Pack{}, ErrInvalid
		}
	}
	n := p.State()
	n.Revision++
	n.Reviews = append(n.Reviews, review)
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Audit = append(n.Audit, Audit{n.Revision, "reviewed", command, review.ReviewerKey, review.EvidenceRef, review.ReviewedAt.UTC()})
	return Pack{n}, nil
}
func (p Pack) Approve(approver, reasoningRef, command string, at time.Time) (Pack, error) {
	if p.state.Approval != nil || !opaque.MatchString(approver) || !opaque.MatchString(reasoningRef) || approver == p.state.AuthorKey || !allReviews(p.state.Reviews) || !keyPattern.MatchString(command) || at.IsZero() || slices.Contains(p.state.AppliedIDs, command) {
		return Pack{}, ErrInvalid
	}
	for _, review := range p.state.Reviews {
		if review.ReviewerKey == approver {
			return Pack{}, ErrInvalid
		}
	}
	n := p.State()
	n.Revision++
	n.Approval = &Approval{approver, reasoningRef, at.UTC()}
	n.AppliedIDs = append(n.AppliedIDs, command)
	n.Audit = append(n.Audit, Audit{n.Revision, "publish_ready", command, approver, reasoningRef, at.UTC()})
	return Pack{n}, nil
}
func (p Pack) PublishReady() bool { return p.state.Approval != nil }
func (p Pack) State() State       { return clone(p.state) }
func (p Pack) ID() string         { return p.state.ID }
func (p Pack) Revision() uint64   { return p.state.Revision }
func parity(master Master, translations []Translation) bool {
	spec := master.Spec()
	if len(translations) != len(spec.Entries) {
		return false
	}
	byKey := map[string]string{}
	for _, translation := range translations {
		if !keyPattern.MatchString(translation.Key) || !validText(translation.Text) || byKey[translation.Key] != "" {
			return false
		}
		byKey[translation.Key] = translation.Text
	}
	for _, entry := range spec.Entries {
		translated, ok := byKey[entry.Key]
		if !ok || !slices.Equal(placeholders(entry.Text), placeholders(translated)) {
			return false
		}
		for _, term := range spec.Terms {
			if term.DoNotTranslate && strings.Contains(entry.Text, term.Value) && !strings.Contains(translated, term.Value) {
				return false
			}
		}
	}
	return true
}
func placeholders(value string) []string {
	matches := placeholderPattern.FindAllStringSubmatch(value, -1)
	result := make([]string, 0, len(matches))
	for _, match := range matches {
		result = append(result, match[1])
	}
	slices.Sort(result)
	return result
}
func validText(value string) bool {
	return strings.TrimSpace(value) != "" && utf8.ValidString(value) && utf8.RuneCountInString(value) <= MaxTextRunes && !strings.ContainsRune(value, '\x00')
}
func validReview(r Review) bool {
	if !opaque.MatchString(r.ReviewerKey) || !opaque.MatchString(r.EvidenceRef) || r.ReviewedAt.IsZero() {
		return false
	}
	required := map[ReviewStage][]Check{StageProfessional: {CheckMeaning, CheckVoice}, StageCommunity: {CheckCulturalFit, CheckDignity}, StageInContext: {CheckScreenshot, CheckTruncation}}[r.Stage]
	copyChecks := append([]Check(nil), r.Checks...)
	slices.Sort(copyChecks)
	slices.Sort(required)
	return slices.Equal(copyChecks, required)
}
func allReviews(reviews []Review) bool {
	if len(reviews) != 3 {
		return false
	}
	seen := map[ReviewStage]bool{}
	for _, review := range reviews {
		if !validReview(review) || seen[review.Stage] {
			return false
		}
		seen[review.Stage] = true
	}
	return true
}
func validState(s State) bool {
	if !opaque.MatchString(s.ID) || !opaque.MatchString(s.AuthorKey) ||
		!localePattern.MatchString(s.Locale) || len(s.Market) != 2 ||
		s.Market != strings.ToUpper(s.Market) || s.Version == 0 ||
		!keyPattern.MatchString(s.MasterID) || s.MasterVersion == 0 ||
		s.ContentDigest != translationDigest(s.Translations) ||
		s.Revision == 0 || len(s.Audit) != int(s.Revision) ||
		len(s.AppliedIDs) != int(s.Revision) {
		return false
	}
	seen := map[string]bool{}
	for _, translation := range s.Translations {
		if !keyPattern.MatchString(translation.Key) || !validText(translation.Text) || seen[translation.Key] {
			return false
		}
		seen[translation.Key] = true
	}
	for index, audit := range s.Audit {
		if audit.Sequence != uint64(index+1) || audit.CommandID != s.AppliedIDs[index] ||
			!keyPattern.MatchString(audit.CommandID) || audit.At.IsZero() {
			return false
		}
	}
	if s.Approval != nil && !allReviews(s.Reviews) {
		return false
	}
	return true
}

func translationDigest(translations []Translation) string {
	canonical := append([]Translation(nil), translations...)
	slices.SortFunc(canonical, func(a, b Translation) int { return strings.Compare(a.Key, b.Key) })
	hash := sha256.New()
	for _, translation := range canonical {
		_, _ = fmt.Fprintf(hash, "%d:%s:%d:%s;", len(translation.Key), translation.Key, len(translation.Text), translation.Text)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
func clone(s State) State {
	s.Translations = append([]Translation(nil), s.Translations...)
	s.Reviews = append([]Review(nil), s.Reviews...)
	for i := range s.Reviews {
		s.Reviews[i].Checks = append([]Check(nil), s.Reviews[i].Checks...)
	}
	s.Audit = append([]Audit(nil), s.Audit...)
	s.AppliedIDs = append([]string(nil), s.AppliedIDs...)
	if s.Approval != nil {
		x := *s.Approval
		s.Approval = &x
	}
	return s
}
