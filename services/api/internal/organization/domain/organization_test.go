package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var at = time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

func command(id string, revision uint64) Command {
	return Command{
		ID: id, ActorKey: strings.Repeat("a", 64),
		ReasonCode: "operator_request", ExpectedRevision: revision, At: at,
	}
}

func registered(t *testing.T) Organization {
	t.Helper()
	organization, err := Register(
		"org_1", "Ashesi University", "billing@ashesi.edu.gh", command("cmd_1", 0))
	if err != nil {
		t.Fatal(err)
	}
	return organization
}

func TestRegisteringRecordsWhoActedAndWhy(t *testing.T) {
	// The audit trail is the point of this aggregate: a code issued in an
	// organization's name has to be traceable to somebody deciding to issue
	// it.
	organization := registered(t)
	if organization.Status() != StatusActive || organization.Revision() != 1 {
		t.Fatalf("status = %q, revision = %d", organization.Status(), organization.Revision())
	}
	events := organization.Events()
	if len(events) != 1 || events[0].Action != ActionRegistered {
		t.Fatalf("events = %#v", events)
	}
	if events[0].ActorKey != strings.Repeat("a", 64) || events[0].ReasonCode != "operator_request" {
		t.Fatal("the trail does not say who acted or why")
	}
}

func TestAnOrganizationIsNotAMember(t *testing.T) {
	// A raw operator id in the trail would make this a directory of who did
	// what. The actor is keyed, like everywhere else.
	unkeyed := command("cmd_1", 0)
	unkeyed.ActorKey = "operator-1"
	if _, err := Register("org_1", "Ashesi", "b@ashesi.edu.gh", unkeyed); !errors.Is(err, ErrInvalidOrganization) {
		t.Fatal("an unkeyed operator was recorded")
	}
}

func TestAnOrganizationNeedsSomewhereToSendAnInvoice(t *testing.T) {
	// The billing contact is the whole commercial relationship. Registering
	// without a usable one produces an organization nobody can invoice.
	for name, email := range map[string]string{
		"nothing":        "",
		"not an address": "billing at ashesi",
		"a display name": "Billing <billing@ashesi.edu.gh>",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Register("org_1", "Ashesi", email, command("cmd_1", 0)); err == nil {
				t.Fatal("registered with no way to invoice them")
			}
		})
	}
}

func TestASuspendedOrganizationIssuesNothing(t *testing.T) {
	organization := registered(t)
	if !organization.Issuing() {
		t.Fatal("an active organization was refused")
	}
	suspended, err := organization.Suspend(command("cmd_2", 1))
	if err != nil {
		t.Fatal(err)
	}
	if suspended.Issuing() {
		t.Fatal("a suspended organization could still have codes issued for it")
	}
	// And suspending twice is a transition that does not exist, rather than a
	// silent no-op that would put a second entry in the trail.
	if _, err := suspended.Suspend(command("cmd_3", 2)); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("err = %v, want ErrInvalidTransition", err)
	}
}

func TestRestoringIsItsOwnEntryInTheTrail(t *testing.T) {
	suspended, err := registered(t).Suspend(command("cmd_2", 1))
	if err != nil {
		t.Fatal(err)
	}
	restored, err := suspended.Restore(command("cmd_3", 2))
	if err != nil || !restored.Issuing() {
		t.Fatalf("restored = %#v, err = %v", restored.Status(), err)
	}
	actions := make([]Action, 0, 3)
	for _, event := range restored.Events() {
		actions = append(actions, event.Action)
	}
	if len(actions) != 3 || actions[2] != ActionRestored {
		t.Fatalf("trail = %v", actions)
	}
}

func TestRenamingIsRecordedRatherThanSilent(t *testing.T) {
	// A code issued last year was issued by whatever the organization was
	// called then, and the trail has to say so.
	renamed, err := registered(t).Rename("Ashesi University College", command("cmd_2", 1))
	if err != nil {
		t.Fatal(err)
	}
	if renamed.Name() != "Ashesi University College" {
		t.Fatalf("name = %q", renamed.Name())
	}
	events := renamed.Events()
	if len(events) != 2 || events[1].Action != ActionRenamed {
		t.Fatalf("a rename left no trail: %#v", events)
	}
}

func TestARetriedCommandIsTheSameAnswer(t *testing.T) {
	// Retries are normal. Answering the second one with a conflict would make
	// an operator suspend an organization twice to find out it worked once.
	organization := registered(t)
	suspended, err := organization.Suspend(command("cmd_2", 1))
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := suspended.Suspend(command("cmd_2", 1))
	if err != nil {
		t.Fatalf("a retry was refused: %v", err)
	}
	if replayed.Revision() != suspended.Revision() {
		t.Fatal("a retry advanced the aggregate a second time")
	}
}

func TestOneCommandIdCannotMeanTwoThings(t *testing.T) {
	// The fingerprint binds the id to what it did. Without it a retry could
	// carry a different intent and be answered with the first one's outcome.
	suspended, err := registered(t).Suspend(command("cmd_2", 1))
	if err != nil {
		t.Fatal(err)
	}
	different := command("cmd_2", 1)
	different.ReasonCode = "something_else"
	if _, err := suspended.Suspend(different); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("err = %v, want ErrCommandMismatch", err)
	}
}

func TestAStaleRevisionIsRefused(t *testing.T) {
	organization := registered(t)
	if _, err := organization.Suspend(command("cmd_2", 7)); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("err = %v, want ErrStaleRevision", err)
	}
}

func TestRehydrationRefusesAnImpossibleOrganization(t *testing.T) {
	// A revision that does not match the trail means events were lost, and an
	// aggregate that hides that would keep appending to a broken history.
	organization := registered(t)
	if _, err := Rehydrate(State{
		ID: organization.ID(), Name: organization.Name(), BillingEmail: organization.BillingEmail(),
		Status: StatusActive, Revision: 5,
		Events: organization.Events(), Commands: organization.Commands(),
	}); !errors.Is(err, ErrInvalidOrganization) {
		t.Fatal("an organization with a broken trail was rehydrated")
	}
	if _, err := Rehydrate(State{
		ID: organization.ID(), Status: "invented", Revision: 1,
		Events: organization.Events(), Commands: organization.Commands(),
	}); !errors.Is(err, ErrInvalidOrganization) {
		t.Fatal("an organization in a status nothing produces was rehydrated")
	}
}

func TestEventsAndCommandsDoNotAliasCallerMemory(t *testing.T) {
	organization := registered(t)
	events := organization.Events()
	events[0].Action = "mutated"
	if organization.Events()[0].Action != ActionRegistered {
		t.Fatal("the trail can be edited from outside")
	}
}
