package mongodb

import (
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

// A store that cannot restore what it wrote is worse than no store: the
// aggregate comes back in a plausible-looking wrong state and the invariants
// that would have caught it were already spent on the way in. This exercises
// the mapping in both directions without a database.
func TestDocumentMappingRestoresEveryFieldItPersists(t *testing.T) {
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)
	consent, err := domain.NewConsentSnapshot("voice.introduction", 1, now)
	if err != nil {
		t.Fatal(err)
	}
	media, err := domain.NewMediaRef("intro_asset_1", "audio/ogg", 184_320, 42*time.Second, digest)
	if err != nil {
		t.Fatal(err)
	}
	original, err := domain.New(
		"introduction_1", "member_1", consent, media,
		domain.NewRetention(now.Add(180*24*time.Hour), false),
		domain.Command{ID: "cmd_1", Fingerprint: digest, At: now},
	)
	if err != nil {
		t.Fatal(err)
	}

	restored := fromDocument(toDocument(original))

	if restored.ID() != original.ID() || restored.OwnerID() != original.OwnerID() {
		t.Fatalf("identity lost: %q/%q", restored.ID(), restored.OwnerID())
	}
	if restored.Status() != original.Status() || restored.DataStatus() != original.DataStatus() {
		t.Fatalf("state lost: %v/%v", restored.Status(), restored.DataStatus())
	}
	if restored.Version() != original.Version() {
		t.Fatalf("version = %d, want %d", restored.Version(), original.Version())
	}
	// Duration is the number the listening gate counts against; a lossy
	// round trip here would silently shorten every recording.
	if restored.Media().Duration() != 42*time.Second {
		t.Fatalf("duration = %v, want 42s", restored.Media().Duration())
	}
	if restored.Media().Checksum() != digest || restored.Media().Size() != 184_320 {
		t.Fatalf("media ref lost: %+v", restored.Media())
	}
	if restored.Consent().PurposeID() != "voice.introduction" || restored.Consent().Version() != 1 {
		t.Fatalf("consent snapshot lost: %+v", restored.Consent())
	}
	if !restored.Retention().Until().Equal(now.Add(180 * 24 * time.Hour)) {
		t.Fatalf("retention lost: %v", restored.Retention().Until())
	}
	// Replay safety is decided from the event history; an aggregate restored
	// without it would accept a command it has already applied.
	if len(restored.Events()) != 1 || !restored.HasCommand("cmd_1") {
		t.Fatalf("history lost: %+v", restored.Events())
	}
}

func TestCreationCommandIsWhatMakesCreateIdempotent(t *testing.T) {
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	digest := strings.Repeat("a", 64)
	consent, _ := domain.NewConsentSnapshot("voice.introduction", 1, now)
	media, _ := domain.NewMediaRef("intro_asset_1", "audio/ogg", 1024, time.Second, digest)
	introduction, err := domain.New(
		"introduction_1", "member_1", consent, media,
		domain.NewRetention(time.Time{}, false),
		domain.Command{ID: "cmd_begin", Fingerprint: digest, At: now},
	)
	if err != nil {
		t.Fatal(err)
	}
	// The unique index is on this field. If it were ever the aggregate id
	// instead, a retried BeginUpload would open a second recording.
	if got := toDocument(introduction).CreationCommandID; got != "cmd_begin" {
		t.Fatalf("CreationCommandID = %q, want the creating command", got)
	}
}
