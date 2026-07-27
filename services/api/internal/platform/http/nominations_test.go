package apihttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/application"
	"github.com/stanleyHayes/obiara/services/api/internal/companions/nnoboa/domain"
)

type fakeNominations struct {
	nominate func(ctx context.Context, in application.NominateInput) (domain.Nomination, error)
	list     func(ctx context.Context, memberID string) ([]domain.Nomination, error)
	consent  func(ctx context.Context, id string) (domain.Nomination, error)
	decline  func(ctx context.Context, id string) (domain.Nomination, error)
}

func (f fakeNominations) Nominate(ctx context.Context, in application.NominateInput) (domain.Nomination, error) {
	return f.nominate(ctx, in)
}

func (f fakeNominations) ListForMember(ctx context.Context, memberID string) ([]domain.Nomination, error) {
	return f.list(ctx, memberID)
}

func (f fakeNominations) Consent(ctx context.Context, id string) (domain.Nomination, error) {
	return f.consent(ctx, id)
}

func (f fakeNominations) Decline(ctx context.Context, id string) (domain.Nomination, error) {
	return f.decline(ctx, id)
}

func pendingNomination(t *testing.T) domain.Nomination {
	t.Helper()
	n, err := domain.NewNomination("mem_12345678", "Auntie Efua", "+233550000101", "aunt",
		time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func nominationMux(n Nominations) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterNominationRoutes(mux, n)
	return mux
}

func TestNominateSuccess(t *testing.T) {
	n := pendingNomination(t)
	fake := fakeNominations{
		nominate: func(_ context.Context, in application.NominateInput) (domain.Nomination, error) {
			if in.MemberID != "mem_12345678" || in.Relationship != "aunt" {
				t.Errorf("input = %+v", in)
			}
			return n, nil
		},
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/nominations",
		strings.NewReader(`{"memberId":"mem_12345678","kinName":"Auntie Efua","kinPhone":"+233550000101","relationship":"aunt"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	nominationMux(fake).ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data nominationResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Status != "pending" || envelope.Data.KinName != "Auntie Efua" {
		t.Fatalf("data = %+v", envelope.Data)
	}
}

func TestNominateValidationFailure(t *testing.T) {
	fake := fakeNominations{
		nominate: func(context.Context, application.NominateInput) (domain.Nomination, error) {
			t.Error("service must not be called on invalid input")
			return domain.Nomination{}, nil
		},
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/nominations",
		strings.NewReader(`{"memberId":"mem_12345678","kinName":"Auntie Efua","kinPhone":"0550000101","relationship":"cousin"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	nominationMux(fake).ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
}

func TestNominateDuplicateConflict(t *testing.T) {
	fake := fakeNominations{
		nominate: func(context.Context, application.NominateInput) (domain.Nomination, error) {
			return domain.Nomination{}, application.ErrDuplicateNomination
		},
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/nominations",
		strings.NewReader(`{"memberId":"mem_12345678","kinName":"Auntie Efua","kinPhone":"+233550000101","relationship":"aunt"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	nominationMux(fake).ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
}

func TestListNominations(t *testing.T) {
	n := pendingNomination(t)
	fake := fakeNominations{
		list: func(_ context.Context, memberID string) ([]domain.Nomination, error) {
			if memberID != "mem_12345678" {
				t.Errorf("memberID = %q", memberID)
			}
			return []domain.Nomination{n}, nil
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/nominations?memberId=mem_12345678", nil)
	response := httptest.NewRecorder()
	nominationMux(fake).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data nominationListResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if len(envelope.Data.Nominations) != 1 {
		t.Fatalf("nominations = %+v", envelope.Data.Nominations)
	}
}

func TestListNominationsRequiresMemberID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/v1/nominations", nil)
	response := httptest.NewRecorder()
	nominationMux(fakeNominations{}).ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
}

func TestConsentNominationNotFound(t *testing.T) {
	fake := fakeNominations{
		consent: func(context.Context, string) (domain.Nomination, error) {
			return domain.Nomination{}, application.ErrNominationNotFound
		},
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/nominations/nom_missing/consent", nil)
	response := httptest.NewRecorder()
	nominationMux(fake).ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
}

func TestDeclineNominationSuccess(t *testing.T) {
	n := pendingNomination(t)
	if err := n.Decline(time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	fake := fakeNominations{
		decline: func(_ context.Context, id string) (domain.Nomination, error) {
			if id != n.ID {
				t.Errorf("id = %q", id)
			}
			return n, nil
		},
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/nominations/"+n.ID+"/decline", nil)
	response := httptest.NewRecorder()
	nominationMux(fake).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data nominationResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Status != "declined" || envelope.Data.RespondedAt == nil {
		t.Fatalf("data = %+v", envelope.Data)
	}
}
