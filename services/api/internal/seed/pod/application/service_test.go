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
	i.EXPECT().Issue(gomock.Any(), "member-3", key(2), "play-1", 5*time.Minute).Return("opaque-token", nil)
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

func TestTheHouseFrontShowsWhatIsRestingForYou(t *testing.T) {
	// Without this a member could only open a pod whose id they already
	// knew, and nothing told them — a house front with no door.
	ctrl := gomock.NewController(t)
	r := NewMockRepository(ctrl)
	a := NewMockAuthorizer(ctrl)
	e := NewMockPlaybackEligibility(ctrl)
	i := NewMockMediaIssuer(ctrl)
	k := NewMockKeyer(ctrl)

	a.EXPECT().Require(gomock.Any(), "member-3", "seed.pod.playback", "").Return(nil)
	k.EXPECT().Key("seed-pod:member", "member-3").Return(key(3), nil)
	// The member's own key is what the query runs on: the house front asks
	// "what is resting for this person", never "what did this person send".
	r.EXPECT().ForRecipient(gomock.Any(), key(3), gomock.Any(), 10).Return(nil, nil)

	service := NewService(r, a, e, i, k, NewMockIDSource(ctrl), func() time.Time { return time.Now().UTC() })
	if _, err := service.Resting(context.Background(), "member-3", 10); err != nil {
		t.Fatal(err)
	}
}

func TestAnUnverifiedMemberHasNoHouseFront(t *testing.T) {
	ctrl := gomock.NewController(t)
	r := NewMockRepository(ctrl)
	a := NewMockAuthorizer(ctrl)
	a.EXPECT().Require(gomock.Any(), "member-3", "seed.pod.playback", "").Return(ErrNotAvailable)
	// No repository expectation: a refused member's pods are never read.

	service := NewService(r, a, NewMockPlaybackEligibility(ctrl), NewMockMediaIssuer(ctrl),
		NewMockKeyer(ctrl), NewMockIDSource(ctrl), func() time.Time { return time.Now().UTC() })
	if _, err := service.Resting(context.Background(), "member-3", 10); !errors.Is(err, ErrNotAvailable) {
		t.Fatalf("err = %v, want ErrNotAvailable", err)
	}
}
