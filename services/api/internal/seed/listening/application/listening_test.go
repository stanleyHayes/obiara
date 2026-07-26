package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/listening/domain"
)

var listeningNow = time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)

func fixedNow() time.Time { return listeningNow }

func TestRecordHeartbeatsCreatesAndSaves(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockListeningRepository(ctrl)
	service := NewListeningService(repository, fixedNow)

	repository.EXPECT().Find(gomock.Any(), "m-1", "asset-1").Return(domain.Playback{}, domain.ErrPlaybackNotFound)
	repository.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, playback domain.Playback) error {
			if playback.TotalSeconds() != 15 || playback.Version() != 1 {
				t.Fatalf("playback total=%v version=%d", playback.TotalSeconds(), playback.Version())
			}
			return nil
		})

	record, err := service.RecordHeartbeats(context.Background(), "m-1", "asset-1", 60, []HeartbeatRange{{Start: 0, End: 10}, {Start: 10, End: 15}})
	if err != nil {
		t.Fatal(err)
	}
	if record.TotalSeconds() != 15 {
		t.Fatalf("total = %v", record.TotalSeconds())
	}
}

func TestRecordHeartbeatsRetriesStaleWrites(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockListeningRepository(ctrl)
	service := NewListeningService(repository, fixedNow)

	existing, _ := domain.NewPlayback("m-1", "asset-1", 60)
	_ = existing.Record(0, 10)
	existing.Commit(listeningNow)

	gomock.InOrder(
		repository.EXPECT().Find(gomock.Any(), "m-1", "asset-1").Return(existing, nil),
		repository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(domain.ErrStalePlayback),
		repository.EXPECT().Find(gomock.Any(), "m-1", "asset-1").Return(existing, nil),
		repository.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil),
	)

	if _, err := service.RecordHeartbeats(context.Background(), "m-1", "asset-1", 60, []HeartbeatRange{{Start: 10, End: 25}}); err != nil {
		t.Fatal(err)
	}
}

func TestEligibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockListeningRepository(ctrl)
	service := NewListeningService(repository, fixedNow)

	// No record yet: ineligible with zero seconds, no error.
	repository.EXPECT().Find(gomock.Any(), "m-1", "asset-1").Return(domain.Playback{}, domain.ErrPlaybackNotFound)
	eligible, total, err := service.Eligibility(context.Background(), "m-1", "asset-1")
	if err != nil || eligible || total != 0 {
		t.Fatalf("empty eligibility = %v, %v, %v", eligible, total, err)
	}

	record, _ := domain.NewPlayback("m-1", "asset-1", 60)
	_ = record.Record(0, 30)
	repository.EXPECT().Find(gomock.Any(), "m-1", "asset-1").Return(record, nil)
	eligible, total, err = service.Eligibility(context.Background(), "m-1", "asset-1")
	if err != nil || !eligible || total != 30 {
		t.Fatalf("eligibility = %v, %v, %v", eligible, total, err)
	}
}

func TestInvalidHeartbeatRejected(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockListeningRepository(ctrl)
	service := NewListeningService(repository, fixedNow)

	repository.EXPECT().Find(gomock.Any(), "m-1", "asset-1").Return(domain.Playback{}, domain.ErrPlaybackNotFound)
	// No Save expectation: invalid input must not persist.
	if _, err := service.RecordHeartbeats(context.Background(), "m-1", "asset-1", 60, []HeartbeatRange{{Start: 9, End: 4}}); err != domain.ErrInvalidRange {
		t.Fatalf("RecordHeartbeats = %v, want invalid range", err)
	}
}
