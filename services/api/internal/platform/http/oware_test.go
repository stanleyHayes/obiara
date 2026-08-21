package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/application"
	sessiondomain "github.com/stanleyHayes/obiara/services/api/internal/games/oware/session/domain"
	identitydomain "github.com/stanleyHayes/obiara/services/api/internal/identity/domain"
)

type owareCapture struct {
	command application.Command
	second  string
}

func (capture *owareCapture) Create(
	_ context.Context,
	command application.Command,
	second string,
	_ time.Duration,
) (sessiondomain.Projection, error) {
	capture.command, capture.second = command, second
	return sessiondomain.Projection{
		ID: "oware-1", RoomRef: strings.Repeat("a", 64),
		Players: []string{strings.Repeat("b", 64), strings.Repeat("c", 64)},
		Houses:  [12]int{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
		Status:  sessiondomain.StatusActive, Revision: 1,
		MoveDeadline: time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC),
		ServerTime:   time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC),
	}, nil
}

func (*owareCapture) Move(context.Context, application.Command, int) (sessiondomain.Projection, error) {
	return sessiondomain.Projection{}, nil
}

func (*owareCapture) View(context.Context, application.Command) (sessiondomain.Projection, error) {
	return sessiondomain.Projection{}, nil
}

type pairCapture struct{ actor string }

func (capture *pairCapture) Pair(_ context.Context, _, actor string) (string, error) {
	capture.actor = actor
	return "member-2", nil
}

func TestOwareCreateDerivesPlayerFromSessionAndRedactsKeys(t *testing.T) {
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	games, pairs := &owareCapture{}, &pairCapture{}
	mux := http.NewServeMux()
	RegisterOwareRoutes(
		mux, games, pairs,
		sessionAuthenticatorStub{authenticate: func(context.Context, string) (identitydomain.Session, error) {
			return memberSession(now), nil
		}},
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/circles/circle-1/oware",
		strings.NewReader(`{}`),
	)
	request.Header.Set("Authorization", "Bearer access")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "oware-create-1")
	response := httptest.NewRecorder()
	Correlation(mux).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if pairs.actor != "member-1" || games.command.ActorID != "member-1" ||
		games.command.RoomID != "circle-1" || games.second != "member-2" {
		t.Fatalf("pair actor=%q command=%+v second=%q", pairs.actor, games.command, games.second)
	}
	body := response.Body.String()
	for _, forbidden := range []string{"roomRef", "players", strings.Repeat("a", 64), strings.Repeat("b", 64)} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response leaked %q: %s", forbidden, body)
		}
	}
}
