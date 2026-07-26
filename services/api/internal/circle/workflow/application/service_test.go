package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/circle/workflow/domain"
)

func TestCreateInviteAuthorizesAndReturnsSecretOnlyOnce(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	tokens := NewMockTokenIssuer(ctrl)
	ids := NewMockIDSource(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, authorizer, tokens, ids, func() time.Time { return now })
	command := Command{ID: "command-1", CircleID: "circle-1", ActorID: "host-1", ReasonCode: "host_invite"}
	authorizer.EXPECT().Require(gomock.Any(), "host-1", "circle-1", "invite.create", "").Return(nil)
	tokens.EXPECT().NewToken().Return("opaque-token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil)
	ids.EXPECT().NewID().Return("invite-1")
	repository.EXPECT().CreateInvite(gomock.Any(), gomock.Any()).Return(nil)
	result, err := service.CreateInvite(context.Background(), command, time.Hour)
	if err != nil || result.Token != "opaque-token" || result.Invite.Status() != domain.InviteActive {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestUnauthorizedApprovalDoesNotReadOrWriteWorkflow(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	service := NewService(repository, authorizer, NewMockTokenIssuer(ctrl), NewMockIDSource(ctrl), time.Now)
	request, err := domain.NewRequest("request-1", "circle-1", "member-1", "direct", domain.Command{
		ID: "request-1", ActorID: "member-1", ReasonCode: "member_request", At: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	repository.EXPECT().FindRequest(gomock.Any(), "request-1").Return(request, nil)
	authorizer.EXPECT().Require(gomock.Any(), "outsider-1", "circle-1", "membership.approve", "member-1").Return(errors.New("denied"))
	_, err = service.Approve(context.Background(), Command{
		ID: "approve-1", ActorID: "outsider-1", RequestID: "request-1",
		ExpectedRevision: 1, ReasonCode: "request_approved",
	})
	if !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("approval = %v", err)
	}
}

func TestApprovalPersistsOptimisticallyAndReplays(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	authorizer := NewMockAuthorizer(ctrl)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	service := NewService(repository, authorizer, NewMockTokenIssuer(ctrl), NewMockIDSource(ctrl), func() time.Time { return now })
	request, err := domain.NewRequest("request-1", "circle-1", "member-1", "direct", domain.Command{
		ID: "request-1", ActorID: "member-1", ReasonCode: "member_request", At: now.Add(-time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	command := Command{
		ID: "approve-1", ActorID: "host-1", RequestID: "request-1",
		ExpectedRevision: 1, ReasonCode: "request_approved",
	}
	repository.EXPECT().FindRequest(gomock.Any(), "request-1").Return(request, nil)
	authorizer.EXPECT().Require(gomock.Any(), "host-1", "circle-1", "membership.approve", "member-1").Return(nil)
	repository.EXPECT().SaveRequest(gomock.Any(), gomock.Any(), uint64(1), "approve-1").Return(nil)
	result, err := service.Approve(context.Background(), command)
	if err != nil || result.Request.Status() != domain.RequestApproved {
		t.Fatalf("approval=%+v err=%v", result, err)
	}
}
