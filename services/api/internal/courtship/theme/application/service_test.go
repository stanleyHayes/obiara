package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/courtship/theme/domain"
)

type ids struct{ value string }

func (source ids) NewID() string { return source.value }

func TestSubmitPersistsOpaqueContentAndReturnsConcealedProjection(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	membership := NewMockMembership(ctrl)
	keyer := NewMockKeyer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	theme, _ := domain.Open("theme-1", []string{key(1), key(2)}, domain.Command{
		ID: "open", ActorKey: key(1), ReasonCode: "member_action", At: now,
	})
	service := NewService(repository, authorizer, membership, keyer, ids{"unused"}, func() time.Time { return now.Add(time.Second) })
	repository.EXPECT().Find(gomock.Any(), "theme-1").Return(theme, nil)
	authorizer.EXPECT().Require(gomock.Any(), "member-a", "courtship.theme.submit", "theme-1").Return(nil)
	membership.EXPECT().RequireParticipant(gomock.Any(), "member-a", "theme-1").Return(nil)
	keyer.EXPECT().Key("courtship-theme:member", "member-a").Return(key(1), nil)
	keyer.EXPECT().Key("courtship-theme:encrypted-content", "cipher-private").Return(key(10), nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "submit-1").DoAndReturn(
		func(_ context.Context, next domain.Theme, _ uint64, _ string) error {
			if next.Projection().Revealed || len(next.Projection().Submissions) != 0 ||
				next.Events()[1].ContentKey != key(10) {
				t.Fatalf("unsafe theme=%#v", next)
			}
			return nil
		},
	)
	result, err := service.Submit(context.Background(), Command{
		ID: "submit-1", ThemeID: "theme-1", ActorID: "member-a",
		ReasonCode: "member_action", ExpectedRevision: 1,
	}, "cipher-private")
	if err != nil || result.Theme.Projection().Revealed {
		t.Fatalf("result=%#v error=%v", result, err)
	}
}

func key(number int) string { return fmt.Sprintf("%064x", number) }
