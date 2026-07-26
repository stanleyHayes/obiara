package domain

import (
	"strings"
	"testing"
	"testing/quick"
	"time"
)

func TestExclusionEndsAtExactNinetiethDay(t *testing.T) {
	at := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	decline, err := New("decline-1", key("a"), key("b"), key("c"), "command-1", at)
	if err != nil {
		t.Fatal(err)
	}
	if !decline.Excludes(at) || !decline.Excludes(at.Add(ExclusionPeriod-time.Nanosecond)) {
		t.Fatal("decline must exclude throughout the half-open 90-day window")
	}
	if decline.Excludes(at.Add(ExclusionPeriod)) || decline.Excludes(at.Add(ExclusionPeriod+time.Nanosecond)) {
		t.Fatal("decline must become eligible at the exact 90-day boundary")
	}
}

func TestExclusionPropertyMatchesHalfOpenWindow(t *testing.T) {
	base := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	decline, _ := New("decline-1", key("a"), key("b"), key("c"), "command-1", base)
	property := func(raw int64) bool {
		const span = int64(200 * 24 * time.Hour)
		offset := time.Duration(raw % span)
		at := base.Add(offset)
		want := offset >= 0 && offset < ExclusionPeriod
		return decline.Excludes(at) == want
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 10_000}); err != nil {
		t.Fatal(err)
	}
}

func TestNotificationHasOnlyNeutralKindAndOpaqueKeys(t *testing.T) {
	notification, err := NewNotification(key("d"), key("c"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if notification.Kind() != NotificationKind || strings.Contains(notification.Kind(), "declin") ||
		strings.Contains(notification.Kind(), "reject") {
		t.Fatalf("unsafe notification kind %q", notification.Kind())
	}
}

func key(value string) string {
	return strings.Repeat(value, 64)
}
