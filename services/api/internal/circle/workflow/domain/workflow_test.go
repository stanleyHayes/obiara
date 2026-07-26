package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func command(id, actor, reason string, revision uint64, at time.Time) Command {
	return Command{ID: id, ActorID: actor, ExpectedRevision: revision, ReasonCode: reason, At: at}
}

func TestInviteExpiresAndCannotBeRedeemedOrRevealed(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	invite, err := NewInvite("invite-1", "circle-1",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		now.Add(time.Hour), command("create-1", "host-1", "host_invite", 0, now))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := invite.Redeem(command("redeem-1", "member-1", "invite_redeem", 1, now.Add(time.Hour))); !errors.Is(err, ErrInviteExpired) {
		t.Fatalf("expired redemption = %v", err)
	}
	expired, err := invite.Expire(command("expire-1", "system", "invite_expired", 1, now.Add(time.Hour)))
	if err != nil || expired.Status() != InviteExpired {
		t.Fatalf("expire = %s, %v", expired.Status(), err)
	}
}

func TestRequestTransitionsAreReplaySafeAndOneWay(t *testing.T) {
	now := time.Now().UTC()
	request, err := NewRequest("request-1", "circle-1", "member-1", "direct",
		command("request-1", "member-1", "member_request", 0, now))
	if err != nil {
		t.Fatal(err)
	}
	approvedCommand := command("approve-1", "host-1", "request_approved", 1, now.Add(time.Second))
	approved, err := request.Approve(approvedCommand)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := approved.Approve(approvedCommand)
	if err != nil || replayed.Revision() != approved.Revision() {
		t.Fatalf("replay = %d, %v", replayed.Revision(), err)
	}
	if _, err := approved.Decline(command("decline-1", "host-1", "request_declined", 2, now.Add(2*time.Second))); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("approved request declined: %v", err)
	}
	mismatch := approvedCommand
	mismatch.ReasonCode = "different_reason"
	if _, err := approved.Approve(mismatch); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("changed replay = %v", err)
	}
}

func TestPropertyRequestHistoryAlwaysRehydrates(t *testing.T) {
	property := func(actions []byte) bool {
		if len(actions) > 80 {
			actions = actions[:80]
		}
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		request, err := NewRequest("request-p", "circle-p", "member-p", "direct",
			command("start-p", "member-p", "member_request", 0, now))
		if err != nil {
			return false
		}
		for index, value := range actions {
			change := command(fmt.Sprintf("property-%d", index), "host-p", "property_action", request.Revision(), now.Add(time.Duration(index+1)*time.Second))
			var next Request
			switch value % 3 {
			case 0:
				next, err = request.Approve(change)
			case 1:
				next, err = request.Decline(change)
			default:
				next, err = request.Expel(change)
			}
			if err == nil {
				request = next
			}
			rehydrated, hydrateErr := RehydrateRequest(RequestState{
				ID: request.ID(), CircleID: request.CircleID(), MemberID: request.MemberID(),
				Source: request.Source(), Status: request.Status(), Revision: request.Revision(),
				Events: request.Events(), Commands: request.Commands(),
			})
			if hydrateErr != nil || rehydrated.Status() != request.Status() || rehydrated.Revision() != request.Revision() {
				return false
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatal(err)
	}
}
