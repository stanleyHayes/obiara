package engagementpolicy

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func reviewedCatalog(t *testing.T) Catalog {
	t.Helper()
	catalog, err := NewCatalog(CatalogSpec{
		ID: "engagement.policy", Version: 4,
		Categories: []Category{
			CategoryTransactional, CategorySafety, CategoryRitual,
			CategoryCompanion, CategoryCourtship,
		},
		Locales: []LocaleRules{{
			Locale: "en",
			Patterns: []ReviewedPattern{
				{ID: "view.now", Kind: KindViewPressure, Phrase: "see who viewed you"},
				{ID: "jealous.miss", Kind: KindJealousy, Phrase: "make them jealous"},
				{ID: "urgency.last", Kind: KindFakeUrgency, Phrase: "last chance"},
				{ID: "popular.everyone", Kind: KindPopularity, Phrase: "everyone wants you"},
				{ID: "romance.reply", Kind: KindRomanticPressure, Phrase: "reply before they lose interest"},
			},
		}},
		ReviewID:    "review.engagement.4",
		ReviewerKey: strings.Repeat("a", 64),
		ReviewedAt:  time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func safeNotification() Notification {
	return Notification{
		Locale: "en", Category: CategoryTransactional,
		Title: "Your receipt is ready", Body: "Open Obiara to view your receipt.",
		Metadata: TemplateMetadata{
			TemplateName: "receipt_ready", CampaignLabel: "transactional_receipt",
			Tags: []string{"receipt", "account"},
		},
	}
}

func TestDeniesEveryPatternAcrossEveryCopySurface(t *testing.T) {
	catalog := reviewedCatalog(t)
	patterns := catalog.Spec().Locales[0].Patterns
	fields := []string{"title", "body", "template_name", "campaign_label", "tag"}
	for _, pattern := range patterns {
		for _, field := range fields {
			t.Run(pattern.ID+"/"+field, func(t *testing.T) {
				notification := safeNotification()
				switch field {
				case "title":
					notification.Title = pattern.Phrase
				case "body":
					notification.Body = pattern.Phrase
				case "template_name":
					notification.Metadata.TemplateName = pattern.Phrase
				case "campaign_label":
					notification.Metadata.CampaignLabel = pattern.Phrase
				case "tag":
					notification.Metadata.Tags = append(notification.Metadata.Tags, pattern.Phrase)
				}
				evaluation, err := Evaluate(catalog, notification)
				if err != nil {
					t.Fatal(err)
				}
				if evaluation.Decision != DecisionDeny || len(evaluation.Findings) == 0 {
					t.Fatalf("pattern allowed: %+v", evaluation)
				}
			})
		}
	}
}

func TestUnicodeCompatibilityCaseAndWhitespaceCannotBypass(t *testing.T) {
	tests := []string{
		"ＬＡＳＴ　ＣＨＡＮＣＥ",
		"Last\t\nChance",
		"laST cHaNcE",
	}
	for _, value := range tests {
		notification := safeNotification()
		notification.Title = value
		evaluation, err := Evaluate(reviewedCatalog(t), notification)
		if err != nil {
			t.Fatal(err)
		}
		if evaluation.Decision != DecisionDeny {
			t.Fatalf("normalized pattern bypassed by %q", value)
		}
	}
}

func TestUnknownLocaleAndCategoryFailClosed(t *testing.T) {
	notification := safeNotification()
	notification.Locale = "fr"
	if _, err := Evaluate(reviewedCatalog(t), notification); err == nil {
		t.Fatal("unknown locale allowed")
	}
	notification = safeNotification()
	notification.Category = "growth"
	if _, err := Evaluate(reviewedCatalog(t), notification); err == nil {
		t.Fatal("unknown category allowed")
	}
}

func TestCatalogRequiresReviewedVersionAndCompleteKinds(t *testing.T) {
	spec := reviewedCatalog(t).Spec()
	spec.Version = 0
	if _, err := NewCatalog(spec); err == nil {
		t.Fatal("unversioned catalog accepted")
	}
	spec = reviewedCatalog(t).Spec()
	spec.Locales[0].Patterns = spec.Locales[0].Patterns[:4]
	if _, err := NewCatalog(spec); err == nil {
		t.Fatal("partially reviewed locale accepted")
	}
	spec = reviewedCatalog(t).Spec()
	spec.ReviewerKey = ""
	if _, err := NewCatalog(spec); err == nil {
		t.Fatal("unreviewed catalog accepted")
	}
}

func TestCatalogIsDefensivelyCopiedAndEvaluationDeterministic(t *testing.T) {
	catalog := reviewedCatalog(t)
	first := catalog.Spec()
	first.Locales[0].Patterns[0].Phrase = "mutated"
	first.Categories[0] = "growth"
	second := catalog.Spec()
	if second.Locales[0].Patterns[0].Phrase == "mutated" || second.Categories[0] == "growth" {
		t.Fatal("catalog leaked mutable backing storage")
	}
	notification := safeNotification()
	notification.Body = "LAST CHANCE. Make them jealous."
	want, err := Evaluate(catalog, notification)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 1000; index++ {
		got, err := Evaluate(catalog, notification)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("non-deterministic evaluation at %d: (%+v, %v)", index, got, err)
		}
	}
}

func TestPatternAndFieldOrderDoNotChangeFindings(t *testing.T) {
	base := reviewedCatalog(t).Spec()
	notification := safeNotification()
	notification.Title = "Last chance"
	notification.Body = "Everyone wants you"
	want, err := Evaluate(reviewedCatalog(t), notification)
	if err != nil {
		t.Fatal(err)
	}
	random := rand.New(rand.NewSource(42))
	for index := 0; index < 500; index++ {
		spec := base
		spec.Locales = cloneLocales(base.Locales)
		random.Shuffle(len(spec.Locales[0].Patterns), func(i, j int) {
			spec.Locales[0].Patterns[i], spec.Locales[0].Patterns[j] =
				spec.Locales[0].Patterns[j], spec.Locales[0].Patterns[i]
		})
		catalog, err := NewCatalog(spec)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Evaluate(catalog, notification)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("order changed result: (%+v, %v)", got, err)
		}
	}
}

func TestEvaluationExposesNoScoreOrMemberInference(t *testing.T) {
	notification := safeNotification()
	notification.Body = "last chance"
	evaluation, err := Evaluate(reviewedCatalog(t), notification)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(evaluation)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(payload))
	for _, forbidden := range []string{
		"score", "rank", "member", "recipient", "attractiveness",
		"dispatch", "quiet", "cap", "model", "vendor",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("forbidden output field %q in %s", forbidden, payload)
		}
	}
}

func TestAllowHasNoFindings(t *testing.T) {
	evaluation, err := Evaluate(reviewedCatalog(t), safeNotification())
	if err != nil {
		t.Fatal(err)
	}
	if evaluation.Decision != DecisionAllow || len(evaluation.Findings) != 0 {
		t.Fatalf("safe copy rejected: %+v", evaluation)
	}
}

func FuzzNormalizedBannedPhraseAlwaysDenied(f *testing.F) {
	f.Add("LAST", "CHANCE", " ")
	f.Add("last", "chance", "\t")
	f.Fuzz(func(t *testing.T, left, right, separator string) {
		// Exercise arbitrary Unicode around a stable prohibited phrase while
		// keeping the input bounded and valid.
		left = truncate(left, 32)
		right = truncate(right, 32)
		separator = truncate(separator, 8)
		notification := safeNotification()
		notification.Title = left + " ＬＡＳＴ　ＣＨＡＮＣＥ " + separator + right
		if !safeText(notification.Title) {
			t.Skip()
		}
		evaluation, err := Evaluate(reviewedCatalog(t), notification)
		if err != nil {
			t.Fatal(err)
		}
		if evaluation.Decision != DecisionDeny {
			t.Fatalf("banned phrase bypassed: %q", notification.Title)
		}
	})
}

func truncate(value string, maximum int) string {
	runes := []rune(value)
	if len(runes) > maximum {
		runes = runes[:maximum]
	}
	return string(runes)
}

func TestFindingsAreCanonical(t *testing.T) {
	notification := safeNotification()
	notification.Title = "last chance and everyone wants you"
	evaluation, err := Evaluate(reviewedCatalog(t), notification)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.IsSortedFunc(evaluation.Findings, func(a, b Finding) int {
		if comparison := strings.Compare(a.Field, b.Field); comparison != 0 {
			return comparison
		}
		return strings.Compare(a.PatternID, b.PatternID)
	}) {
		t.Fatal("findings not canonical")
	}
}
