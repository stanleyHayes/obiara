package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

type ids struct{}

func (ids) NewID() string { return "pod-1" }
func key(n int) string    { return fmt.Sprintf("%064x", n) }
func fixture(t *testing.T) (*MockRepository, *MockAuthorizer, *MockPlaybackEligibility, *MockMediaIssuer, *MockKeyer, Service) {
	c := gomock.NewController(t)
	r, a, e, i, k := NewMockRepository(c), NewMockAuthorizer(c), NewMockPlaybackEligibility(c), NewMockMediaIssuer(c), NewMockKeyer(c)
	return r, a, e, i, k, NewService(r, a, e, i, k, ids{}, func() time.Time { return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC) })
}
func pod(t *testing.T) domain.Pod {
	p, e := domain.Create("pod-1", key(1), key(2), []string{key(3)}, time.Date(2026, 7, 26, 13, 0, 0, 0, time.UTC), domain.Command{ID: "create-1", ActorKey: key(1), ReasonCode: "user_requested", At: time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)})
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func TestPlaybackRevalidatesBeforeReplayAndUsesOpaqueMedia(t *testing.T) {
	r, a, e, i, k, s := fixture(t)
	p := pod(t)
	played, _ := p.Play(domain.Command{ID: "play-1", ActorKey: key(3), ReasonCode: "user_requested", ExpectedRevision: 1, At: time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)})
	r.EXPECT().Find(gomock.Any(), "pod-1").Return(played, nil)
	a.EXPECT().Require(gomock.Any(), "member-3", "seed.pod.playback", "pod-1")
	e.EXPECT().Revalidate(gomock.Any(), "member-3", "pod-1")
	k.EXPECT().Key("seed-pod:member", "member-3").Return(key(3), nil)
	i.EXPECT().Issue(gomock.Any(), key(2), "play-1", 5*time.Minute).Return("opaque-token", nil)
	x, err := s.Playback(context.Background(), Command{ID: "play-1", PodID: "pod-1", ActorID: "member-3", ReasonCode: "user_requested", ExpectedRevision: 1})
	if err != nil || !x.Replayed || x.PlaybackToken != "opaque-token" {
		t.Fatalf("%+v %v", x, err)
	}
}
func TestPlaybackDenialsArePrivacyNeutral(t *testing.T) {
	r, a, e, _, _, s := fixture(t)
	p := pod(t)
	r.EXPECT().Find(gomock.Any(), "pod-1").Return(p, nil)
	a.EXPECT().Require(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	e.EXPECT().Revalidate(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("revoked"))
	_, err := s.Playback(context.Background(), Command{ID: "play-1", PodID: "pod-1", ActorID: "member-3", ReasonCode: "user_requested", ExpectedRevision: 1})
	if !errors.Is(err, ErrNotAvailable) {
		t.Fatal(err)
	}
}
