package domain

import (
	"testing"
	"time"
)

var accra = time.FixedZone("Africa/Accra", 0)

func TestDueAtCalendar(t *testing.T) {
	day := time.Date(2026, time.July, 27, 0, 0, 0, 0, accra) // a Monday

	dawn, ok := DueAt(KindDawn, day, accra)
	if !ok || dawn.Hour() != 6 {
		t.Fatalf("dawn = %v, %v", dawn, ok)
	}
	monday, ok := DueAt(KindMonday, day, accra)
	if !ok || monday.Hour() != 8 || monday.Weekday() != time.Monday {
		t.Fatalf("monday = %v, %v", monday, ok)
	}
	if _, ok := DueAt(KindSunday, day, accra); ok {
		t.Fatal("sunday ritual must not be due on a Monday")
	}
	if _, ok := DueAt(KindFireHerald, day, accra); ok {
		t.Fatal("fire herald is not a calendar ritual")
	}

	sunday := time.Date(2026, time.August, 2, 0, 0, 0, 0, accra)
	reflection, ok := DueAt(KindSunday, sunday, accra)
	if !ok || reflection.Hour() != 18 {
		t.Fatalf("sunday = %v, %v", reflection, ok)
	}
}

func TestDueAtHonorsMemberTimezone(t *testing.T) {
	apia, _ := time.LoadLocation("Pacific/Apia")
	day := time.Date(2026, time.July, 27, 0, 0, 0, 0, time.UTC)
	accraDue, _ := DueAt(KindDawn, day, accra)
	apiaDue, _ := DueAt(KindDawn, day, apia)
	if !apiaDue.Before(accraDue) {
		t.Fatalf("apia dawn %v should precede accra dawn %v in UTC", apiaDue, accraDue)
	}
}

func TestHeraldAt(t *testing.T) {
	start := time.Date(2026, time.July, 31, 20, 0, 0, 0, time.UTC)
	if herald := HeraldAt(start); herald.Hour() != 19 || herald.Minute() != 15 {
		t.Fatalf("herald = %v, want 19:15", herald)
	}
}

func TestDedupKey(t *testing.T) {
	if key := DedupKey("m-1", KindDawn, "2026-07-27"); key != "m-1|dawn|2026-07-27" {
		t.Fatalf("key = %q", key)
	}
}
