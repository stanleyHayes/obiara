package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/host/domain"
	"go.uber.org/mock/gomock"
)

var serviceTime = time.Date(2026, 7, 26, 20, 0, 0, 0, time.UTC)

func submit() SubmitRequest {
	return SubmitRequest{
		CommandID: "submission:1", ApplicantID: "member:1",
		ProofReference: "evidence:1", InstitutionKind: domain.InstitutionUniversity,
		IssuerID: "issuer:1", IssuedAt: serviceTime.Add(-time.Hour),
		ExpiresAt: serviceTime.Add(180 * 24 * time.Hour),
	}
}

func expectSubmission(repository *MockRepository, keyer *MockKeyer, ids *MockIDSource) {
	keyer.EXPECT().Key("host_applicant", "member:1").Return(strings.Repeat("a", 64), nil)
	keyer.EXPECT().Key("institution_issuer", "issuer:1").Return(strings.Repeat("b", 64), nil)
	ids.EXPECT().NewID().Return("host:1")
	repository.EXPECT().Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, value domain.Application) (domain.Application, bool, error) {
			return value, false, nil
		})
}

func TestVerifiedProviderIsTheOnlyAutomaticApproval(t *testing.T) {
	controller := gomock.NewController(t)
	repository := NewMockRepository(controller)
	provider := NewMockInstitutionalProvider(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	expectSubmission(repository, keyer, ids)
	provider.EXPECT().Verify(gomock.Any(), gomock.Any()).
		Return(ProviderResult{Outcome: OutcomeVerified, ProviderRef: "provider:proof"}, nil)
	repository.EXPECT().Update(gomock.Any(), gomock.Cond(func(value domain.Application) bool {
		return value.Status() == domain.StatusApproved && value.Eligible(serviceTime)
	}), uint64(1), "submission:1.provider").Return(nil)

	result, err := NewService(
		repository, provider, NewMockManualReviewQueue(controller), keyer, ids,
		func() time.Time { return serviceTime },
	).Submit(context.Background(), submit())
	if err != nil || !result.Application.Eligible(serviceTime) {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
}

func TestOutageAndUncertaintyQueueWithoutSilentPass(t *testing.T) {
	tests := []struct {
		name   string
		result ProviderResult
		err    error
		reason domain.Reason
	}{
		{"outage", ProviderResult{}, errors.New("provider down"), domain.ReasonProviderUnavailable},
		{"uncertain", ProviderResult{Outcome: OutcomeUncertain, ProviderRef: "provider:uncertain"}, nil, domain.ReasonProviderUncertain},
		{"unknown", ProviderResult{Outcome: "new_vendor_state"}, nil, domain.ReasonProviderUncertain},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			repository := NewMockRepository(controller)
			provider := NewMockInstitutionalProvider(controller)
			reviews := NewMockManualReviewQueue(controller)
			keyer := NewMockKeyer(controller)
			ids := NewMockIDSource(controller)
			expectSubmission(repository, keyer, ids)
			provider.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(test.result, test.err)
			repository.EXPECT().Update(gomock.Any(), gomock.Cond(func(value domain.Application) bool {
				return value.Status() == domain.StatusQueuedManual && value.Reason() == test.reason &&
					!value.Eligible(serviceTime)
			}), uint64(1), "submission:1.provider").Return(nil)
			reviews.EXPECT().Enqueue(gomock.Any(), ReviewTask{
				ApplicationID: "host:1", ProofReference: "evidence:1", Reason: test.reason,
			}).Return(nil)

			result, err := NewService(
				repository, provider, reviews, keyer, ids,
				func() time.Time { return serviceTime },
			).Submit(context.Background(), submit())
			if result.Application.Eligible(serviceTime) || !errors.Is(err, ErrManualReviewRequired) {
				t.Fatalf("uncertain result=%+v, err=%v", result, err)
			}
		})
	}
}
