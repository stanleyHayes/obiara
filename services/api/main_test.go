package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	circledomain "github.com/stanleyHayes/obiara/services/api/internal/circle/domain"
	"testing"
	"time"

	podapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/application"
	poddomain "github.com/stanleyHayes/obiara/services/api/internal/seed/pod/domain"
	sowapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/sow/application"
)

// This file exists because every defect in agent_plan.md §61, §63 and §64
// lived in a bridge in main.go, and main.go had no tests at all. A bridge is
// small, but it is the only place a cross-context rule is written down, so
// "too small to test" is exactly backwards.

// blocklist answers the block question and remembers what it was asked.
type blocklist struct {
	pairs map[[2]string]bool
	err   error
	asked [][2]string
}

func (b *blocklist) IsBlocked(_ context.Context, memberID, otherID string) (bool, error) {
	b.asked = append(b.asked, [2]string{memberID, otherID})
	return b.pairs[[2]string{memberID, otherID}], b.err
}

func TestABlockIsHonouredInBothDirections(t *testing.T) {
	// A block is a decision to be apart, not a one-way filter the blocker can
	// step around. Asking one way only would let whoever did the blocking
	// keep reaching the person they blocked.
	for name, blocked := range map[string][2]string{
		"they blocked you": {"them", "you"},
		"you blocked them": {"you", "them"},
	} {
		t.Run(name, func(t *testing.T) {
			list := &blocklist{pairs: map[[2]string]bool{blocked: true}}
			isBlocked, err := sproutBlockBridge{safety: list}.Blocked(context.Background(), "you", "them")
			if err != nil || !isBlocked {
				t.Fatalf("blocked = %v, err = %v", isBlocked, err)
			}
		})
	}
}

func TestAnUnreadableBlockListIsNotAnAbsentBlock(t *testing.T) {
	list := &blocklist{err: errors.New("safety down")}
	if _, err := (sproutBlockBridge{safety: list}).Blocked(context.Background(), "you", "them"); err == nil {
		t.Fatal("a failed check was reported as no block")
	}
}

// placedPods records what delivery asked the pod context to do.
type placedPods struct {
	commands  []podapplication.Command
	proposals []podapplication.Proposal
	err       error
}

func (p *placedPods) Create(
	_ context.Context, command podapplication.Command, proposal podapplication.Proposal,
) (podapplication.Result, error) {
	p.commands = append(p.commands, command)
	p.proposals = append(p.proposals, proposal)
	return podapplication.Result{}, p.err
}

func TestDeliveringASowPlacesOnePodPerRecording(t *testing.T) {
	pods := &placedPods{}
	err := sowDeliveryBridge{pods: pods}.Place(context.Background(), sowapplication.Deliverable{
		SowID: "sow_1", SowerID: "the-sower", TargetID: "the-recipient",
		MediaRefs: []string{"recording-a", "recording-b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pods.proposals) != 2 {
		t.Fatalf("%d pods placed, want 2", len(pods.proposals))
	}
	for index, proposal := range pods.proposals {
		// The pod is the sower's recording resting at somebody's door, so it
		// is placed as them — the same act as POST /v1/seed/pods, reached
		// from the other side.
		if proposal.OwnerID != "the-sower" {
			t.Fatalf("pod %d owned by %q", index, proposal.OwnerID)
		}
		// A sow is toward one person. However many recipients a pod allows,
		// this one has exactly the person the sow named.
		if len(proposal.RecipientIDs) != 1 || proposal.RecipientIDs[0] != "the-recipient" {
			t.Fatalf("pod %d left for %v", index, proposal.RecipientIDs)
		}
		if proposal.TTL != poddomain.RestingPeriod {
			t.Fatalf("pod %d rests for %v", index, proposal.TTL)
		}
	}
	if pods.proposals[0].MediaRef != "recording-a" || pods.proposals[1].MediaRef != "recording-b" {
		t.Fatalf("carried %q and %q", pods.proposals[0].MediaRef, pods.proposals[1].MediaRef)
	}
	// Derived from the sow, so a retried delivery is recognised as the same
	// one and leaves one pod rather than two.
	if pods.commands[0].ID == pods.commands[1].ID {
		t.Fatal("two recordings shared one command id")
	}
	for _, command := range pods.commands {
		if command.ActorID != "the-sower" {
			t.Fatalf("placed as %q", command.ActorID)
		}
	}
}

func TestARetriedDeliveryAsksForTheSamePods(t *testing.T) {
	// The pod store is unique on command id, so identical ids are what makes
	// a retry leave one pod. If these ids carried a clock or a counter, every
	// retry would place another pod at somebody's house front.
	deliverable := sowapplication.Deliverable{
		SowID: "sow_1", SowerID: "s", TargetID: "t", MediaRefs: []string{"a", "b"},
	}
	first, second := &placedPods{}, &placedPods{}
	if err := (sowDeliveryBridge{pods: first}).Place(context.Background(), deliverable); err != nil {
		t.Fatal(err)
	}
	if err := (sowDeliveryBridge{pods: second}).Place(context.Background(), deliverable); err != nil {
		t.Fatal(err)
	}
	for index := range first.commands {
		if first.commands[index].ID != second.commands[index].ID {
			t.Fatalf("retry asked for %q where the first asked for %q",
				second.commands[index].ID, first.commands[index].ID)
		}
	}
}

func TestAFailedPlacementStopsRatherThanCarryingOn(t *testing.T) {
	pods := &placedPods{err: errors.New("pod store down")}
	if err := (sowDeliveryBridge{pods: pods}).Place(context.Background(), sowapplication.Deliverable{
		SowID: "sow_1", SowerID: "s", TargetID: "t", MediaRefs: []string{"a", "b"},
	}); err == nil {
		t.Fatal("a failed placement was reported as delivered")
	}
	if len(pods.commands) != 1 {
		t.Fatalf("%d placements attempted after the first failed", len(pods.commands))
	}
}

// restingFor is a pod store holding one pod for one person.
type restingFor struct {
	recipientKey, mediaRef string
	err                    error
	calls                  int
}

func (r *restingFor) HoldsFor(_ context.Context, recipientKey, mediaRef string, _ time.Time) (bool, error) {
	r.calls++
	if r.err != nil {
		return false, r.err
	}
	return recipientKey == r.recipientKey && mediaRef == r.mediaRef, nil
}

// prefixKeyer stands in for the HMAC keyer: a different namespace gives a
// different value, which is the property that matters here.
type prefixKeyer struct{}

func (prefixKeyer) Key(namespace, value string) (string, error) { return namespace + ":" + value, nil }

func podEntitlement(pods restingPods, list *blocklist) podRecipientEntitlement {
	return podRecipientEntitlement{
		pods: pods, keyer: prefixKeyer{},
		blocks: sproutBlockBridge{safety: list}, now: time.Now,
	}
}

func TestARecipientOfARestingPodMayHearIt(t *testing.T) {
	// The keyed value has to match what the pod context stored, which is why
	// the namespace is one exported constant and not a literal in two files.
	pods := &restingFor{
		recipientKey: podapplication.MemberKeyNamespace + ":the-recipient",
		mediaRef:     "recording-1",
	}
	mayHear, err := podEntitlement(pods, &blocklist{}).MayHear(
		context.Background(), "the-recipient", "the-sower", "recording-1")
	if err != nil || !mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}
}

func TestSomebodyWithNoPodMayNotHelpThemselves(t *testing.T) {
	pods := &restingFor{recipientKey: "someone-else", mediaRef: "recording-1"}
	mayHear, err := podEntitlement(pods, &blocklist{}).MayHear(
		context.Background(), "a-stranger", "the-sower", "recording-1")
	if err != nil || mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}
}

func TestABlockClosesAPodBeforeTheStoreIsEvenAsked(t *testing.T) {
	// Two people who have decided to be apart do not hear each other. Asked
	// first because a grant that ignored a block would be a way around every
	// other place the rule is applied.
	pods := &restingFor{
		recipientKey: podapplication.MemberKeyNamespace + ":the-recipient",
		mediaRef:     "recording-1",
	}
	list := &blocklist{pairs: map[[2]string]bool{{"the-recipient", "the-sower"}: true}}
	mayHear, err := podEntitlement(pods, list).MayHear(
		context.Background(), "the-recipient", "the-sower", "recording-1")
	if err != nil || mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}
	if pods.calls != 0 {
		t.Fatal("a blocked pair still had the pod store searched for them")
	}
}

func TestAnUnreadablePodStoreRefuses(t *testing.T) {
	pods := &restingFor{err: errors.New("mongo down")}
	mayHear, err := podEntitlement(pods, &blocklist{}).MayHear(
		context.Background(), "the-recipient", "the-sower", "recording-1")
	if err == nil || mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}
}

func TestAVoiceOfIntroductionIsHeardUnlessSomebodyWasShutOut(t *testing.T) {
	// This is the recording people meet each other through, so the question
	// is not who was invited to hear it but who has not been shut out.
	open := voiceOfIntroductionEntitlement{blocks: sproutBlockBridge{safety: &blocklist{}}}
	mayHear, err := open.MayHear(context.Background(), "a-listener", "the-member", "asset-1")
	if err != nil || !mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}

	shut := voiceOfIntroductionEntitlement{blocks: sproutBlockBridge{
		safety: &blocklist{pairs: map[[2]string]bool{{"the-member", "a-listener"}: true}},
	}}
	mayHear, err = shut.MayHear(context.Background(), "a-listener", "the-member", "asset-1")
	if err != nil || mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}
}

func TestAnUnreadableBlockListClosesAVoice(t *testing.T) {
	closed := voiceOfIntroductionEntitlement{blocks: sproutBlockBridge{
		safety: &blocklist{err: errors.New("safety down")},
	}}
	if mayHear, err := closed.MayHear(context.Background(), "a", "b", "asset-1"); err == nil || mayHear {
		t.Fatalf("mayHear = %v, err = %v", mayHear, err)
	}
}

// twoPersonCircle is a circle with exactly the two members a private game
// needs, which is the only shape circleGamePairResolver will pair.
type twoPersonCircle struct {
	circle circledomain.Circle
	err    error
}

func (c twoPersonCircle) Get(context.Context, string, string) (circledomain.Circle, error) {
	return c.circle, c.err
}

// circleOf builds the smallest circle a private game can happen in: an owner
// and one member. It is rehydrated through the domain rather than faked, so
// the fixture cannot describe a circle the product could not have.
func circleOf(t *testing.T, owner, member string) circledomain.Circle {
	t.Helper()
	at := time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)

	memberships := make([]circledomain.Membership, 0, 2)
	for id, state := range map[string]circledomain.MembershipState{
		owner: circledomain.StateOwner, member: circledomain.StateMember,
	} {
		built, err := circledomain.NewMembership(id, state, at)
		if err != nil {
			t.Fatal(err)
		}
		memberships = append(memberships, built)
	}

	// Founded, asked to join, admitted.
	steps := []struct {
		actor, member, from string
		to                  circledomain.MembershipState
	}{
		{owner, owner, "", circledomain.StateOwner},
		{member, member, "", circledomain.StateRequested},
		{owner, member, string(circledomain.StateRequested), circledomain.StateMember},
	}
	history := make([]circledomain.Transition, 0, len(steps))
	commands := make([]circledomain.AppliedCommand, 0, len(steps))
	for index, step := range steps {
		revision := uint64(index + 1)
		id := fmt.Sprintf("cmd_%d", revision)
		transition, err := circledomain.NewTransition(
			revision, id, step.actor, step.member, step.from, step.to, at)
		if err != nil {
			t.Fatal(err)
		}
		command, err := circledomain.NewAppliedCommand(
			id, strings.Repeat(fmt.Sprintf("%d", index), 64), revision)
		if err != nil {
			t.Fatal(err)
		}
		history = append(history, transition)
		commands = append(commands, command)
	}

	circle, err := circledomain.Rehydrate(circledomain.State{
		ID: "circle_1", Type: circledomain.TypeCommunity,
		Visibility:  circledomain.VisibilityPrivate,
		Memberships: memberships, History: history, Commands: commands,
		Revision: uint64(len(steps)), UpdatedAt: at,
	})
	if err != nil {
		t.Fatal(err)
	}
	return circle
}

func TestAPrivateCircleGamePairsTheOtherMember(t *testing.T) {
	resolver := circleGamePairResolver{
		circles: twoPersonCircle{circle: circleOf(t, "you", "them")},
		blocks:  sproutBlockBridge{safety: &blocklist{}},
	}
	other, err := resolver.Pair(context.Background(), "circle_1", "you")
	if err != nil || other != "them" {
		t.Fatalf("other = %q, err = %v", other, err)
	}
}

func TestNoPrivateGameIsPairedAcrossABlock(t *testing.T) {
	// Pairing is direct contact: it is the product putting two people
	// together, not two people happening to share a room. Every private
	// circle game goes through this one resolver — Ampe, Oware, Anansesem,
	// the competition — and so does every revalidation of a game already in
	// progress, so this is the one place the rule has to be.
	for name, blocked := range map[string][2]string{
		"they blocked you": {"them", "you"},
		"you blocked them": {"you", "them"},
	} {
		t.Run(name, func(t *testing.T) {
			resolver := circleGamePairResolver{
				circles: twoPersonCircle{circle: circleOf(t, "you", "them")},
				blocks: sproutBlockBridge{
					safety: &blocklist{pairs: map[[2]string]bool{blocked: true}},
				},
			}
			_, err := resolver.Pair(context.Background(), "circle_1", "you")
			if err == nil {
				t.Fatal("a game was paired across a block")
			}
			// Word for word the refusal a circle that is not a pair gets. The
			// difference between the two is the rejection signal a block
			// exists to withhold.
			if err.Error() != "private game requires exactly two active circle members" {
				t.Fatalf("the refusal named the block: %v", err)
			}
		})
	}
}

func TestAnUnreadableBlockListPairsNobody(t *testing.T) {
	resolver := circleGamePairResolver{
		circles: twoPersonCircle{circle: circleOf(t, "you", "them")},
		blocks:  sproutBlockBridge{safety: &blocklist{err: errors.New("safety down")}},
	}
	if _, err := resolver.Pair(context.Background(), "circle_1", "you"); err == nil {
		t.Fatal("a failed check paired them anyway")
	}
}

func TestAPairResolverWithNoBlockCheckPairsNobody(t *testing.T) {
	// A missing check is not permission — the same rule the reach rules, the
	// media policy and the sow's arrival check all follow.
	resolver := circleGamePairResolver{circles: twoPersonCircle{circle: circleOf(t, "you", "them")}}
	if _, err := resolver.Pair(context.Background(), "circle_1", "you"); err == nil {
		t.Fatal("an uncomposed check paired them anyway")
	}
}

func TestEveryWayIntoAGamePassesTheSameCheck(t *testing.T) {
	// RequireParticipant, Revalidate and RevalidateAuthors all route through
	// Pair. If one of them ever stops doing so, a game already in progress
	// would keep running across a block placed after it started.
	resolver := circleGamePairResolver{
		circles: twoPersonCircle{circle: circleOf(t, "you", "them")},
		blocks: sproutBlockBridge{
			safety: &blocklist{pairs: map[[2]string]bool{{"you", "them"}: true}},
		},
	}
	if err := resolver.RequireParticipant(context.Background(), "circle_1", "you"); err == nil {
		t.Fatal("a blocked member was admitted to a game in progress")
	}
	if err := resolver.Revalidate(context.Background(), "circle_1", "you", "them"); err == nil {
		t.Fatal("a game in progress revalidated across a block")
	}
	if err := resolver.RevalidateAuthors(context.Background(), "circle_1", "you", "them"); err == nil {
		t.Fatal("co-authorship revalidated across a block")
	}
}
