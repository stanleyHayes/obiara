package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

func TestUpsertCreatesAndReplaysProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	now := time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC)
	service := NewService(repository, nil, func() time.Time { return now })
	command := validCommand()

	repository.EXPECT().Find(gomock.Any(), "member-1").Return(domain.Profile{}, ErrNotFound)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "cmd-1").Return(nil)
	created, err := service.Upsert(context.Background(), command)
	if err != nil || created.Replayed || created.Profile.Revision() != 1 {
		t.Fatalf("created = %#v, error = %v", created, err)
	}

	repository.EXPECT().Find(gomock.Any(), "member-1").Return(created.Profile, nil)
	replayed, err := service.Upsert(context.Background(), command)
	if err != nil || !replayed.Replayed || replayed.Profile.Revision() != 1 {
		t.Fatalf("replayed = %#v, error = %v", replayed, err)
	}
}

func TestUpsertMapsOptimisticConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	service := NewService(repository, nil, time.Now)
	command := validCommand()
	current := makeProfile(t, command)
	command.CommandID = "cmd-2"
	command.ExpectedRevision = 1

	repository.EXPECT().Find(gomock.Any(), "member-1").Return(current, nil)
	repository.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(1), "cmd-2").Return(ErrOptimisticConflict)
	_, err := service.Upsert(context.Background(), command)
	if !errors.Is(err, domain.ErrStaleRevision) {
		t.Fatalf("error = %v, want %v", err, domain.ErrStaleRevision)
	}
}

func TestViewRequiresConsentForEachCommunityField(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	consent := NewMockConsentEvaluator(ctrl)
	command := validCommand()
	command.DisplayName.Visibility = domain.VisibilityCommunity
	command.DisplayName.ConsentRef = "consent-display"
	command.Introduction.Visibility = domain.VisibilityCommunity
	command.Introduction.ConsentRef = "consent-introduction"
	profile := makeProfile(t, command)
	service := NewService(repository, consent, time.Now)

	repository.EXPECT().Find(gomock.Any(), "member-1").Return(profile, nil)
	consent.EXPECT().Allows(gomock.Any(), "member-1", "consent-display").Return(true, nil)
	consent.EXPECT().Allows(gomock.Any(), "member-1", "consent-introduction").Return(false, nil)
	if _, err := service.View(context.Background(), "member-1", domain.AudienceCommunity); !errors.Is(err, ErrConsentDenied) {
		t.Fatalf("error = %v, want %v", err, ErrConsentDenied)
	}

	repository.EXPECT().Find(gomock.Any(), "member-1").Return(profile, nil)
	consent.EXPECT().Allows(gomock.Any(), "member-1", "consent-display").Return(true, nil)
	consent.EXPECT().Allows(gomock.Any(), "member-1", "consent-introduction").Return(true, nil)
	view, err := service.View(context.Background(), "member-1", domain.AudienceCommunity)
	if err != nil || view.DisplayName == nil || view.Introduction == nil {
		t.Fatalf("view = %#v, error = %v", view, err)
	}
}

func validCommand() UpsertCommand {
	return UpsertCommand{
		CommandID: "cmd-1", MemberID: "member-1",
		DisplayName:  FieldInput{Value: "Ama", Visibility: domain.VisibilityCircles},
		Introduction: FieldInput{Value: "Here to build community.", Visibility: domain.VisibilityPrivate},
	}
}

func makeProfile(t *testing.T, command UpsertCommand) domain.Profile {
	t.Helper()
	display, err := domain.NewField(command.DisplayName.Value, command.DisplayName.Visibility, command.DisplayName.ConsentRef, 80, true)
	if err != nil {
		t.Fatal(err)
	}
	introduction, err := domain.NewField(command.Introduction.Value, command.Introduction.Visibility, command.Introduction.ConsentRef, 280, false)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := domain.Create(command.MemberID, domain.Change{
		CommandID: command.CommandID, DisplayName: display, Introduction: introduction,
		RecordedAt: time.Date(2026, 7, 26, 16, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	return profile
}
