package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
)

// TestStartLoginReportsDeliveryFailure covers the case that made admin
// sign-in undiagnosable: the code was minted, the email provider rejected
// it, and the operator saw a bare 500 with nothing in the logs.
func TestStartLoginReportsDeliveryFailure(t *testing.T) {
	service, principals, challenges, _, _, sender := newService(t)
	providerErr := errors.New("resend: delivery failed: status 403, provider error validation_error")

	principals.EXPECT().FindByEmail(gomock.Any(), "root@example.test").Return(passwordPrincipal(t), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), "root@example.test", gomock.Any()).Return(providerErr)

	err := service.StartLogin(context.Background(), "root@example.test", testPassword)
	if err == nil {
		t.Fatal("StartLogin reported success while the code was never delivered")
	}
	if !errors.Is(err, ErrCodeDeliveryFailed) {
		t.Errorf("error %v should wrap ErrCodeDeliveryFailed so transport can map it to a 503", err)
	}
	// The provider cause must survive for the logs, or triage has nothing.
	if !errors.Is(err, providerErr) {
		t.Errorf("error %v lost the provider cause", err)
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error %v should carry the provider status for triage", err)
	}
}

// TestStepUpReportsDeliveryFailure covers the same path for step-up, which
// gates every privileged action.
func TestStepUpReportsDeliveryFailure(t *testing.T) {
	service, principals, challenges, sessions, _, sender := newService(t)
	sessions.EXPECT().FindByID(gomock.Any(), "sess_root").
		Return(steppedUpSession("sess_root", "adm_root", nil), nil)
	principals.EXPECT().FindByID(gomock.Any(), "adm_root").Return(adminPrincipal(), nil)
	challenges.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	sender.EXPECT().SendMfaCode(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("resend: delivery failed: transport"))

	if err := service.StepUpStart(context.Background(), "sess_root"); !errors.Is(err, ErrCodeDeliveryFailed) {
		t.Fatalf("StepUpStart error = %v, want ErrCodeDeliveryFailed", err)
	}
}
