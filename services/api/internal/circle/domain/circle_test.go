package domain

import (
	"errors"
	"fmt"
	"testing"
	"testing/quick"
	"time"
)

func TestCircleIsPrivateAndDenyByDefault(t *testing.T) {
	circle := newCircle(t)
	if circle.Visibility() != VisibilityPrivate {
		t.Fatalf("visibility = %q, want %q", circle.Visibility(), VisibilityPrivate)
	}
	for _, capability := range []Capability{CapabilityDiscover, CapabilityView, CapabilityPost, CapabilityManage, Capability("unknown")} {
		if circle.Allows("stranger", capability) {
			t.Fatalf("stranger unexpectedly allowed %q", capability)
		}
	}
	if !circle.Allows("owner-1", CapabilityManage) {
		t.Fatal("owner cannot manage circle")
	}
}

func TestMembershipTransitionsAreOneWayAndAudited(t *testing.T) {
	circle := newCircle(t)
	circle = apply(t, circle, "member-1", StateRequested, "member-1")
	if circle.Allows("member-1", CapabilityView) {
		t.Fatal("requested member can view private circle")
	}
	circle = apply(t, circle, "member-1", StateMember, "owner-1")
	circle = apply(t, circle, "member-1", StateHost, "owner-1")
	circle = apply(t, circle, "member-1", StateExpelled, "owner-1")
	if circle.Allows("member-1", CapabilityDiscover) {
		t.Fatal("expelled member retained access")
	}
	if len(circle.History()) != int(circle.Revision()) {
		t.Fatalf("history length = %d, revision = %d", len(circle.History()), circle.Revision())
	}
	rejoin := testCommand(circle, "rejoin", "member-1", "membership.request", "member-1")
	if _, err := circle.Request("member-1", rejoin); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("rejoin error = %v, want %v", err, ErrInvalidTransition)
	}
}

func TestReplayAndOptimisticRevision(t *testing.T) {
	circle := newCircle(t)
	command := testCommand(circle, "request-1", "member-1", "membership.request", "member-1")
	next, err := circle.Request("member-1", command)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := next.Request("member-1", command)
	if err != nil || replay.Revision() != next.Revision() {
		t.Fatalf("replay revision = %d, error = %v", replay.Revision(), err)
	}
	command.Payload = "member-2"
	if _, err := next.Request("member-1", command); !errors.Is(err, ErrCommandMismatch) {
		t.Fatalf("mismatch error = %v, want %v", err, ErrCommandMismatch)
	}
	stale := testCommand(circle, "request-2", "member-2", "membership.request", "member-2")
	if _, err := next.Request("member-2", stale); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("stale error = %v, want %v", err, ErrStaleRevision)
	}
}

func TestRandomTransitionsNeverReactivateTerminalMemberships(t *testing.T) {
	property := func(steps []byte) bool {
		circle := newCircle(t)
		terminal := false
		for index, step := range steps {
			memberID := "member-property"
			var target MembershipState
			switch step % 5 {
			case 0:
				target = StateRequested
			case 1:
				target = StateMember
			case 2:
				target = StateHost
			case 3:
				target = StateLeft
			default:
				target = StateExpelled
			}
			actor := "owner-1"
			if target == StateRequested || target == StateLeft {
				actor = memberID
			}
			before := circle.Revision()
			next, err := transitionForTest(circle, memberID, target, testCommand(
				circle, fmt.Sprintf("property-%d", index), actor, "property", string(target),
			))
			if err == nil {
				circle = next
				if target == StateLeft || target == StateExpelled {
					terminal = true
				}
				if terminal && membershipState(circle, memberID) != target {
					return false
				}
				if circle.Revision() != before+1 || len(circle.History()) != int(circle.Revision()) {
					return false
				}
			} else if terminal {
				state := membershipState(circle, memberID)
				if state != StateLeft && state != StateExpelled {
					return false
				}
			}
		}
		return true
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatal(err)
	}
}

func apply(t *testing.T, circle Circle, memberID string, target MembershipState, actor string) Circle {
	t.Helper()
	next, err := transitionForTest(circle, memberID, target, testCommand(
		circle, fmt.Sprintf("%s-%d", target, circle.Revision()), actor, "membership."+string(target), memberID,
	))
	if err != nil {
		t.Fatal(err)
	}
	return next
}

func transitionForTest(circle Circle, memberID string, target MembershipState, command Command) (Circle, error) {
	switch target {
	case StateRequested:
		return circle.Request(memberID, command)
	case StateMember:
		return circle.Approve(memberID, command)
	case StateHost:
		return circle.PromoteHost(memberID, command)
	case StateLeft:
		return circle.Leave(memberID, command)
	case StateExpelled:
		return circle.Expel(memberID, command)
	default:
		return Circle{}, ErrInvalidTransition
	}
}

func membershipState(circle Circle, memberID string) MembershipState {
	for _, membership := range circle.Memberships() {
		if membership.MemberID() == memberID {
			return membership.State()
		}
	}
	return ""
}

func newCircle(t *testing.T) Circle {
	t.Helper()
	circle, err := Create("circle-1", TypeCommunity, "owner-1", Command{
		ID: "create-1", ActorID: "owner-1", Kind: "circle.create",
		Payload: string(TypeCommunity), RecordedAt: baseTime(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return circle
}

func testCommand(circle Circle, id, actor, kind, payload string) Command {
	return Command{
		ID: id, ActorID: actor, Kind: kind, Payload: payload,
		ExpectedRevision: circle.Revision(),
		RecordedAt:       baseTime().Add(time.Duration(circle.Revision()) * time.Minute),
	}
}

func baseTime() time.Time {
	return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
}
