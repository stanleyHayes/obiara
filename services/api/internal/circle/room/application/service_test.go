package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
	"time"
)

func TestVoiceRequiresPostAndEventsRequireHost(t *testing.T) {
	for _, x := range []struct {
		voice bool
		cap   Capability
	}{{true, CapabilityPost}, {false, CapabilityHost}} {
		c := gomock.NewController(t)
		a := NewMockAuthorizer(c)
		r := NewMockRepository(c)
		k := NewMockKeyer(c)
		i := NewMockIDs(c)
		a.EXPECT().Authorize(gomock.Any(), Decision{"circle:1", "member:1", x.cap}).Return(nil)
		k.EXPECT().Key("circle_room_actor", "member:1").Return(strings.Repeat("a", 64), nil)
		i.EXPECT().NewID().Return("entry:1")
		r.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, e domain.Entry) (domain.Entry, bool, error) { return e, false, nil })
		n := time.Now()
		q := Create{CommandID: "cmd:1", CircleID: "circle:1", ActorID: "member:1", Retention: time.Hour}
		var err error
		if x.voice {
			q.Media, _ = domain.NewMediaRef("asset:1", "", "audio/ogg", time.Minute)
			_, err = NewService(a, r, k, i, func() time.Time { return n }).Voice(context.Background(), q)
		} else {
			q.ContentRef = "content:1"
			q.StartsAt = n.Add(time.Hour)
			q.EndsAt = n.Add(2 * time.Hour)
			_, err = NewService(a, r, k, i, func() time.Time { return n }).Event(context.Background(), q)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}
