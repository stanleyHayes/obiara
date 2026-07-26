package application

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/member/domain"
)

func TestRegisterMemberUsesRepositoryPort(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	repository := NewMockMemberRepository(controller)
	now := time.Date(2026, time.July, 26, 12, 0, 0, 0, time.UTC)
	expected, err := domain.NewMember("member-1", "member@example.com", now)
	if err != nil {
		t.Fatal(err)
	}

	repository.EXPECT().
		Create(gomock.Any(), expected).
		Return(nil)

	handler := NewRegisterMember(repository, func() time.Time { return now })
	actual, err := handler.Handle(context.Background(), RegisterMemberCommand{
		ID:             "member-1",
		Email:          "MEMBER@example.com",
		IdempotencyKey: "request-1",
	})
	if err != nil {
		t.Fatalf("register member: %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("member mismatch: got %#v want %#v", actual, expected)
	}
}

func TestRegisterMemberRequiresIdempotencyKey(t *testing.T) {
	t.Parallel()

	controller := gomock.NewController(t)
	repository := NewMockMemberRepository(controller)
	handler := NewRegisterMember(repository, time.Now)

	_, err := handler.Handle(context.Background(), RegisterMemberCommand{
		ID:    "member-1",
		Email: "member@example.com",
	})
	if err != ErrIdempotencyKeyRequired {
		t.Fatalf("got error %v, want %v", err, ErrIdempotencyKeyRequired)
	}
}
