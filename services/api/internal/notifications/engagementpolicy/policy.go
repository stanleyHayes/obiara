// Package engagementpolicy is a pure, fail-closed boundary for notification
// copy. It decides only whether reviewed copy is allowed; it cannot generate,
// rank, persist, dispatch, or bypass delivery controls.
package engagementpolicy

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	MaxPatternsPerLocale = 128
	MaxTags              = 16
	MaxFieldRunes        = 4096
)

var (
	ErrInvalidCatalog      = errors.New("invalid reviewed engagement policy catalog")
	ErrInvalidNotification = errors.New("invalid notification copy")
)

var (
	tokenPattern  = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,63}$`)
	localePattern = regexp.MustCompile(`^[a-z]{2}(?:-[A-Z]{2})?$`)
	opaquePattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type Category string

const (
	CategoryTransactional Category = "transactional"
	CategorySafety        Category = "safety"
	CategoryRitual        Category = "ritual"
	CategoryCompanion     Category = "companion"
	CategoryCourtship     Category = "courtship"
)

type PatternKind string

const (
	KindViewPressure     PatternKind = "view_pressure"
	KindJealousy         PatternKind = "jealousy"
	KindFakeUrgency      PatternKind = "fake_urgency"
	KindPopularity       PatternKind = "popularity"
	KindRomanticPressure PatternKind = "romantic_pressure"
)

type ReviewedPattern struct {
	ID     string
	Kind   PatternKind
	Phrase string
}

type LocaleRules struct {
	Locale   string
	Patterns []ReviewedPattern
}

type CatalogSpec struct {
	ID          string
	Version     uint64
	Categories  []Category
	Locales     []LocaleRules
	ReviewID    string
	ReviewerKey string
	ReviewedAt  time.Time
}

type Catalog struct{ spec CatalogSpec }

func NewCatalog(spec CatalogSpec) (Catalog, error) {
	spec.Categories = append([]Category(nil), spec.Categories...)
	spec.Locales = cloneLocales(spec.Locales)
	spec.ReviewedAt = spec.ReviewedAt.UTC()
	if !tokenPattern.MatchString(spec.ID) || spec.Version == 0 ||
		len(spec.Categories) == 0 || !uniqueCategories(spec.Categories) ||
		len(spec.Locales) == 0 || !tokenPattern.MatchString(spec.ReviewID) ||
		!opaquePattern.MatchString(spec.ReviewerKey) || spec.ReviewedAt.IsZero() {
		return Catalog{}, ErrInvalidCatalog
	}
	slices.Sort(spec.Categories)
	seenLocales := map[string]struct{}{}
	for localeIndex := range spec.Locales {
		rules := &spec.Locales[localeIndex]
		if !localePattern.MatchString(rules.Locale) ||
			len(rules.Patterns) == 0 || len(rules.Patterns) > MaxPatternsPerLocale {
			return Catalog{}, ErrInvalidCatalog
		}
		if _, exists := seenLocales[rules.Locale]; exists {
			return Catalog{}, ErrInvalidCatalog
		}
		seenLocales[rules.Locale] = struct{}{}
		seenIDs := map[string]struct{}{}
		seenKinds := map[PatternKind]struct{}{}
		for patternIndex := range rules.Patterns {
			pattern := &rules.Patterns[patternIndex]
			pattern.Phrase = normalize(pattern.Phrase)
			if !tokenPattern.MatchString(pattern.ID) || !validKind(pattern.Kind) ||
				pattern.Phrase == "" || runeCount(pattern.Phrase) > 256 {
				return Catalog{}, ErrInvalidCatalog
			}
			if _, exists := seenIDs[pattern.ID]; exists {
				return Catalog{}, ErrInvalidCatalog
			}
			seenIDs[pattern.ID] = struct{}{}
			seenKinds[pattern.Kind] = struct{}{}
		}
		// Every locale must cover every prohibited pattern class. A partially
		// reviewed locale is not accepted as a safe catalog.
		if len(seenKinds) != 5 {
			return Catalog{}, ErrInvalidCatalog
		}
		slices.SortFunc(rules.Patterns, func(a, b ReviewedPattern) int {
			return strings.Compare(a.ID, b.ID)
		})
	}
	slices.SortFunc(spec.Locales, func(a, b LocaleRules) int {
		return strings.Compare(a.Locale, b.Locale)
	})
	return Catalog{spec: spec}, nil
}

func (catalog Catalog) Spec() CatalogSpec {
	spec := catalog.spec
	spec.Categories = append([]Category(nil), catalog.spec.Categories...)
	spec.Locales = cloneLocales(catalog.spec.Locales)
	return spec
}

type TemplateMetadata struct {
	TemplateName  string
	CampaignLabel string
	Tags          []string
}

type Notification struct {
	Locale   string
	Category Category
	Title    string
	Body     string
	Metadata TemplateMetadata
}

type Decision string

const (
	DecisionAllow Decision = "allow"
	DecisionDeny  Decision = "deny"
)

type Finding struct {
	PatternID string
	Kind      PatternKind
	Field     string
}

type Evaluation struct {
	CatalogID      string
	CatalogVersion uint64
	Decision       Decision
	Findings       []Finding
}

func Evaluate(catalog Catalog, notification Notification) (Evaluation, error) {
	if !localePattern.MatchString(notification.Locale) ||
		!validCategory(notification.Category) ||
		!containsCategory(catalog.spec.Categories, notification.Category) ||
		len(notification.Metadata.Tags) > MaxTags {
		return Evaluation{}, ErrInvalidNotification
	}
	fields := []struct {
		name  string
		value string
	}{
		{name: "title", value: notification.Title},
		{name: "body", value: notification.Body},
		{name: "template_name", value: notification.Metadata.TemplateName},
		{name: "campaign_label", value: notification.Metadata.CampaignLabel},
	}
	for _, tag := range notification.Metadata.Tags {
		fields = append(fields, struct {
			name  string
			value string
		}{name: "tag", value: tag})
	}
	for _, field := range fields {
		if runeCount(field.value) > MaxFieldRunes || !safeText(field.value) {
			return Evaluation{}, ErrInvalidNotification
		}
	}
	rules, ok := localeRules(catalog.spec.Locales, notification.Locale)
	if !ok {
		return Evaluation{}, ErrInvalidNotification
	}
	var findings []Finding
	for _, field := range fields {
		value := normalize(field.value)
		for _, pattern := range rules.Patterns {
			if strings.Contains(value, pattern.Phrase) {
				findings = append(findings, Finding{
					PatternID: pattern.ID, Kind: pattern.Kind, Field: field.name,
				})
			}
		}
	}
	slices.SortFunc(findings, func(a, b Finding) int {
		if comparison := strings.Compare(a.Field, b.Field); comparison != 0 {
			return comparison
		}
		return strings.Compare(a.PatternID, b.PatternID)
	})
	decision := DecisionAllow
	if len(findings) > 0 {
		decision = DecisionDeny
	}
	return Evaluation{
		CatalogID: catalog.spec.ID, CatalogVersion: catalog.spec.Version,
		Decision: decision, Findings: findings,
	}, nil
}

func normalize(value string) string {
	value = norm.NFKC.String(value)
	value = strings.Map(func(character rune) rune {
		if unicode.IsSpace(character) {
			return ' '
		}
		return unicode.ToLower(character)
	}, value)
	return strings.Join(strings.Fields(value), " ")
}

func safeText(value string) bool {
	for _, character := range value {
		if character == '\u0000' || character == '\uFFFD' ||
			unicode.Is(unicode.Cf, character) {
			return false
		}
	}
	return true
}

func validCategory(category Category) bool {
	return slices.Contains([]Category{
		CategoryTransactional, CategorySafety, CategoryRitual,
		CategoryCompanion, CategoryCourtship,
	}, category)
}

func uniqueCategories(categories []Category) bool {
	seen := map[Category]struct{}{}
	for _, category := range categories {
		if !validCategory(category) {
			return false
		}
		if _, exists := seen[category]; exists {
			return false
		}
		seen[category] = struct{}{}
	}
	return true
}

func validKind(kind PatternKind) bool {
	return slices.Contains([]PatternKind{
		KindViewPressure, KindJealousy, KindFakeUrgency,
		KindPopularity, KindRomanticPressure,
	}, kind)
}

func containsCategory(categories []Category, category Category) bool {
	return slices.Contains(categories, category)
}

func localeRules(locales []LocaleRules, locale string) (LocaleRules, bool) {
	for _, rules := range locales {
		if rules.Locale == locale {
			return rules, true
		}
	}
	return LocaleRules{}, false
}

func cloneLocales(locales []LocaleRules) []LocaleRules {
	result := append([]LocaleRules(nil), locales...)
	for index := range result {
		result[index].Patterns = append([]ReviewedPattern(nil), result[index].Patterns...)
	}
	return result
}

func runeCount(value string) int {
	return len([]rune(value))
}
