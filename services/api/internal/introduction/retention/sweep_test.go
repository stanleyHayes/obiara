package retention

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/application"
	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

var swept = time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func introduction(t *testing.T, id, assetID string) domain.Introduction {
	t.Helper()
	digest := strings.Repeat("a", 64)
	consent, err := domain.NewConsentSnapshot("voice.introduction", 1, swept)
	if err != nil {
		t.Fatal(err)
	}
	media, err := domain.NewMediaRef(assetID, "audio/ogg", 1024, 42*time.Second, digest)
	if err != nil {
		t.Fatal(err)
	}
	value, err := domain.New(id, "member_1", domain.PromptArrival, consent, media,
		domain.NewRetention(time.Time{}, false),
		domain.Command{ID: "cmd_" + id, Fingerprint: digest, At: swept})
	if err != nil {
		t.Fatal(err)
	}
	return value
}

type stubStore struct {
	due []domain.Introduction
	err error
}

func (s stubStore) DueForPurge(context.Context, time.Time, int) ([]domain.Introduction, error) {
	return s.due, s.err
}

type stubEraser struct {
	erased []string
	fail   map[string]error
}

func (e *stubEraser) Erase(_ context.Context, assetID string) error {
	if err, ok := e.fail[assetID]; ok {
		return err
	}
	e.erased = append(e.erased, assetID)
	return nil
}

type stubPurger struct {
	purged []string
	err    error
}

func (p *stubPurger) Purge(_ context.Context, id, _ string) (domain.Introduction, error) {
	if p.err != nil {
		return domain.Introduction{}, p.err
	}
	p.purged = append(p.purged, id)
	return domain.Introduction{}, nil
}

func TestTheSweepErasesBytesBeforeItRecordsThePurge(t *testing.T) {
	// Recording a purge before the audio is gone would log an erasure that
	// never happened, and nothing afterwards would look for those bytes.
	store := stubStore{due: []domain.Introduction{
		introduction(t, "introduction_1", "asset_1"),
		introduction(t, "introduction_2", "asset_2"),
	}}
	eraser := &stubEraser{}
	purger := &stubPurger{}

	count, err := NewSweeper(store, eraser, purger, func() time.Time { return swept }, quietLogger()).
		Once(context.Background())
	if err != nil {
		t.Fatalf("Once = %v", err)
	}
	if count != 2 {
		t.Fatalf("purged %d, want 2", count)
	}
	if len(eraser.erased) != 2 || len(purger.purged) != 2 {
		t.Fatalf("erased=%v purged=%v", eraser.erased, purger.purged)
	}
}

func TestOneRecordingThatWillNotEraseDoesNotStopTheRest(t *testing.T) {
	// This is also how a legal hold is honoured: the eraser refuses it, the
	// aggregate keeps purge_pending, and the next pass tries again.
	store := stubStore{due: []domain.Introduction{
		introduction(t, "introduction_held", "asset_held"),
		introduction(t, "introduction_ok", "asset_ok"),
	}}
	eraser := &stubEraser{fail: map[string]error{
		"asset_held": errors.New("asset is under legal hold"),
	}}
	purger := &stubPurger{}

	count, err := NewSweeper(store, eraser, purger, func() time.Time { return swept }, quietLogger()).
		Once(context.Background())
	if err != nil {
		t.Fatalf("Once = %v", err)
	}
	if count != 1 {
		t.Fatalf("purged %d, want 1", count)
	}
	if len(purger.purged) != 1 || purger.purged[0] != "introduction_ok" {
		t.Fatalf("purged = %v", purger.purged)
	}
	// The held recording's bytes must still be there.
	for _, asset := range eraser.erased {
		if asset == "asset_held" {
			t.Fatal("a held recording was erased")
		}
	}
}

func TestAReplayedPurgeCountsAsDone(t *testing.T) {
	// The command id is derived from the aggregate, so a retried sweep replays
	// rather than opening a second purge. Treating the replay as a failure
	// would keep the recording in the queue forever.
	store := stubStore{due: []domain.Introduction{introduction(t, "introduction_1", "asset_1")}}
	eraser := &stubEraser{}
	purger := &stubPurger{err: application.ErrCommandAlreadyUsed}

	count, err := NewSweeper(store, eraser, purger, func() time.Time { return swept }, quietLogger()).
		Once(context.Background())
	if err != nil {
		t.Fatalf("Once = %v", err)
	}
	if count != 1 {
		t.Fatalf("a replayed purge counted %d, want 1", count)
	}
}

func TestAnUnreadableQueueIsReportedNotSwallowed(t *testing.T) {
	store := stubStore{err: errors.New("mongo unavailable")}
	_, err := NewSweeper(store, &stubEraser{}, &stubPurger{}, func() time.Time { return swept }, quietLogger()).
		Once(context.Background())
	if err == nil {
		t.Fatal("a failing queue read was reported as a clean sweep")
	}
}
