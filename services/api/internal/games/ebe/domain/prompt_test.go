package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestOnlyVersionedReviewerApprovedPromptsEnterCatalog(t *testing.T) {
	spec := approvedPromptSpec("prompt-one", 1, "ak", "Synthetic reviewed cue")
	prompt, err := NewApprovedPrompt(spec)
	if err != nil {
		t.Fatal(err)
	}
	if prompt.Spec().Version != 1 || prompt.Spec().Language != "ak" || prompt.Spec().Source.Citation == "" {
		t.Fatalf("provenance lost: %+v", prompt.Spec())
	}

	unapproved := spec
	unapproved.Review.Decision = "pending"
	if _, err = NewApprovedPrompt(unapproved); !errors.Is(err, ErrPromptUnapproved) {
		t.Fatalf("pending review = %v", err)
	}
	unversioned := spec
	unversioned.Version = 0
	if _, err = NewApprovedPrompt(unversioned); !errors.Is(err, ErrPromptInvalid) {
		t.Fatalf("unversioned prompt = %v", err)
	}
	noSource := spec
	noSource.Source = Source{}
	if _, err = NewApprovedPrompt(noSource); !errors.Is(err, ErrPromptInvalid) {
		t.Fatalf("missing source = %v", err)
	}
}

func TestPromptDigestBindsVersionLanguageSourceAndReview(t *testing.T) {
	base := approvedPromptSpec("prompt-one", 1, "ak", "Synthetic reviewed cue")
	original := mustPrompt(t, base)
	mutations := []PromptSpec{
		func() PromptSpec { value := base; value.Version++; return value }(),
		func() PromptSpec { value := base; value.Language = "en"; return value }(),
		func() PromptSpec {
			value := base
			value.Source.Citation = "Different test catalogue record"
			return value
		}(),
		func() PromptSpec { value := base; value.Review.ID = "review-two"; return value }(),
	}
	for _, mutation := range mutations {
		if digest := mustPrompt(t, mutation).Digest(); digest == original.Digest() {
			t.Fatalf("provenance mutation did not change digest: %+v", mutation)
		}
	}
}

func TestAnswersAreBoundedAndNormalizedWithoutBecomingContent(t *testing.T) {
	prompt := mustPrompt(t, approvedPromptSpec("prompt-one", 1, "ak", "Synthetic reviewed cue"))
	for _, answer := range []string{"  TEST   ANSWER  ", "test answer"} {
		accepted, err := prompt.Accepts(answer)
		if err != nil || !accepted {
			t.Fatalf("answer %q accepted=%v err=%v", answer, accepted, err)
		}
	}
	tooLong := strings.Repeat("a", MaxAnswerRunes+1)
	if _, err := prompt.Accepts(tooLong); !errors.Is(err, ErrAnswerInvalid) {
		t.Fatalf("oversized answer = %v", err)
	}
}

func FuzzPromptAnswerBounds(f *testing.F) {
	f.Add("test answer")
	f.Add(" TEST   ANSWER ")
	f.Add(strings.Repeat("x", MaxAnswerRunes+1))
	prompt := mustPrompt(f, approvedPromptSpec("prompt-one", 1, "ak", "Synthetic reviewed cue"))
	f.Fuzz(func(t *testing.T, answer string) {
		accepted, err := prompt.Accepts(answer)
		if !boundedText(answer, MaxAnswerRunes) {
			if !errors.Is(err, ErrAnswerInvalid) {
				t.Fatalf("invalid answer accepted=%v err=%v", accepted, err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if accepted && normalizeAnswer(answer) != "test answer" {
			t.Fatalf("unexpected accepted form %q", answer)
		}
	})
}

func approvedPromptSpec(id string, version uint64, language, cue string) PromptSpec {
	return PromptSpec{
		ID:              id,
		Version:         version,
		Language:        language,
		Cue:             cue,
		AcceptedAnswers: []string{"test answer", "reviewed variant"},
		Source: Source{
			Kind:     SourceInstitutionalArchive,
			Citation: "Synthetic test catalogue record; not cultural authority",
			Locator:  "https://example.invalid/test-catalogue/record",
		},
		Review: Review{
			ID:          "review-one",
			ReviewerKey: strings.Repeat("a", 64),
			Decision:    DecisionApproved,
			ReviewedAt:  time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC),
		},
	}
}

type testingT interface {
	Helper()
	Fatal(...any)
}

func mustPrompt(t testingT, spec PromptSpec) Prompt {
	t.Helper()
	prompt, err := NewApprovedPrompt(spec)
	if err != nil {
		t.Fatal(err)
	}
	return prompt
}
