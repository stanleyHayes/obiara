package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/safeguarding/domain"
	"go.uber.org/mock/gomock"
)

var testNow = time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

func testRestriction(t *testing.T) domain.Restriction {
	t.Helper()
	restriction, err := domain.NewRestriction(
		"restriction:1", "command:1", strings.Repeat("a", 64),
		strings.Repeat("b", 64), testNow,
	)
	if err != nil {
		t.Fatal(err)
	}
	return restriction
}

func expectKeys(store *MockRestrictionStore, keyer *MockKeyer) {
	keyer.EXPECT().Key("member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("case:1").Return(strings.Repeat("b", 64), nil)
	store.EXPECT().FindBySubjectKey(gomock.Any(), strings.Repeat("a", 64)).
		Return(domain.Restriction{}, ErrRestrictionNotFound)
}

func TestUnder18IsBlockedBeforeImmediatePurge(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockRestrictionStore(controller)
	purger := NewMockArtifactPurger(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	expectKeys(store, keyer)
	ids.EXPECT().NewID().Return("restriction:1")
	restriction := testRestriction(t)
	store.EXPECT().CreateBlocked(gomock.Any(), restriction, gomock.Any()).Return(restriction, false, nil)
	purger.EXPECT().Purge(gomock.Any(), "member:1", "case:1").Return(nil)
	store.EXPECT().CompletePurge(gomock.Any(), gomock.Cond(func(value domain.Restriction) bool {
		return value.PurgeStatus() == domain.PurgeCompleted && value.PurgedWithinSLA()
	}), uint64(1)).Return(nil)
	purged, _ := restriction.MarkPurged(testNow, 1)
	store.EXPECT().FindByID(gomock.Any(), restriction.ID()).Return(purged, nil)

	decision, err := NewService(store, purger, keyer, ids, func() time.Time { return testNow }).
		Assess(context.Background(), Assessment{
			CommandID: "command:1", SubjectID: "member:1", SourceRef: "case:1",
			DateOfBirth: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
		})
	if !errors.Is(err, ErrUnder18) || decision.Allowed ||
		decision.Restriction.PurgeStatus() != domain.PurgeCompleted {
		t.Fatalf("under-18 decision = %+v, %v", decision, err)
	}
}

func TestPurgeFailureNeverAllowsUnder18(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockRestrictionStore(controller)
	purger := NewMockArtifactPurger(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	expectKeys(store, keyer)
	ids.EXPECT().NewID().Return("restriction:1")
	restriction := testRestriction(t)
	store.EXPECT().CreateBlocked(gomock.Any(), restriction, gomock.Any()).Return(restriction, false, nil)
	purger.EXPECT().Purge(gomock.Any(), "member:1", "case:1").Return(errors.New("storage unavailable"))

	decision, err := NewService(store, purger, keyer, ids, func() time.Time { return testNow }).
		Assess(context.Background(), Assessment{
			CommandID: "command:1", SubjectID: "member:1", SourceRef: "case:1",
			DateOfBirth: time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
		})
	if decision.Allowed || !errors.Is(err, ErrUnder18) || !errors.Is(err, ErrPurgePending) {
		t.Fatalf("purge failure weakened hard block: %+v, %v", decision, err)
	}
}

func TestExistingBlockCannotBeBypassedWithAdultBirthDate(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockRestrictionStore(controller)
	keyer := NewMockKeyer(controller)
	restriction := testRestriction(t)
	keyer.EXPECT().Key("member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("case:other").Return(strings.Repeat("c", 64), nil)
	store.EXPECT().FindBySubjectKey(gomock.Any(), strings.Repeat("a", 64)).Return(restriction, nil)

	decision, err := NewService(store, NewMockArtifactPurger(controller), keyer, NewMockIDSource(controller), func() time.Time { return testNow }).
		Assess(context.Background(), Assessment{
			CommandID: "command:other", SubjectID: "member:1", SourceRef: "case:other",
			DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		})
	if decision.Allowed || !decision.Replayed || !errors.Is(err, ErrUnder18) {
		t.Fatalf("existing block was bypassed: %+v, %v", decision, err)
	}
}

func TestAdultAtBoundaryIsAllowed(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockRestrictionStore(controller)
	keyer := NewMockKeyer(controller)
	keyer.EXPECT().Key("member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("case:1").Return(strings.Repeat("b", 64), nil)
	store.EXPECT().FindBySubjectKey(gomock.Any(), strings.Repeat("a", 64)).
		Return(domain.Restriction{}, ErrRestrictionNotFound)

	decision, err := NewService(store, NewMockArtifactPurger(controller), keyer, NewMockIDSource(controller), func() time.Time { return testNow }).
		Assess(context.Background(), Assessment{
			SubjectID: "member:1", SourceRef: "case:1",
			DateOfBirth: time.Date(2008, 7, 26, 0, 0, 0, 0, time.UTC),
		})
	if err != nil || !decision.Allowed {
		t.Fatalf("18-year-old should be allowed: %+v, %v", decision, err)
	}
}

func TestPendingPurgeRetryUsesStoredJob(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockRestrictionStore(controller)
	purger := NewMockArtifactPurger(controller)
	restriction := testRestriction(t)
	job := PurgeJob{
		RestrictionID: restriction.ID(), SubjectID: "member:1",
		SourceRef: "case:1", PurgeDueAt: restriction.PurgeDueAt(),
	}
	store.EXPECT().FindPending(gomock.Any(), restriction.PurgeDueAt(), 10).Return([]PurgeJob{job}, nil)
	store.EXPECT().FindByID(gomock.Any(), restriction.ID()).Return(restriction, nil)
	purger.EXPECT().Purge(gomock.Any(), "member:1", "case:1").Return(nil)
	store.EXPECT().CompletePurge(gomock.Any(), gomock.Any(), uint64(1)).Return(nil)

	completed, err := NewService(store, purger, NewMockKeyer(controller), NewMockIDSource(controller), func() time.Time {
		return restriction.PurgeDueAt()
	}).PurgePending(context.Background(), restriction.PurgeDueAt(), 10)
	if err != nil || completed != 1 {
		t.Fatalf("completed=%d, err=%v", completed, err)
	}
}
