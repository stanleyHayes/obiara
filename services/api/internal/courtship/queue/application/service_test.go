package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/courtship/queue/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestSubmitUsesOpaqueKeysAndExpectedRevision(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Now()
	keyer.EXPECT().Key("room", "room").Return(key(1), nil)
	keyer.EXPECT().Key("device", "device").Return(key(2), nil)
	keyer.EXPECT().Key("actor", "actor").Return(key(3), nil)
	keyer.EXPECT().Key("payload", "payload").Return(key(4), nil)
	state, _ := domain.Open(key(1))
	repository.EXPECT().State(gomock.Any(), key(1)).Return(state, nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), gomock.Any(), uint64(0)).DoAndReturn(func(_ context.Context, next domain.State, event domain.Event, _ uint64) (domain.Event, bool, error) {
		if event.RoomKey != key(1) || event.DeviceKey != key(2) || event.Sequence != 1 {
			t.Fatalf("event=%#v", event)
		}
		return event, false, nil
	})
	service := New(repository, keyer, func() time.Time { return now })
	result, err := service.Submit(context.Background(), SubmitCommand{"command", "room", "device", "actor", "payload", 0})
	if err != nil || result.Event.Sequence != 1 {
		t.Fatal(result, err)
	}
}
func TestOfflineRetryFromStaleCursorReturnsOriginalEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository, keyer := NewMockRepository(ctrl), NewMockKeyer(ctrl)
	for namespace, value := range map[string]string{"room": "room", "device": "device", "actor": "actor", "payload": "payload"} {
		index := map[string]int{"room": 1, "device": 2, "actor": 3, "payload": 4}[namespace]
		keyer.EXPECT().Key(namespace, value).Return(key(index), nil)
	}
	state, _ := domain.Rehydrate(key(1), 1, 1)
	repository.EXPECT().State(gomock.Any(), key(1)).Return(state, nil)
	fp := fingerprint("command", key(1), key(2), key(3), key(4), 0)
	original := domain.Event{RoomKey: key(1), Sequence: 1, CommandID: "command", Fingerprint: fp}
	repository.EXPECT().EventByCommand(gomock.Any(), key(1), "command").Return(original, nil)
	service := New(repository, keyer, time.Now)
	result, err := service.Submit(context.Background(), SubmitCommand{"command", "room", "device", "actor", "payload", 0})
	if err != nil || !result.Replayed || result.Event.Sequence != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}
func key(n int) string {
	const hex = "0000000000000000000000000000000000000000000000000000000000000000"
	b := []byte(hex)
	b[len(b)-1] = byte('0' + n)
	return string(b)
}
