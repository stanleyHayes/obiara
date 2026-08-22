package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const goodToken = "ExponentPushToken[xxxxxxxxxxxxxxxxxxxxxx]"

func TestNewRegistrationValidatesTheToken(t *testing.T) {
	now := time.Date(2026, 8, 22, 6, 0, 0, 0, time.UTC)

	t.Run("accepts both documented shapes", func(t *testing.T) {
		for _, token := range []string{
			"ExponentPushToken[abc123]",
			"ExpoPushToken[abc123]",
		} {
			if _, err := NewRegistration("mem_1", token, PlatformIOS, now); err != nil {
				t.Errorf("NewRegistration(%q) = %v", token, err)
			}
		}
	})

	// A malformed token cannot be delivered to, and Expo rejects whole
	// batches that contain one, so it must not reach the registry.
	t.Run("rejects malformed tokens", func(t *testing.T) {
		for _, token := range []string{
			"", "abc123", "ExponentPushToken[]", "ExponentPushToken[abc",
			"fcm:abc123", strings.Repeat("x", 200),
		} {
			if _, err := NewRegistration("mem_1", token, PlatformIOS, now); !errors.Is(err, ErrTokenInvalid) {
				t.Errorf("NewRegistration(%q) = %v, want ErrTokenInvalid", token, err)
			}
		}
	})

	t.Run("requires a member", func(t *testing.T) {
		if _, err := NewRegistration("  ", goodToken, PlatformIOS, now); !errors.Is(err, ErrMemberRequired) {
			t.Errorf("error = %v, want ErrMemberRequired", err)
		}
	})

	t.Run("rejects an unknown platform", func(t *testing.T) {
		if _, err := NewRegistration("mem_1", goodToken, Platform("blackberry"), now); !errors.Is(err, ErrPlatformInvalid) {
			t.Errorf("error = %v, want ErrPlatformInvalid", err)
		}
	})
}

// TestCopyNeverNamesAnyone pins FR-701 for the lock screen: a push preview is
// readable by anyone holding the phone, so no template may carry a
// counterpart, a room, or what the member is doing.
func TestCopyNeverNamesAnyone(t *testing.T) {
	forbidden := []string{"%s", "{", "}", "matched", "message from"}
	for name := range templates {
		copy, err := CopyFor(name)
		if err != nil {
			t.Fatalf("CopyFor(%q): %v", name, err)
		}
		if copy.Title == "" || copy.Body == "" {
			t.Errorf("template %q has empty copy", name)
		}
		for _, needle := range forbidden {
			if strings.Contains(strings.ToLower(copy.Body), needle) {
				t.Errorf("template %q body contains %q; push copy must be parameterless", name, needle)
			}
		}
	}
}

func TestCopyForUnknownTemplate(t *testing.T) {
	if _, err := CopyFor("does_not_exist"); !errors.Is(err, ErrTemplateUnknown) {
		t.Errorf("error = %v, want ErrTemplateUnknown", err)
	}
	// The fallback must still be safe to show on a lock screen.
	fallback := FallbackCopy()
	if fallback.Title == "" || fallback.Body == "" {
		t.Error("fallback copy is empty")
	}
}
