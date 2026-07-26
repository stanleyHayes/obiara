package application

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/counsel/isolation/domain"
	"go.uber.org/mock/gomock"
)

func TestExplicitSafetyEscalationRevalidatesAndPublishesOnlyMinimalEvent(t *testing.T) {
	controller := gomock.NewController(t)
	scope := NewMockScope(controller)
	consent := NewMockConsent(controller)
	authority := NewMockAuthority(controller)
	safety := NewMockSafetySink(controller)
	ids := NewMockIDSource(controller)
	session, actor, subject := isolationKey("a"), isolationKey("b"), isolationKey("c")
	eventID := isolationKey("d")
	now := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	request := Request{
		SessionKey:           session,
		ActorKey:             actor,
		SubjectKey:           subject,
		ConsentVersion:       9,
		ExplicitConfirmation: true,
	}
	want := domain.SafetyEvent{
		ID:             eventID,
		SubjectKey:     subject,
		Reason:         domain.ReasonExplicitSafetySupport,
		OccurredAt:     now,
		ConsentVersion: 9,
	}
	gomock.InOrder(
		scope.EXPECT().ContainsBoth(gomock.Any(), session, actor, subject).Return(true, nil),
		consent.EXPECT().CurrentAllows(gomock.Any(), subject, SafetyEscalationPurpose, uint64(9)).Return(true, nil),
		authority.EXPECT().AuthorizeEscalation(gomock.Any(), actor, subject).Return(nil),
		ids.EXPECT().NewID().Return(eventID),
		safety.EXPECT().Publish(gomock.Any(), want).Return(nil),
	)
	got, err := New(scope, consent, authority, safety, ids, func() time.Time { return now }).
		Escalate(context.Background(), request)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("event=%+v err=%v", got, err)
	}
}

func TestEscalationRequiresExplicitConfirmation(t *testing.T) {
	controller := gomock.NewController(t)
	service := New(
		NewMockScope(controller),
		NewMockConsent(controller),
		NewMockAuthority(controller),
		NewMockSafetySink(controller),
		NewMockIDSource(controller),
		time.Now,
	)
	if _, err := service.Escalate(context.Background(), Request{
		SessionKey:     isolationKey("a"),
		ActorKey:       isolationKey("b"),
		SubjectKey:     isolationKey("c"),
		ConsentVersion: 1,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("implicit escalation = %v", err)
	}
}

func TestWithdrawalOrAuthorityFailureNeverPublishes(t *testing.T) {
	for _, test := range []struct {
		name          string
		consentAllows bool
		consentErr    error
		authorityErr  error
	}{
		{name: "withdrawn"},
		{name: "consent unavailable", consentErr: errors.New("unavailable")},
		{name: "authority withdrawn", consentAllows: true, authorityErr: errors.New("denied")},
	} {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			scope := NewMockScope(controller)
			consent := NewMockConsent(controller)
			authority := NewMockAuthority(controller)
			safety := NewMockSafetySink(controller)
			ids := NewMockIDSource(controller)
			session, actor, subject := isolationKey("a"), isolationKey("b"), isolationKey("c")
			scope.EXPECT().ContainsBoth(gomock.Any(), session, actor, subject).Return(true, nil)
			consent.EXPECT().CurrentAllows(gomock.Any(), subject, SafetyEscalationPurpose, uint64(1)).
				Return(test.consentAllows, test.consentErr)
			if test.consentAllows && test.consentErr == nil {
				authority.EXPECT().AuthorizeEscalation(gomock.Any(), actor, subject).Return(test.authorityErr)
			}
			_, err := New(scope, consent, authority, safety, ids, time.Now).Escalate(
				context.Background(),
				Request{
					SessionKey:           session,
					ActorKey:             actor,
					SubjectKey:           subject,
					ConsentVersion:       1,
					ExplicitConfirmation: true,
				},
			)
			if !errors.Is(err, ErrUnavailable) {
				t.Fatalf("failure = %v", err)
			}
		})
	}
}

func isolationKey(character string) string {
	return strings.Repeat(character, 64)
}
