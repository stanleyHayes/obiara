package application

import (
	"context"
	"github.com/stanleyHayes/obiara/services/api/internal/cloth/ceremony/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestPublishRevalidatesImmediatelyBeforeFixedNeutralPublish(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepository(ctrl)
	keyer := NewMockKeyer(ctrl)
	ids := NewMockIDSource(ctrl)
	validator := NewMockPublishRevalidator(ctrl)
	publisher := NewMockCirclePublisher(ctrl)
	now := time.Now()
	x, _ := domain.Open("ceremony", [2]string{key(1), key(2)}, domain.Command{ID: "open", ActorKey: key(1), Fingerprint: key(9), At: now})
	x, _ = x.Confirm(command("c1", key(1), domain.ActionConfirmed, "", 1, now))
	x, _ = x.Confirm(command("c2", key(2), domain.ActionConfirmed, "", 2, now))
	x, _ = x.ProposeAnnouncement(key(3), command("p", key(1), domain.ActionAnnouncementProposed, key(3), 3, now))
	x, _ = x.ConsentAnnouncement(command("a1", key(1), domain.ActionAnnouncementConsented, key(3), 4, now))
	x, _ = x.ConsentAnnouncement(command("a2", key(2), domain.ActionAnnouncementConsented, key(3), 5, now))
	repo.EXPECT().Find(gomock.Any(), "ceremony").Return(x, nil)
	keyer.EXPECT().Key("member", "alice").Return(key(1), nil)
	validator.EXPECT().Authorize(gomock.Any(), x.State().Members, key(3)).Return(nil)
	publisher.EXPECT().Publish(gomock.Any(), PublishRequest{"publish", key(3), domain.AnnouncementKind}).Return(nil)
	repo.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(6), "publish").Return(nil)
	service := New(repo, keyer, ids, validator, publisher, func() time.Time { return now })
	if _, err := service.Publish(context.Background(), Command{"publish", "ceremony", "alice", "", 6}); err != nil {
		t.Fatal(err)
	}
}
func command(id, actor string, action domain.Action, destination string, revision uint64, at time.Time) domain.Command {
	return domain.Command{ID: id, ActorKey: actor, Fingerprint: domain.Fingerprint("ceremony", action, actor, destination, revision), ExpectedRevision: revision, At: at}
}
func key(n int) string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = '0'
	}
	b[63] = byte('0' + n)
	return string(b)
}
