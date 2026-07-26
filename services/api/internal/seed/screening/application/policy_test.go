package application

import (
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestNormalizeMultilingualTextWithoutTranslation(t *testing.T) {
	tests := map[string]string{
		"  hello   there ":  "hello there",
		"  Meda   wo ase  ": "Meda wo ase",
		"  Oyiwaladɔŋŋ \n ": "Oyiwaladɔŋŋ",
		" Ｈｅｌｌｏ ":           "Hello",
	}
	for input, want := range tests {
		got, err := normalizeText(input)
		if err != nil || got != want {
			t.Fatalf("normalize %q = %q, %v", input, got, err)
		}
	}
}

func TestNormalizationPropertyAcrossReviewedLanguageSamples(t *testing.T) {
	samples := []string{"Meda wo ase", "Oyiwaladɔŋŋ", "Akpe", "Thank you", "Ɛyɛ"}
	for trial := 0; trial < 1000; trial++ {
		for _, sample := range samples {
			input := strings.Repeat(" ", trial%4) + strings.ReplaceAll(sample, " ", strings.Repeat(" ", trial%5+1))
			got, err := normalizeText(input)
			if err != nil || got != sample {
				t.Fatalf("trial %d sample %q = %q, %v", trial, sample, got, err)
			}
		}
	}
}

func TestPureNormalizationIsRaceSafe(t *testing.T) {
	const input = "  Meda   wo ase  "
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				got, err := normalizeText(input)
				if err != nil || got != "Meda wo ase" {
					t.Errorf("normalize = %q, %v", got, err)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzNormalizeIsIdempotentAndBounded(f *testing.F) {
	f.Add("Meda wo ase")
	f.Add(" Oyiwaladɔŋŋ ")
	f.Add(strings.Repeat("x", MaxTextRunes+1))
	f.Fuzz(func(t *testing.T, input string) {
		normalized, err := normalizeText(input)
		if err != nil {
			return
		}
		second, secondErr := normalizeText(normalized)
		if secondErr != nil || second != normalized {
			t.Fatalf("not idempotent: %q => %q => %q (%v)", input, normalized, second, secondErr)
		}
		if len([]byte(normalized)) > MaxTextBytes || len([]rune(normalized)) > MaxTextRunes {
			t.Fatalf("normalization exceeded bounds")
		}
		if reflect.DeepEqual([]byte(input), []byte{}) {
			t.Fatal("empty input should not normalize")
		}
	})
}
