package domain

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestProjectRequiresReciprocalPreferenceAndScopedTrust(t *testing.T) {
	requester := key("a")
	reciprocal := key("b")
	oneSided := key("c")
	noTrust := key("d")
	candidates, err := Project(requester, []ReciprocalPreference{
		preference(reciprocal, true, true),
		preference(oneSided, true, false),
		preference(noTrust, true, true),
	}, []TrustSummary{
		{CandidateKey: reciprocal, Reason: TrustVouched, Hops: 2},
		{CandidateKey: oneSided, Reason: TrustSharedCircle, Hops: 1},
	}, MaxCandidates)
	if err != nil {
		t.Fatal(err)
	}
	want := []Candidate{{
		CandidateKey: reciprocal,
		Reasons:      []ReasonCode{ReasonReciprocalPreference, ReasonVouchedConnection},
	}}
	if !reflect.DeepEqual(candidates, want) {
		t.Fatalf("candidates = %+v, want %+v", candidates, want)
	}
}

func TestProjectionIsStableNotRanked(t *testing.T) {
	requester := key("a")
	first := key("b")
	second := key("c")
	preferences := []ReciprocalPreference{
		preference(second, true, true),
		preference(first, true, true),
	}
	trust := []TrustSummary{
		{CandidateKey: second, Reason: TrustHost, Hops: 4},
		{CandidateKey: first, Reason: TrustKnown, Hops: 2},
		{CandidateKey: first, Reason: TrustSharedCircle, Hops: 1},
		{CandidateKey: first, Reason: TrustSharedCircle, Hops: 3},
	}
	got, err := Project(requester, preferences, trust, MaxCandidates)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].CandidateKey != first || got[1].CandidateKey != second {
		t.Fatalf("opaque-key enumeration = %+v", got)
	}
	if !reflect.DeepEqual(got[0].Reasons, []ReasonCode{
		ReasonReciprocalPreference,
		ReasonSharedCircle,
		ReasonKnownConnection,
	}) {
		t.Fatalf("bounded reasons = %v", got[0].Reasons)
	}
}

func TestProjectionRejectsSelfDuplicateRawOrUnboundedInput(t *testing.T) {
	requester := key("a")
	duplicate := preference(key("b"), true, true)
	cases := []struct {
		preferences []ReciprocalPreference
		trust       []TrustSummary
	}{
		{preferences: []ReciprocalPreference{preference(requester, true, true)}},
		{preferences: []ReciprocalPreference{duplicate, duplicate}},
		{
			preferences: []ReciprocalPreference{preference(key("b"), true, true)},
			trust:       []TrustSummary{{CandidateKey: key("b"), Reason: "raw_path", Hops: 1}},
		},
		{
			preferences: []ReciprocalPreference{preference(key("b"), true, true)},
			trust:       []TrustSummary{{CandidateKey: key("b"), Reason: TrustKnown, Hops: MaxTrustHops + 1}},
		},
	}
	for index, test := range cases {
		if _, err := Project(requester, test.preferences, test.trust, MaxCandidates); err == nil {
			t.Fatalf("case %d accepted invalid input", index)
		}
	}
	tooMany := make([]ReciprocalPreference, MaxInputCandidates+1)
	if _, err := Project(requester, tooMany, nil, MaxCandidates); err != ErrInputBound {
		t.Fatalf("unbounded input = %v", err)
	}
}

func TestProjectionOrderProperty(t *testing.T) {
	requester := key("a")
	preferences := make([]ReciprocalPreference, 50)
	trust := make([]TrustSummary, 50)
	allowed := []TrustReason{TrustSharedCircle, TrustVouched, TrustKnown, TrustHost}
	for index := range preferences {
		candidate := numberedKey(index + 1)
		preferences[index] = preference(candidate, true, true)
		trust[index] = TrustSummary{CandidateKey: candidate, Reason: allowed[index%len(allowed)], Hops: index%MaxTrustHops + 1}
	}
	want, err := Project(requester, preferences, trust, MaxCandidates)
	if err != nil {
		t.Fatal(err)
	}
	rng := rand.New(rand.NewSource(20260726))
	for trial := 0; trial < 1000; trial++ {
		shuffledPreferences := slices.Clone(preferences)
		shuffledTrust := slices.Clone(trust)
		rng.Shuffle(len(shuffledPreferences), func(i, j int) {
			shuffledPreferences[i], shuffledPreferences[j] = shuffledPreferences[j], shuffledPreferences[i]
		})
		rng.Shuffle(len(shuffledTrust), func(i, j int) {
			shuffledTrust[i], shuffledTrust[j] = shuffledTrust[j], shuffledTrust[i]
		})
		got, projectErr := Project(requester, shuffledPreferences, shuffledTrust, MaxCandidates)
		if projectErr != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("trial %d changed projection: err=%v", trial, projectErr)
		}
	}
}

func TestPureProjectionIsRaceSafe(t *testing.T) {
	requester := key("a")
	preferences := []ReciprocalPreference{preference(key("b"), true, true)}
	trust := []TrustSummary{{CandidateKey: key("b"), Reason: TrustKnown, Hops: 2}}
	want, err := Project(requester, preferences, trust, MaxCandidates)
	if err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 100 {
				got, projectErr := Project(requester, preferences, trust, MaxCandidates)
				if projectErr != nil || !reflect.DeepEqual(got, want) {
					t.Errorf("concurrent projection: %v", projectErr)
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzProjectionNeverAdmitsOneSidedOrPathless(f *testing.F) {
	f.Add(true, true, uint8(1), uint8(1))
	f.Add(true, false, uint8(2), uint8(4))
	f.Fuzz(func(t *testing.T, requesterExplicit, candidateExplicit bool, reasonRaw, hopsRaw uint8) {
		requester := key("a")
		candidate := key("b")
		allowed := []TrustReason{TrustSharedCircle, TrustVouched, TrustKnown, TrustHost}
		hops := int(hopsRaw%MaxTrustHops) + 1
		got, err := Project(
			requester,
			[]ReciprocalPreference{preference(candidate, requesterExplicit, candidateExplicit)},
			[]TrustSummary{{CandidateKey: candidate, Reason: allowed[int(reasonRaw)%len(allowed)], Hops: hops}},
			MaxCandidates,
		)
		if err != nil {
			t.Fatal(err)
		}
		wantCount := 0
		if requesterExplicit && candidateExplicit {
			wantCount = 1
		}
		if len(got) != wantCount {
			t.Fatalf("one-sided/path rule candidates=%+v", got)
		}
	})
}

func preference(candidate string, requester, candidateAllows bool) ReciprocalPreference {
	return ReciprocalPreference{
		CandidateKey:               candidate,
		RequesterExplicit:          requester,
		CandidateExplicit:          candidateAllows,
		RequesterPreferenceVersion: 1,
		CandidatePreferenceVersion: 1,
	}
}

func key(character string) string {
	return strings.Repeat(character, 64)
}

func numberedKey(number int) string {
	const hexadecimal = "0123456789abcdef"
	prefix := hexadecimal[(number/16)%16:(number/16)%16+1] + hexadecimal[number%16:number%16+1]
	return strings.Repeat("0", 62) + prefix
}
