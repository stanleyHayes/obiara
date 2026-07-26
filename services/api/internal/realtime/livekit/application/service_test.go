package application

import (
	"context"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestServiceDelegatesOpaqueRequest(t *testing.T) {
	c := gomock.NewController(t)
	i := NewMockTokenIssuer(c)
	r := JoinRequest{RoomKey: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ParticipantKey: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", Role: RoleListener, ServerTime: time.Now(), TTL: time.Minute}
	i.EXPECT().Issue(gomock.Any(), r).Return(JoinToken{Signed: "jwt"}, nil)
	got, e := NewService(i).Issue(context.Background(), r)
	if e != nil || got.Signed != "jwt" {
		t.Fatal(e)
	}
}
