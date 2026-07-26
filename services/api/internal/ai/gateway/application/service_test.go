package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/ai/gateway/domain"
	"go.uber.org/mock/gomock"
)

type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

func approvedPolicy(t *testing.T) domain.Policy {
	t.Helper()
	policy, err := domain.NewPolicy(
		"grounded resonance", "GH", domain.CapabilityResonance,
		[]domain.DataClass{domain.DataConsentedText},
		domain.Model{
			Vendor: "approved-vendor", Name: "approved-model",
			Regions:       []string{"GH"},
			Capabilities:  []domain.Capability{domain.CapabilityResonance},
			DataClasses:   []domain.DataClass{domain.DataConsentedText},
			MaxInputBytes: 4096, MaxOutputTokens: 256,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func TestExecuteRevalidatesConsentRedactsAndAuditsMetadataOnly(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	consent := NewMockConsentGate(ctrl)
	redactor := NewMockRedactor(ctrl)
	vendor := NewMockVendor(ctrl)
	audit := NewMockAuditSink(ctrl)
	now := time.Date(2026, 7, 26, 22, 0, 0, 0, time.UTC)
	service := NewService(consent, redactor, vendor, audit, fixedClock{at: now})
	policy := approvedPolicy(t)
	cmd := ExecuteCommand{
		RequestID: "req_8Km2qP4vN7xR5tZa", ActorKey: "actor_hmac_8Km2qP4v",
		Input: "My email is private@example.com", MaxOutputTokens: 128, Policy: policy,
	}

	gomock.InOrder(
		consent.EXPECT().RequireCurrent(gomock.Any(), cmd.ActorKey, policy.Purpose, policy.DataClasses),
		redactor.EXPECT().Redact(gomock.Any(), cmd.Input, policy.DataClasses).Return("My email is [redacted]", 1, nil),
		vendor.EXPECT().Invoke(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, request VendorRequest) (VendorResponse, error) {
			if strings.Contains(request.Input, "private@example.com") {
				t.Fatal("raw identifier reached vendor")
			}
			return VendorResponse{Output: "A grounded explanation.", VendorRequestID: "vendor-opaque"}, nil
		}),
		audit.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, record AuditRecord) error {
			if record.Outcome != "completed" || record.Redactions != 1 || record.OccurredAt != now {
				t.Fatalf("audit record = %+v", record)
			}
			serialized := strings.ToLower(strings.Join([]string{
				record.RequestID, record.ActorKey, record.Purpose, record.Vendor,
				record.Model, record.Region, record.Outcome, record.ReasonCode,
			}, " "))
			if strings.Contains(serialized, "private@example.com") || strings.Contains(serialized, "grounded explanation") {
				t.Fatal("audit persisted content")
			}
			return nil
		}),
	)

	output, err := service.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if output != "A grounded explanation." {
		t.Fatalf("output = %q", output)
	}
}

func TestExecuteFailsClosedBeforeVendorWithoutConsent(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	consent := NewMockConsentGate(ctrl)
	redactor := NewMockRedactor(ctrl)
	vendor := NewMockVendor(ctrl)
	audit := NewMockAuditSink(ctrl)
	service := NewService(consent, redactor, vendor, audit, fixedClock{at: time.Now()})
	policy := approvedPolicy(t)
	cmd := ExecuteCommand{RequestID: "req_8Km2qP4vN7xR5tZa", ActorKey: "actor_hmac_8Km2qP4v", Input: "consented text", MaxOutputTokens: 32, Policy: policy}
	consent.EXPECT().RequireCurrent(gomock.Any(), cmd.ActorKey, policy.Purpose, policy.DataClasses).Return(errors.New("revoked"))
	audit.EXPECT().Record(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, record AuditRecord) error {
		if record.Outcome != "denied" || record.ReasonCode != "consent" {
			t.Fatalf("audit = %+v", record)
		}
		return nil
	})
	if _, err := service.Execute(context.Background(), cmd); !errors.Is(err, domain.ErrNotConsented) {
		t.Fatalf("error = %v", err)
	}
}

func TestExecuteHasNoSilentVendorFallback(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	consent := NewMockConsentGate(ctrl)
	redactor := NewMockRedactor(ctrl)
	vendor := NewMockVendor(ctrl)
	audit := NewMockAuditSink(ctrl)
	service := NewService(consent, redactor, vendor, audit, fixedClock{at: time.Now()})
	policy := approvedPolicy(t)
	cmd := ExecuteCommand{RequestID: "req_8Km2qP4vN7xR5tZa", ActorKey: "actor_hmac_8Km2qP4v", Input: "safe input", MaxOutputTokens: 32, Policy: policy}
	vendorErr := errors.New("vendor unavailable")
	gomock.InOrder(
		consent.EXPECT().RequireCurrent(gomock.Any(), cmd.ActorKey, policy.Purpose, policy.DataClasses),
		redactor.EXPECT().Redact(gomock.Any(), cmd.Input, policy.DataClasses).Return(cmd.Input, 0, nil),
		vendor.EXPECT().Invoke(gomock.Any(), gomock.Any()).Return(VendorResponse{}, vendorErr),
		audit.EXPECT().Record(gomock.Any(), gomock.Any()),
	)
	if _, err := service.Execute(context.Background(), cmd); !errors.Is(err, vendorErr) {
		t.Fatalf("error = %v", err)
	}
}
