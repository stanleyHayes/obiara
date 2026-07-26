package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/consent/domain"
	"go.uber.org/mock/gomock"
)

var serviceTime = time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)

func applicationPurpose(t *testing.T, version uint64) domain.Purpose {
	t.Helper()
	purpose, err := domain.NewPurpose(domain.NewPurposeParams{
		ID: "terms.service", Kind: domain.PurposeTerms, Version: version,
		ContentRef: "content.terms.v1", Status: domain.PurposeActive,
		EffectiveSince: serviceTime.Add(-time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	return purpose
}

func applicationEvidence(t *testing.T, version uint64) domain.Evidence {
	t.Helper()
	evidence, err := domain.NewEvidence(domain.EvidenceAcknowledgement, version, "evidence.terms.v1")
	if err != nil {
		t.Fatal(err)
	}
	return evidence
}

func applicationCommand(t *testing.T) Command {
	t.Helper()
	return Command{
		CommandID: "command:grant", SubjectID: "member:123", PurposeID: "terms.service",
		PurposeVersion: 1, ExpectedRevision: 0, ActorID: "member:123",
		ActorKind: domain.ActorSubject, Source: domain.SourceMobile,
		Evidence: applicationEvidence(t, 1),
	}
}

func TestServiceGrantPersistsWithOptimisticRevision(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)
	command := applicationCommand(t)

	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), Key{SubjectID: "member:123", PurposeID: "terms.service"}).
		Return(domain.Record{}, ErrNotFound)
	records.EXPECT().Save(gomock.Any(), gomock.Cond(func(record domain.Record) bool {
		history := record.History()
		return record.Revision() == 1 && len(history) == 1 &&
			history[0].RecordedAt().Equal(serviceTime) &&
			history[0].CommandID() == command.CommandID
	}), uint64(0), command.CommandID).Return(nil)

	result, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Grant(context.Background(), command)
	if err != nil || result.Replayed || result.Record.Revision() != 1 {
		t.Fatalf("unexpected result: %+v err=%v", result, err)
	}
}

func TestServiceReplayReturnsExistingRecordWithoutSaving(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)
	record, _ := domain.NewRecord("member:123", purpose.ID())
	record, _ = record.Grant(domain.Change{
		CommandID: "command:grant", ExpectedRevision: 0, Purpose: purpose,
		ActorID: "member:123", ActorKind: domain.ActorSubject, Source: domain.SourceMobile,
		Evidence: applicationEvidence(t, 1), RecordedAt: serviceTime,
	})

	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), gomock.Any()).Return(record, nil)

	result, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Grant(context.Background(), applicationCommand(t))
	if err != nil || !result.Replayed || result.Record.Revision() != 1 {
		t.Fatalf("replay was not recognized: %+v err=%v", result, err)
	}
}

func TestServiceRejectsReusedCommandWithDifferentMeaning(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)
	record, _ := domain.NewRecord("member:123", purpose.ID())
	record, _ = record.Grant(domain.Change{
		CommandID: "command:grant", ExpectedRevision: 0, Purpose: purpose,
		ActorID: "member:123", ActorKind: domain.ActorSubject, Source: domain.SourceMobile,
		Evidence: applicationEvidence(t, 1), RecordedAt: serviceTime,
	})
	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), gomock.Any()).Return(record, nil)

	command := applicationCommand(t)
	command.Source = domain.SourceWeb
	_, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Grant(context.Background(), command)
	if !errors.Is(err, domain.ErrCommandMismatch) {
		t.Fatalf("expected command mismatch, got %v", err)
	}
}

func TestServiceRecoversConcurrentCommandReplay(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)
	empty, _ := domain.NewRecord("member:123", purpose.ID())
	applied, _ := empty.Grant(domain.Change{
		CommandID: "command:grant", ExpectedRevision: 0, Purpose: purpose,
		ActorID: "member:123", ActorKind: domain.ActorSubject, Source: domain.SourceMobile,
		Evidence: applicationEvidence(t, 1), RecordedAt: serviceTime,
	})

	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), gomock.Any()).Return(empty, nil)
	records.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "command:grant").
		Return(ErrCommandAlreadyApplied)
	records.EXPECT().Find(gomock.Any(), gomock.Any()).Return(applied, nil)

	result, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Grant(context.Background(), applicationCommand(t))
	if err != nil || !result.Replayed || result.Record.Revision() != 1 {
		t.Fatalf("concurrent replay was not recovered: %+v err=%v", result, err)
	}
}

func TestServiceMapsOptimisticConflictToStaleRevision(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)
	empty, _ := domain.NewRecord("member:123", purpose.ID())

	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), gomock.Any()).Return(empty, nil)
	records.EXPECT().Save(gomock.Any(), gomock.Any(), uint64(0), "command:grant").
		Return(ErrOptimisticConflict)

	_, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Grant(context.Background(), applicationCommand(t))
	if !errors.Is(err, domain.ErrStaleRevision) {
		t.Fatalf("expected stale revision, got %v", err)
	}
}

func TestServiceEffectiveDeniesMissingRecord(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)

	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), Key{SubjectID: "member:404", PurposeID: "terms.service"}).
		Return(domain.Record{}, ErrNotFound)

	effective, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Effective(context.Background(), "member:404", "terms.service", 1)
	if err != nil || effective {
		t.Fatalf("missing record must deny: effective=%v err=%v", effective, err)
	}
}

func TestServiceWithdrawRetainsPriorGrant(t *testing.T) {
	controller := gomock.NewController(t)
	records := NewMockRepository(controller)
	purposes := NewMockPurposeCatalog(controller)
	purpose := applicationPurpose(t, 1)
	record, _ := domain.NewRecord("member:123", purpose.ID())
	record, _ = record.Grant(domain.Change{
		CommandID: "command:grant", ExpectedRevision: 0, Purpose: purpose,
		ActorID: "member:123", ActorKind: domain.ActorSubject, Source: domain.SourceMobile,
		Evidence: applicationEvidence(t, 1), RecordedAt: serviceTime.Add(-time.Minute),
	})
	command := applicationCommand(t)
	command.CommandID = "command:withdraw"
	command.ExpectedRevision = 1

	purposes.EXPECT().FindVersion(gomock.Any(), "terms.service", uint64(1)).Return(purpose, nil)
	records.EXPECT().Find(gomock.Any(), gomock.Any()).Return(record, nil)
	records.EXPECT().Save(gomock.Any(), gomock.Cond(func(next domain.Record) bool {
		history := next.History()
		return len(history) == 2 &&
			history[0].Action() == domain.ActionGranted &&
			history[1].Action() == domain.ActionWithdrawn
	}), uint64(1), "command:withdraw").Return(nil)

	result, err := NewService(records, purposes, func() time.Time { return serviceTime }).
		Withdraw(context.Background(), command)
	if err != nil || result.Record.Effective(purpose, serviceTime) {
		t.Fatalf("withdrawal should revoke consent: %+v err=%v", result, err)
	}
}
