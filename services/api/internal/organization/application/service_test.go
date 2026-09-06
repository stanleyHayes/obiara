package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/organization/domain"
	"go.uber.org/mock/gomock"
)

var at = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

// keyer is the operator keyer: a different operator gives a different digest,
// which is the property the audit trail depends on.
type keyer struct{ err error }

func (k keyer) Key(namespace, value string) (string, error) {
	if k.err != nil {
		return "", k.err
	}
	return strings.Repeat("a", 63) + string(rune('0'+len(value)%10)), nil
}

type fixedID string

func (i fixedID) NewID() string { return string(i) }

func organizationFor(t *testing.T) domain.Organization {
	t.Helper()
	organization, err := domain.Register("org_1", "Ashesi", "billing@ashesi.edu.gh", domain.Command{
		ID: "cmd_1", ActorKey: strings.Repeat("a", 64), ReasonCode: "operator_request", At: at,
	})
	if err != nil {
		t.Fatal(err)
	}
	return organization
}

func TestRegisteringKeysTheOperatorBeforeAnythingIsWritten(t *testing.T) {
	// An audit trail proves somebody acted. It must not become a directory of
	// who, so the raw operator id never reaches the aggregate.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, organization domain.Organization) error {
			events := organization.Events()
			if len(events) != 1 {
				t.Fatalf("events = %#v", events)
			}
			if strings.Contains(events[0].ActorKey, "operator") {
				t.Fatalf("the raw operator reached the trail: %q", events[0].ActorKey)
			}
			return nil
		})

	service := New(repository, keyer{}, fixedID("org_1"), func() time.Time { return at })
	result, err := service.Register(context.Background(), RegisterCommand{
		CommandID: "cmd_1", OperatorID: "operator-1", ReasonCode: "operator_request",
		Name: "Ashesi", BillingEmail: "billing@ashesi.edu.gh",
	})
	if err != nil || result.Organization.Name() != "Ashesi" {
		t.Fatalf("result = %#v, err = %v", result, err)
	}
}

func TestARetriedRegistrationDoesNotCreateASecondBody(t *testing.T) {
	// The operator asked for one organization. Answering a retry with a
	// second would leave two, and a code issued afterwards would name the
	// wrong one.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(ErrCommandApplied)
	repository.EXPECT().FindByCommand(gomock.Any(), "cmd_1").Return(organizationFor(t), nil)

	service := New(repository, keyer{}, fixedID("org_2"), func() time.Time { return at })
	result, err := service.Register(context.Background(), RegisterCommand{
		CommandID: "cmd_1", OperatorID: "operator-1", ReasonCode: "operator_request",
		Name: "Ashesi", BillingEmail: "billing@ashesi.edu.gh",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Replayed || result.Organization.ID() != "org_1" {
		t.Fatalf("result = %#v", result)
	}
}

func TestSuspendingWritesOnceAndIsSafeToRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	organization := organizationFor(t)
	repository.EXPECT().FindByID(gomock.Any(), "org_1").Return(organization, nil)
	repository.EXPECT().Append(gomock.Any(), gomock.Any(), uint64(1), "cmd_2").Return(nil)

	service := New(repository, keyer{}, fixedID("org_1"), func() time.Time { return at })
	result, err := service.Suspend(context.Background(), ChangeCommand{
		CommandID: "cmd_2", OperatorID: "operator-1", ReasonCode: "unpaid_invoice",
		OrganizationID: "org_1", ExpectedRevision: 1,
	})
	if err != nil || result.Organization.Issuing() {
		t.Fatalf("result = %#v, err = %v", result, err)
	}

	// The same command again, against the already-suspended aggregate: the
	// domain recognises the replay and nothing is written. No Append
	// expectation, so a second write fails this.
	suspended := result.Organization
	repository.EXPECT().FindByID(gomock.Any(), "org_1").Return(suspended, nil)
	replayed, err := service.Suspend(context.Background(), ChangeCommand{
		CommandID: "cmd_2", OperatorID: "operator-1", ReasonCode: "unpaid_invoice",
		OrganizationID: "org_1", ExpectedRevision: 1,
	})
	if err != nil || !replayed.Replayed {
		t.Fatalf("replayed = %#v, err = %v", replayed, err)
	}
}

func TestAnUnkeyableOperatorWritesNothing(t *testing.T) {
	// If who acted cannot be established, nothing is recorded. An audit entry
	// naming nobody is worse than no entry: it says an act was accounted for
	// when it was not.
	ctrl := gomock.NewController(t)
	repository := NewMockRepository(ctrl)
	// No expectations at all: the repository must not be touched.

	service := New(repository, keyer{err: errors.New("hmac unavailable")},
		fixedID("org_1"), func() time.Time { return at })
	if _, err := service.Register(context.Background(), RegisterCommand{
		CommandID: "cmd_1", OperatorID: "operator-1", ReasonCode: "operator_request",
		Name: "Ashesi", BillingEmail: "billing@ashesi.edu.gh",
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if _, err := service.Suspend(context.Background(), ChangeCommand{
		CommandID: "cmd_2", OperatorID: "operator-1", ReasonCode: "unpaid_invoice",
		OrganizationID: "org_1", ExpectedRevision: 1,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestAnUncomposedServiceRefuses(t *testing.T) {
	// A missing dependency is not a reason to write an organization with no
	// trail behind it.
	if _, err := (Service{}).Register(context.Background(), RegisterCommand{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if _, err := (Service{}).List(context.Background(), 10); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}
