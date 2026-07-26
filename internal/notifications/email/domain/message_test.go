package domain

import (
	"strings"
	"testing"
)

func TestNewMessageValidation(t *testing.T) {
	if _, err := NewMessage("not-an-email", TemplateOpsAlert, nil); err != ErrInvalidRecipient {
		t.Fatalf("bad recipient = %v", err)
	}
	if _, err := NewMessage("ops@example.test", Template("marketing"), nil); err != ErrInvalidTemplate {
		t.Fatalf("marketing template = %v, want rejected", err)
	}
	if _, err := NewMessage("ops@example.test", TemplateOpsAlert, map[string]string{"x": strings.Repeat("y", 501)}); err != ErrParamTooLong {
		t.Fatalf("long param = %v", err)
	}
	message, err := NewMessage("ops@example.test", TemplateVerificationHelp, map[string]string{"eta": "10 minutes"})
	if err != nil {
		t.Fatal(err)
	}
	if message.Template() != TemplateVerificationHelp || message.Params()["eta"] != "10 minutes" {
		t.Fatalf("message = %#v", message)
	}
}
