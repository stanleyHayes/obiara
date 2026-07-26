package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/themeprogression/domain"
)

type ids struct{ value string }

func (source ids) NewID() string { return source.value }

func TestOpenRequiresDurableThemeOneRevealEvidence(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	membership := NewMockMembership(ctrl)
	evidence := NewMockThemeOneEvidence(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, authorizer, membership, evidence, keyer, ids{"progression-1"}, func() time.Time { return now })
	command := Command{ID: "open", ActorID: "member-a", ReasonCode: "member_action"}
	pair := Pair{FirstMemberID: "member-a", SecondMemberID: "member-b"}
	authorizer.EXPECT().Require(gomock.Any(), "member-a", "courtship.themeprogression.open", "").Return(nil)
	membership.EXPECT().RevalidatePair(gomock.Any(), "member-a", "member-b").Return(nil)
	evidence.EXPECT().BothRevealed(gomock.Any(), "member-a", "member-b").Return("", false, nil)
	if _, err := service.Open(context.Background(), command, pair); err != ErrNotAvailable {
		t.Fatalf("missing evidence error=%v", err)
	}
}

func TestSubmitPersistsOpaqueResponseWhileProjectionStaysConcealed(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	membership := NewMockMembership(ctrl)
	evidence := NewMockThemeOneEvidence(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	progression, _ := domain.Open("progression-1", []string{key(1), key(2)}, key(9), domain.Command{
		ID: "open", ActorKey: key(1), ReasonCode: "member_action", At: now,
	})
	service := NewService(repository, authorizer, membership, evidence, keyer, ids{"unused"}, func() time.Time { return now.Add(time.Second) })
	repository.EXPECT().Find(gomock.Any(), "progression-1").Return(progression, nil)
	authorizer.EXPECT().Require(gomock.Any(), "member-a", "courtship.themeprogression.submit", "progression-1").Return(nil)
	membership.EXPECT().RequireParticipant(gomock.Any(), "member-a", "progression-1").Return(nil)
	keyer.EXPECT().Key("courtship-themeprogression:member", "member-a").Return(key(1), nil)
	keyer.EXPECT().Key("courtship-themeprogression:encrypted-content", "cipher-private").Return(key(20), nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "submit").DoAndReturn(
		func(_ context.Context, next domain.Progression, _ uint64, _ string) error {
			current := next.Projection().Themes[0]
			if current.Revealed || len(current.Submissions) != 0 || next.Events()[1].ContentKey != key(20) {
				t.Fatalf("unsafe progression=%#v", next)
			}
			return nil
		},
	)
	_, err := service.Submit(context.Background(), Command{
		ID: "submit", ProgressionID: "progression-1", ActorID: "member-a",
		ReasonCode: "member_action", ExpectedRevision: 1,
	}, domain.ThemeTwo, "cipher-private")
	if err != nil {
		t.Fatal(err)
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
