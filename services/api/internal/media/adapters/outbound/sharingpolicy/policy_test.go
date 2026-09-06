package sharingpolicy

import (
	"context"
	"errors"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
)

// says is a fixed entitlement answer.
type says struct {
	yes   bool
	err   error
	calls int
}

func (s *says) MayHear(context.Context, string, string, string) (bool, error) {
	s.calls++
	return s.yes, s.err
}

func read(subject, owner, purpose string) application.AccessDecision {
	return application.AccessDecision{
		SubjectID: subject, OwnerID: owner, AssetID: "asset-1",
		Purpose: purpose, Action: application.ActionRead,
	}
}

func TestTheOwnerNeedsNoEntitlement(t *testing.T) {
	// The floor ownerpolicy set, kept: your own recording is yours, and no
	// bridge has to be reachable for you to hear it back.
	entitlement := &says{}
	policy := New([]string{"voice.introduction"},
		map[string]Entitlement{"voice.introduction": entitlement})
	if err := policy.Authorize(context.Background(), read("member-1", "member-1", "voice.introduction")); err != nil {
		t.Fatalf("owner refused: %v", err)
	}
	if entitlement.calls != 0 {
		t.Fatal("the owner was asked to prove an entitlement to their own voice")
	}
}

func TestAnEntitledListenerIsAdmitted(t *testing.T) {
	policy := New([]string{"seed.pod.playback"},
		map[string]Entitlement{"seed.pod.playback": &says{yes: true}})
	if err := policy.Authorize(context.Background(), read("recipient", "sower", "seed.pod.playback")); err != nil {
		t.Fatalf("an entitled listener was refused: %v", err)
	}
}

func TestEveryOtherAnswerRefuses(t *testing.T) {
	for name, entitlement := range map[string]Entitlement{
		"the entitlement says no":       &says{},
		"the entitlement cannot answer": &says{yes: true, err: errors.New("down")},
	} {
		t.Run(name, func(t *testing.T) {
			policy := New([]string{"seed.pod.playback"},
				map[string]Entitlement{"seed.pod.playback": entitlement})
			if err := policy.Authorize(context.Background(),
				read("stranger", "owner", "seed.pod.playback")); !errors.Is(err, application.ErrAccessDenied) {
				t.Fatalf("err = %v, want ErrAccessDenied", err)
			}
		})
	}
}

func TestAPurposeWithNoEntitlementStaysOwnerOnly(t *testing.T) {
	// The direction that matters when somebody adds a purpose and forgets
	// the bridge: it must close, not open.
	policy := New([]string{"voice.introduction", "something.new"},
		map[string]Entitlement{"voice.introduction": &says{yes: true}})
	if err := policy.Authorize(context.Background(),
		read("listener", "owner", "something.new")); !errors.Is(err, application.ErrAccessDenied) {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
	// And a nil entitlement is a missing wire, not permission.
	nilWired := New([]string{"voice.introduction"},
		map[string]Entitlement{"voice.introduction": nil})
	if err := nilWired.Authorize(context.Background(),
		read("listener", "owner", "voice.introduction")); !errors.Is(err, application.ErrAccessDenied) {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
}

func TestAnUnlistedPurposeIsRefusedEvenForTheOwner(t *testing.T) {
	// Carried over from ownerpolicy: a wildcard here would give any future
	// caller access the moment it invented a purpose string.
	policy := New([]string{"voice.introduction"}, nil)
	if err := policy.Authorize(context.Background(),
		read("member-1", "member-1", "anything.else")); !errors.Is(err, application.ErrAccessDenied) {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
	if err := New(nil, nil).Authorize(context.Background(),
		read("member-1", "member-1", "voice.introduction")); !errors.Is(err, application.ErrAccessDenied) {
		t.Fatal("an empty policy must admit nothing")
	}
}

func TestSharingIsReadingAndNeverWriting(t *testing.T) {
	// An entitlement says somebody may hear a recording. It must not also
	// let them upload as its owner.
	entitlement := &says{yes: true}
	policy := New([]string{"voice.introduction"},
		map[string]Entitlement{"voice.introduction": entitlement})
	if err := policy.Authorize(context.Background(), application.AccessDecision{
		SubjectID: "listener", OwnerID: "owner", AssetID: "asset-1",
		Purpose: "voice.introduction", Action: application.ActionUpload,
	}); !errors.Is(err, application.ErrAccessDenied) {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
	if entitlement.calls != 0 {
		t.Fatal("an upload as somebody else got as far as asking")
	}
}

func TestAMissingSubjectOrOwnerRefuses(t *testing.T) {
	policy := New([]string{"voice.introduction"},
		map[string]Entitlement{"voice.introduction": &says{yes: true}})
	for name, decision := range map[string]application.AccessDecision{
		"no subject": read("", "owner", "voice.introduction"),
		"no owner":   read("listener", "", "voice.introduction"),
	} {
		t.Run(name, func(t *testing.T) {
			if err := policy.Authorize(context.Background(), decision); !errors.Is(err, application.ErrAccessDenied) {
				t.Fatalf("err = %v, want ErrAccessDenied", err)
			}
		})
	}
}
