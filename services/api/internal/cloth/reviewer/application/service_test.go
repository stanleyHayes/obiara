package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/cloth/reviewer/domain"
	"go.uber.org/mock/gomock"
)

type fixedToken string

func (token fixedToken) Token() (string, error) { return string(token), nil }

type fixedID string

func (id fixedID) NewID() string { return string(id) }
func appKey(number int) string   { return fmt.Sprintf("%064x", number) }

func TestCreateStoresDigestsAndReturnsInviteOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository, authorizer, policy := NewMockRepository(ctrl), NewMockAuthorizer(ctrl), NewMockPairPolicy(ctrl)
	keyer, watermarker := NewMockKeyer(ctrl), NewMockWatermarker(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, authorizer, policy, keyer, fixedToken("raw-invite"), fixedToken("raw-session"), watermarker, fixedID("review-1"), func() time.Time { return now })
	authorizer.EXPECT().Require(gomock.Any(), "member-a", "cloth.reviewer.create", "")
	policy.EXPECT().Revalidate(gomock.Any(), "member-a", "member-b")
	keyer.EXPECT().Key("cloth-reviewer:member", "member-a").Return(appKey(1), nil)
	keyer.EXPECT().Key("cloth-reviewer:member", "member-b").Return(appKey(2), nil)
	keyer.EXPECT().Key("cloth-reviewer:reviewer", "reviewer").Return(appKey(3), nil)
	keyer.EXPECT().Key("cloth-reviewer:invite", "raw-invite").Return(appKey(4), nil)
	keyer.EXPECT().Key("cloth-reviewer:otp", "123456").Return(appKey(5), nil)
	watermarker.EXPECT().Ref(appKey(3), "review-1").Return(appKey(6), nil)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, review domain.Review) error {
		if review.InviteDigest() != appKey(4) || review.OTPDigest() != appKey(5) ||
			review.SessionDigest() != "" || review.WatermarkRef() != appKey(6) {
			t.Fatalf("unsafe persisted review: %+v", review)
		}
		return nil
	})
	result, err := service.Create(context.Background(), Command{ID: "create", ActorID: "member-a", FirstMemberID: "member-a", SecondMemberID: "member-b"}, CreateRequest{
		ReviewerID: "reviewer", OTP: "123456", QuestionRefs: []string{"question-1"},
		OTPValidity: 10 * time.Minute, InviteValidity: 24 * time.Hour,
	})
	if err != nil || result.InviteToken != "raw-invite" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
