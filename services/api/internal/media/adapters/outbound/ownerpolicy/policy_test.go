package ownerpolicy

import (
	"context"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/media/application"
)

func TestOnlyTheOwnerIsAdmitted(t *testing.T) {
	policy := New("voice.introduction")
	if err := policy.Authorize(context.Background(), application.AccessDecision{
		SubjectID: "member-1", OwnerID: "member-1",
		Purpose: "voice.introduction", Action: application.ActionRead,
	}); err != nil {
		t.Fatalf("owner refused: %v", err)
	}
	for name, decision := range map[string]application.AccessDecision{
		"another member": {SubjectID: "member-2", OwnerID: "member-1", Purpose: "voice.introduction"},
		"no subject":     {SubjectID: "", OwnerID: "member-1", Purpose: "voice.introduction"},
		"no owner":       {SubjectID: "member-1", OwnerID: "", Purpose: "voice.introduction"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := policy.Authorize(context.Background(), decision); err != application.ErrAccessDenied {
				t.Fatalf("err = %v, want ErrAccessDenied", err)
			}
		})
	}
}

func TestAnUnknownPurposeIsRefusedNotWaved(t *testing.T) {
	// A wildcard here would give any future caller access to every member's
	// media the moment it invented a purpose string.
	policy := New("voice.introduction")
	if err := policy.Authorize(context.Background(), application.AccessDecision{
		SubjectID: "member-1", OwnerID: "member-1", Purpose: "anything.else",
	}); err != application.ErrAccessDenied {
		t.Fatalf("err = %v, want ErrAccessDenied", err)
	}
	if err := New().Authorize(context.Background(), application.AccessDecision{
		SubjectID: "member-1", OwnerID: "member-1", Purpose: "voice.introduction",
	}); err != application.ErrAccessDenied {
		t.Fatal("an empty policy must admit nothing")
	}
}
