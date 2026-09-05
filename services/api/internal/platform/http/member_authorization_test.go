package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	safetydomain "github.com/stanleyHayes/obiara/internal/safety/domain"
	circleroomapplication "github.com/stanleyHayes/obiara/services/api/internal/circle/room/application"
	circleroomdomain "github.com/stanleyHayes/obiara/services/api/internal/circle/room/domain"
	consentdomain "github.com/stanleyHayes/obiara/services/api/internal/consent/consentmap/domain"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
	profiledomain "github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
	gardendomain "github.com/stanleyHayes/obiara/services/api/internal/seed/garden/domain"
	listeningapplication "github.com/stanleyHayes/obiara/services/api/internal/seed/listening/application"
	listeningdomain "github.com/stanleyHayes/obiara/services/api/internal/seed/listening/domain"
	subandomain "github.com/stanleyHayes/obiara/services/api/internal/suban/domain"
)

type circleRoomCapture struct{ actorID string }

func (*circleRoomCapture) Voice(context.Context, circleroomapplication.Create) (circleroomdomain.Entry, error) {
	return circleroomdomain.Entry{}, nil
}
func (*circleRoomCapture) Event(context.Context, circleroomapplication.Create) (circleroomdomain.Entry, error) {
	return circleroomdomain.Entry{}, nil
}
func (*circleRoomCapture) Notice(context.Context, circleroomapplication.Create) (circleroomdomain.Entry, error) {
	return circleroomdomain.Entry{}, nil
}
func (capture *circleRoomCapture) List(_ context.Context, _, actorID string, _ int) ([]circleroomdomain.Entry, error) {
	capture.actorID = actorID
	return nil, nil
}
func (*circleRoomCapture) Delete(context.Context, string, string, string) (circleroomdomain.Entry, error) {
	return circleroomdomain.Entry{}, nil
}

type consentCapture struct {
	memberID string
	purpose  consentdomain.Purpose
	enabled  bool
}

func (capture *consentCapture) Switchboard(_ context.Context, memberID string) (map[consentdomain.Purpose]bool, error) {
	capture.memberID = memberID
	return map[consentdomain.Purpose]bool{consentdomain.PurposeIdentitySafety: true}, nil
}

func (capture *consentCapture) Set(_ context.Context, memberID string, purpose consentdomain.Purpose, enabled bool) (bool, error) {
	capture.memberID, capture.purpose, capture.enabled = memberID, purpose, enabled
	return enabled, nil
}

type safetyCapture struct {
	reporter string
	subject  string
	surface  safetydomain.Surface
}

func (capture *safetyCapture) File(
	_ context.Context,
	reporterID, subjectID string,
	_ safetydomain.Category,
	surface safetydomain.Surface,
	_, _ string,
) (string, safetydomain.Tier, error) {
	capture.reporter, capture.subject, capture.surface = reporterID, subjectID, surface
	return "report-1", safetydomain.TierA, nil
}
func (*safetyCapture) Block(context.Context, string, string) error   { return nil }
func (*safetyCapture) Unblock(context.Context, string, string) error { return nil }

type subanCapture struct{ subject string }

func (capture *subanCapture) Marks(_ context.Context, subject string) ([]subandomain.Mark, error) {
	capture.subject = subject
	return nil, nil
}
func (*subanCapture) Events(context.Context, string) ([]subandomain.Event, error) {
	return nil, nil
}

func memberSession(now time.Time) identitydomain.Session {
	return identitydomain.Reconstitute(
		"session-1", "member-1", "device-1", identitydomain.StatusActive,
		"access", now.Add(time.Hour), "refresh", "", now.Add(24*time.Hour),
		1, now, now,
	)
}

func TestSafetyReportDerivesReporterFromSession(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &safetyCapture{}
	mux := http.NewServeMux()
	RegisterSafetyRoutes(
		mux, capture,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
	)
	request := httptest.NewRequest(
		http.MethodPost, "/v1/reports",
		strings.NewReader(`{"subjectId":"member-2","category":"minor_safety","surface":"game"}`),
	)
	request.Header.Set("Authorization", "Bearer access")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if capture.reporter != "member-1" || capture.subject != "member-2" ||
		capture.surface != safetydomain.SurfaceGame {
		t.Fatalf("capture=%+v", capture)
	}
}

func TestConsentSwitchboardDerivesMemberFromSession(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &consentCapture{}
	mux := http.NewServeMux()
	RegisterConsentRoutes(
		mux, capture,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
	)

	read := httptest.NewRequest(http.MethodGet, "/v1/consent", nil)
	read.Header.Set("Authorization", "Bearer access")
	readResponse := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(readResponse, read)
	if readResponse.Code != http.StatusOK || capture.memberID != "member-1" {
		t.Fatalf("read status=%d member=%q body=%s", readResponse.Code, capture.memberID, readResponse.Body.String())
	}

	update := httptest.NewRequest(
		http.MethodPut, "/v1/consent/purposes/product_analytics",
		strings.NewReader(`{"enabled":false}`),
	)
	update.Header.Set("Authorization", "Bearer access")
	update.Header.Set("Content-Type", "application/json")
	updateResponse := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(updateResponse, update)
	if updateResponse.Code != http.StatusOK || capture.memberID != "member-1" ||
		capture.purpose != consentdomain.PurposeProductAnalytics || capture.enabled {
		t.Fatalf(
			"update status=%d member=%q purpose=%q enabled=%v body=%s",
			updateResponse.Code, capture.memberID, capture.purpose, capture.enabled, updateResponse.Body.String(),
		)
	}
}

func TestCircleRoomListDerivesActorFromSession(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &circleRoomCapture{}
	mux := http.NewServeMux()
	RegisterCircleRoomRoutes(
		mux, capture,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
		verifiedGate(),
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/circles/circle-1/room", nil)
	request.Header.Set("Authorization", "Bearer access")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusOK || capture.actorID != "member-1" {
		t.Fatalf("status=%d actor=%q body=%s", response.Code, capture.actorID, response.Body.String())
	}
}

func TestSubanRejectsCrossMemberRead(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &subanCapture{}
	mux := http.NewServeMux()
	RegisterSubanRoutes(
		mux, capture, nil,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/suban/marks/member-2", nil)
	request.Header.Set("Authorization", "Bearer access")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || capture.subject != "" {
		t.Fatalf("status=%d subject=%q body=%s", response.Code, capture.subject, response.Body.String())
	}
}

type doorwayCapture struct{ memberID string }

func (capture *doorwayCapture) Set(_ context.Context, memberID, text string, custom bool) (profiledomain.DoorwayQuestion, error) {
	capture.memberID = memberID
	return profiledomain.NewDoorwayQuestion(memberID, text, custom, time.Now())
}
func (*doorwayCapture) Get(context.Context, string) (profiledomain.DoorwayQuestion, error) {
	return profiledomain.DoorwayQuestion{}, nil
}

type vaultCapture struct{}

func (*vaultCapture) Add(context.Context, string, string, int) (profiledomain.VaultItem, error) {
	return profiledomain.VaultItem{}, nil
}
func (*vaultCapture) ViewFor(context.Context, string, string) ([]profiledomain.VaultItemView, error) {
	return nil, nil
}

func TestDoorwayQuestionDerivesOwnerFromSession(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &doorwayCapture{}
	mux := http.NewServeMux()
	RegisterDoorwayRoutes(
		mux, capture, &vaultCapture{},
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
	)
	request := httptest.NewRequest(http.MethodPut, "/v1/doorway-question", strings.NewReader(`{"text":"What feels like home?","custom":true}`))
	request.Header.Set("Authorization", "Bearer access")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusOK || capture.memberID != "member-1" {
		t.Fatalf("status=%d member=%q body=%s", response.Code, capture.memberID, response.Body.String())
	}
}

type listeningCapture struct{ listenerID string }

func (capture *listeningCapture) RecordHeartbeats(_ context.Context, listenerID, assetID string, duration float64, ranges []listeningapplication.HeartbeatRange) (listeningdomain.Playback, error) {
	capture.listenerID = listenerID
	playback, err := listeningdomain.NewPlayback(listenerID, assetID, duration)
	if err != nil {
		return listeningdomain.Playback{}, err
	}
	for _, heartbeat := range ranges {
		if err := playback.Record(heartbeat.Start, heartbeat.End); err != nil {
			return listeningdomain.Playback{}, err
		}
	}
	return playback, nil
}
func (*listeningCapture) Eligibility(context.Context, string, string) (bool, float64, error) {
	return false, 0, nil
}

func TestListeningDerivesListenerFromSession(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &listeningCapture{}
	mux := http.NewServeMux()
	RegisterListeningRoutes(
		mux, capture,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
		verifiedGate(),
	)
	request := httptest.NewRequest(http.MethodPost, "/v1/listening/heartbeats", strings.NewReader(`{"voiceAssetId":"asset-1","assetDuration":30,"ranges":[{"start":0,"end":5}]}`))
	request.Header.Set("Authorization", "Bearer access")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusOK || capture.listenerID != "member-1" {
		t.Fatalf("status=%d listener=%q body=%s", response.Code, capture.listenerID, response.Body.String())
	}
}

type gardenCapture struct{ memberID string }

func (capture *gardenCapture) Summary(_ context.Context, memberID string) (gardendomain.Summary, error) {
	capture.memberID = memberID
	return gardendomain.Summary{
		AsOf:    time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC),
		Message: "Nothing needs your attention today.",
	}, nil
}

func TestGardenDerivesOwnerFromSession(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	capture := &gardenCapture{}
	mux := http.NewServeMux()
	RegisterGardenRoutes(
		mux, capture,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
	)
	request := httptest.NewRequest(http.MethodGet, "/v1/garden", nil)
	request.Header.Set("Authorization", "Bearer access")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)
	if response.Code != http.StatusOK || capture.memberID != "member-1" {
		t.Fatalf("status=%d member=%q body=%s", response.Code, capture.memberID, response.Body.String())
	}
}
