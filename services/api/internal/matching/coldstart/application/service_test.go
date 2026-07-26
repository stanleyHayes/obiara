package application

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stanleyHayes/obiara/services/api/internal/matching/coldstart/domain"
	"go.uber.org/mock/gomock"
)

func TestGenerateRevalidatesAuthorityVisibilityAndProjectsSafeReasons(t *testing.T) {
	controller := gomock.NewController(t)
	authority := NewMockAuthority(controller)
	preferences := NewMockPreferences(controller)
	trust := NewMockTrustPaths(controller)
	visibility := NewMockVisibility(controller)
	requester, first, hidden := appKey("a"), appKey("b"), appKey("c")
	preferenceInput := []domain.ReciprocalPreference{
		appPreference(hidden, true, true),
		appPreference(first, true, true),
	}
	trustInput := []domain.TrustSummary{
		{CandidateKey: hidden, Reason: domain.TrustSharedCircle, Hops: 1},
		{CandidateKey: first, Reason: domain.TrustVouched, Hops: 2},
	}
	gomock.InOrder(
		authority.EXPECT().AuthorizeColdStart(gomock.Any(), requester).Return(nil),
		preferences.EXPECT().Reciprocal(gomock.Any(), requester, domain.MaxInputCandidates).Return(preferenceInput, nil),
		trust.EXPECT().Summaries(gomock.Any(), requester, domain.MaxInputCandidates).Return(trustInput, nil),
		visibility.EXPECT().CanIntroduce(gomock.Any(), requester, first).Return(true, nil),
		visibility.EXPECT().CanIntroduce(gomock.Any(), requester, hidden).Return(false, nil),
		authority.EXPECT().AuthorizeColdStart(gomock.Any(), requester).Return(nil),
	)

	got, err := New(authority, preferences, trust, visibility).Generate(context.Background(), Request{
		RequesterKey: requester,
		Limit:        domain.MaxCandidates,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []Explanation{{
		CandidateKey: first,
		Reasons:      []domain.ReasonCode{domain.ReasonReciprocalPreference, domain.ReasonVouchedConnection},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("explanations = %+v", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"score", "rank", "path", "edge", "popular", "skill", "trait", "model"} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("projection leaked forbidden field %q: %s", forbidden, encoded)
		}
	}
}

func TestFinalAuthorityWithdrawalFailsClosed(t *testing.T) {
	controller := gomock.NewController(t)
	authority := NewMockAuthority(controller)
	preferences := NewMockPreferences(controller)
	trust := NewMockTrustPaths(controller)
	visibility := NewMockVisibility(controller)
	requester, candidate := appKey("a"), appKey("b")
	gomock.InOrder(
		authority.EXPECT().AuthorizeColdStart(gomock.Any(), requester).Return(nil),
		preferences.EXPECT().Reciprocal(gomock.Any(), requester, domain.MaxInputCandidates).Return(
			[]domain.ReciprocalPreference{appPreference(candidate, true, true)}, nil,
		),
		trust.EXPECT().Summaries(gomock.Any(), requester, domain.MaxInputCandidates).Return(
			[]domain.TrustSummary{{CandidateKey: candidate, Reason: domain.TrustKnown, Hops: 2}}, nil,
		),
		visibility.EXPECT().CanIntroduce(gomock.Any(), requester, candidate).Return(true, nil),
		authority.EXPECT().AuthorizeColdStart(gomock.Any(), requester).Return(errors.New("withdrawn")),
	)
	if _, err := New(authority, preferences, trust, visibility).Generate(context.Background(), Request{
		RequesterKey: requester,
		Limit:        10,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("withdrawal = %v", err)
	}
}

func TestVisibilityDependencyFailureFailsClosed(t *testing.T) {
	controller := gomock.NewController(t)
	authority := NewMockAuthority(controller)
	preferences := NewMockPreferences(controller)
	trust := NewMockTrustPaths(controller)
	visibility := NewMockVisibility(controller)
	requester, candidate := appKey("a"), appKey("b")
	authority.EXPECT().AuthorizeColdStart(gomock.Any(), requester).Return(nil)
	preferences.EXPECT().Reciprocal(gomock.Any(), requester, domain.MaxInputCandidates).Return(
		[]domain.ReciprocalPreference{appPreference(candidate, true, true)}, nil,
	)
	trust.EXPECT().Summaries(gomock.Any(), requester, domain.MaxInputCandidates).Return(
		[]domain.TrustSummary{{CandidateKey: candidate, Reason: domain.TrustKnown, Hops: 2}}, nil,
	)
	visibility.EXPECT().CanIntroduce(gomock.Any(), requester, candidate).Return(false, errors.New("unavailable"))
	if _, err := New(authority, preferences, trust, visibility).Generate(context.Background(), Request{
		RequesterKey: requester,
		Limit:        10,
	}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("visibility failure = %v", err)
	}
}

func appPreference(candidate string, requester, candidateAllows bool) domain.ReciprocalPreference {
	return domain.ReciprocalPreference{
		CandidateKey:               candidate,
		RequesterExplicit:          requester,
		CandidateExplicit:          candidateAllows,
		RequesterPreferenceVersion: 1,
		CandidatePreferenceVersion: 1,
	}
}

func appKey(character string) string {
	return strings.Repeat(character, 64)
}
