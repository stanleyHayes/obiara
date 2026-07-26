package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/drum/domain"
)

type ids struct{ value string }

func (source ids) NewID() string { return source.value }

func TestOpenRequiresVoiceAndPersistsOpaqueStage(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	membership := NewMockMembership(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, authorizer, membership, keyer, ids{"stage-1"}, func() time.Time { return now })
	command := Command{ID: "open-1", ActorID: "member-a", ReasonCode: "member_action"}
	pair := Pair{FirstMemberID: "member-a", SecondMemberID: "member-b"}
	if _, err := service.Open(context.Background(), command, pair, ""); err != ErrNotAvailable {
		t.Fatalf("text-only open error=%v", err)
	}
	authorizer.EXPECT().Require(gomock.Any(), "member-a", "courtship.drum.open", "").Return(nil)
	membership.EXPECT().RevalidatePair(gomock.Any(), "member-a", "member-b").Return(nil)
	keyer.EXPECT().Key("courtship-drum:member", "member-a").Return(key(1), nil)
	keyer.EXPECT().Key("courtship-drum:member", "member-b").Return(key(2), nil)
	keyer.EXPECT().Key("courtship-drum:voice", "voice-private").Return(key(9), nil)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, stage domain.Stage) error {
		if stage.Beats()[0].Medium != domain.MediumVoice || stage.Beats()[0].ContentKey != key(9) {
			t.Fatalf("stage=%#v", stage)
		}
		return nil
	})
	if _, err := service.Open(context.Background(), command, pair, "voice-private"); err != nil {
		t.Fatal(err)
	}
}

func TestServiceCannotOverrideServerTurnAuthority(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	membership := NewMockMembership(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	stage, _ := domain.Open("stage-1", []string{key(1), key(2)}, key(9), domain.Command{
		ID: "open", ActorKey: key(1), ReasonCode: "member_action", At: now,
	})
	service := NewService(repository, authorizer, membership, keyer, ids{"unused"}, func() time.Time { return now.Add(time.Second) })
	repository.EXPECT().Find(gomock.Any(), "stage-1").Return(stage, nil)
	authorizer.EXPECT().Require(gomock.Any(), "member-a", "courtship.drum.turn", "stage-1").Return(nil)
	membership.EXPECT().RequireParticipant(gomock.Any(), "member-a", "stage-1").Return(nil)
	keyer.EXPECT().Key("courtship-drum:member", "member-a").Return(key(1), nil)
	keyer.EXPECT().Key("courtship-drum:text", "text-private").Return(key(10), nil)
	_, err := service.AddText(context.Background(), Command{
		ID: "double", StageID: "stage-1", ActorID: "member-a", ReasonCode: "member_action", ExpectedRevision: 1,
	}, "text-private")
	if err != domain.ErrNotTurn {
		t.Fatalf("double-turn error=%v", err)
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
