package circlepolicy

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"testing"
	"time"

	circledomain "github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/source/domain"
)

var at = time.Date(2026, time.September, 5, 12, 0, 0, 0, time.UTC)

type stubCircles struct {
	circle circledomain.Circle
	err    error
}

func (s stubCircles) Find(context.Context, string) (circledomain.Circle, error) {
	return s.circle, s.err
}

// circleWith builds a real circle through the aggregate's own transitions.
// Rehydrate replays history and demands it reconstruct the roster exactly, so
// a hand-written State is both fragile and a weaker test than driving the
// domain the way the application does.
func circleWith(t *testing.T, members map[string]circledomain.MembershipState) circledomain.Circle {
	t.Helper()
	command := func(id, actor string, revision uint64) circledomain.Command {
		return circledomain.Command{
			ID: id, ExpectedRevision: revision, ActorID: actor,
			Kind: "test", Payload: id, RecordedAt: at,
		}
	}

	circle, err := circledomain.Create("circle_1", circledomain.TypeInterest,
		"circle_owner", command("cmd_create", "circle_owner", 0))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Sorted so the history is the same on every run; map order is not.
	ids := make([]string, 0, len(members))
	for id := range members {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	step := 0
	next := func() circledomain.Command {
		step++
		return command(fmt.Sprintf("cmd_%d", step), "circle_owner", circle.Revision())
	}
	for _, id := range ids {
		// Everyone asks first; the owner then moves them where the test wants.
		circle, err = circle.Request(id, circledomain.Command{
			ID: fmt.Sprintf("cmd_req_%s", id), ExpectedRevision: circle.Revision(),
			ActorID: id, Kind: "test", Payload: id, RecordedAt: at,
		})
		if err != nil {
			t.Fatalf("request %s: %v", id, err)
		}
		switch members[id] {
		case circledomain.StateRequested:
			// Already there.
		case circledomain.StateMember:
			circle, err = circle.Approve(id, next())
		case circledomain.StateHost:
			if circle, err = circle.Approve(id, next()); err == nil {
				circle, err = circle.PromoteHost(id, next())
			}
		case circledomain.StateLeft:
			if circle, err = circle.Approve(id, next()); err == nil {
				circle, err = circle.Leave(id, circledomain.Command{
					ID: fmt.Sprintf("cmd_leave_%s", id), ExpectedRevision: circle.Revision(),
					ActorID: id, Kind: "test", Payload: id, RecordedAt: at,
				})
			}
		default:
			t.Fatalf("unsupported fixture state %q", members[id])
		}
		if err != nil {
			t.Fatalf("transition %s to %s: %v", id, members[id], err)
		}
	}
	return circle
}

func TestOnlyASettledMemberMayOpenACircleSource(t *testing.T) {
	// Without this a member could name any circle id and receive its roster.
	// Who is inside a private circle is exactly what this product keeps quiet.
	circle := circleWith(t, map[string]circledomain.MembershipState{
		"member_in":       circledomain.StateMember,
		"member_pending":  circledomain.StateRequested,
		"member_departed": circledomain.StateLeft,
	})
	authorizer := NewAuthorizer(stubCircles{circle: circle})

	if err := authorizer.Require(context.Background(), "member_in", "seed.source.open",
		domain.SourceCircle, "circle_1"); err != nil {
		t.Fatalf("a settled member was refused: %v", err)
	}
	for _, outsider := range []string{"member_pending", "member_departed", "stranger"} {
		if err := authorizer.Require(context.Background(), outsider, "seed.source.open",
			domain.SourceCircle, "circle_1"); !errors.Is(err, ErrNotPermitted) {
			t.Fatalf("%s was admitted", outsider)
		}
	}
}

func TestAnUnreadableCircleIsRefusedNotAssumed(t *testing.T) {
	authorizer := NewAuthorizer(stubCircles{err: errors.New("gone")})
	if err := authorizer.Require(context.Background(), "member_in", "seed.source.open",
		domain.SourceCircle, "circle_1"); !errors.Is(err, ErrNotPermitted) {
		t.Fatal("a circle that could not be read was treated as permitting")
	}
}

func TestNobodyIsIntroducedToThemselves(t *testing.T) {
	// The resolver returns the whole roster because it does not know who is
	// asking. This is where that is known, so this is where it is removed.
	circle := circleWith(t, map[string]circledomain.MembershipState{
		"member_in": circledomain.StateMember,
	})
	visibility := NewVisibility(stubCircles{circle: circle})

	visible, err := visibility.Visible(context.Background(), "member_in", "member_in",
		domain.SourceCircle, "circle_1")
	if err != nil {
		t.Fatal(err)
	}
	if visible {
		t.Fatal("a member was offered an introduction to themselves")
	}
}

func TestSomeoneWhoJustLeftIsNotOfferedThroughTheCircle(t *testing.T) {
	// The roster was read a moment earlier. Re-reading here is what stops a
	// member who walked out in between being introduced through it anyway.
	circle := circleWith(t, map[string]circledomain.MembershipState{
		"member_in":   circledomain.StateMember,
		"member_gone": circledomain.StateLeft,
	})
	visibility := NewVisibility(stubCircles{circle: circle})

	for candidate, want := range map[string]bool{
		"member_in":   true,
		"member_gone": false,
		"stranger":    false,
	} {
		visible, err := visibility.Visible(context.Background(), "asker", candidate,
			domain.SourceCircle, "circle_1")
		if err != nil {
			t.Fatal(err)
		}
		if visible != want {
			t.Fatalf("%s visible = %v, want %v", candidate, visible, want)
		}
	}
}

func TestOnlyTheSourceTypeWithAResolverIsAllowed(t *testing.T) {
	// Returning candidates for a source nothing can resolve would be a promise
	// the next step cannot keep.
	policy := NewPolicy()
	if err := policy.Allow(context.Background(), domain.SourceCircle, "circle_1"); err != nil {
		t.Fatalf("circle source refused: %v", err)
	}
	for _, unwired := range []domain.SourceType{
		domain.SourceTrust, domain.SourceHost, domain.SourceCohort,
	} {
		if err := policy.Allow(context.Background(), unwired, "ref"); !errors.Is(err, ErrNotPermitted) {
			t.Fatalf("%s was allowed with no resolver behind it", unwired)
		}
	}
}
