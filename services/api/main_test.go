package main

import (
	"context"
	"errors"
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
