package apihttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/profile/application"
	"github.com/stanleyHayes/obiara/services/api/internal/profile/domain"
)

// Doorway is the inbound port for the doorway question (E03-S09).
type Doorway interface {
	Set(ctx context.Context, memberID, text string, custom bool) (domain.DoorwayQuestion, error)
	Get(ctx context.Context, memberID string) (domain.DoorwayQuestion, error)
}

// Vault is the inbound port for the photo vault (E03-S09).
type Vault interface {
	Add(ctx context.Context, memberID, assetID string, position int) (domain.VaultItem, error)
	ViewFor(ctx context.Context, ownerID, viewerID string) ([]domain.VaultItemView, error)
}

// RegisterDoorwayRoutes adds doorway question and photo vault routes.
func RegisterDoorwayRoutes(mux *http.ServeMux, doorway Doorway, vault Vault) {
	mux.Handle("PUT /v1/doorway-question", setDoorwayQuestionHandler(doorway))
	mux.Handle("GET /v1/doorway-question/{memberId}", getDoorwayQuestionHandler(doorway))
	mux.Handle("POST /v1/photo-vault/items", addVaultItemHandler(vault))
	mux.Handle("GET /v1/photo-vault/{ownerId}", viewVaultHandler(vault))
}

type doorwayQuestionRequest struct {
	MemberID string `json:"memberId"`
	Text     string `json:"text"`
	Custom   bool   `json:"custom"`
}

type doorwayQuestionResponse struct {
	Text      string    `json:"text"`
	Custom    bool      `json:"custom"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func setDoorwayQuestionHandler(doorway Doorway) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body doorwayQuestionRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.MemberID = strings.TrimSpace(body.MemberID)
		if body.MemberID == "" || len(body.MemberID) > maxIdentifierLength || !identifierPattern.MatchString(body.MemberID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}

		question, err := doorway.Set(r.Context(), body.MemberID, body.Text, body.Custom)
		if err != nil {
			writeDoorwayError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, doorwayQuestionResponse{
			Text:      question.Text(),
			Custom:    question.Custom(),
			UpdatedAt: question.UpdatedAt(),
		})
	})
}

func getDoorwayQuestionHandler(doorway Doorway) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		question, err := doorway.Get(r.Context(), r.PathValue("memberId"))
		if err != nil {
			writeDoorwayError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusOK, doorwayQuestionResponse{
			Text:      question.Text(),
			Custom:    question.Custom(),
			UpdatedAt: question.UpdatedAt(),
		})
	})
}

type vaultItemRequest struct {
	MemberID string `json:"memberId"`
	AssetID  string `json:"assetId"`
	Position int    `json:"position"`
}

type vaultItemResponse struct {
	ItemID   string `json:"itemId"`
	Position int    `json:"position"`
}

func addVaultItemHandler(vault Vault) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(w, r, http.StatusUnsupportedMediaType, APIError{
				Code:    "unsupported_media_type",
				Message: "Content-Type must be application/json.",
			})
			return
		}

		var body vaultItemRequest
		if err := decodeJSON(w, r, &body); err != nil {
			writeError(w, r, http.StatusBadRequest, APIError{
				Code:    "invalid_json",
				Message: "The request body must be one valid JSON object.",
			})
			return
		}

		body.MemberID = strings.TrimSpace(body.MemberID)
		body.AssetID = strings.TrimSpace(body.AssetID)
		var details []FieldError
		if body.MemberID == "" || len(body.MemberID) > maxIdentifierLength || !identifierPattern.MatchString(body.MemberID) {
			details = append(details, FieldError{Field: "memberId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if body.AssetID == "" || len(body.AssetID) > maxIdentifierLength || !identifierPattern.MatchString(body.AssetID) {
			details = append(details, FieldError{Field: "assetId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"})
		}
		if body.Position < 0 || body.Position >= domain.VaultCapacity {
			details = append(details, FieldError{Field: "position", Reason: "must be 0-11"})
		}
		if len(details) > 0 {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: details,
			})
			return
		}

		item, err := vault.Add(r.Context(), body.MemberID, body.AssetID, body.Position)
		if err != nil {
			writeDoorwayError(w, r, err)
			return
		}
		writeSuccess(w, r, http.StatusCreated, vaultItemResponse{ItemID: item.ID(), Position: item.Position()})
	})
}

type vaultItemViewJSON struct {
	AssetID  string `json:"assetId"`
	Position int    `json:"position"`
	Veiled   bool   `json:"veiled"`
}

type vaultViewResponse struct {
	Items []vaultItemViewJSON `json:"items"`
}

func viewVaultHandler(vault Vault) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ownerID := r.PathValue("ownerId")
		if ownerID == "" || len(ownerID) > maxIdentifierLength || !identifierPattern.MatchString(ownerID) {
			writeError(w, r, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "One or more fields are invalid.",
				Details: []FieldError{{Field: "ownerId", Reason: "must be 1-128 letters, numbers, dots, underscores, colons, or hyphens"}},
			})
			return
		}
		viewerID := strings.TrimSpace(r.URL.Query().Get("viewerId"))
		views, err := vault.ViewFor(r.Context(), ownerID, viewerID)
		if err != nil {
			writeDoorwayError(w, r, err)
			return
		}
		items := make([]vaultItemViewJSON, 0, len(views))
		for _, view := range views {
			items = append(items, vaultItemViewJSON{AssetID: view.AssetID, Position: view.Position, Veiled: view.Veiled})
		}
		writeSuccess(w, r, http.StatusOK, vaultViewResponse{Items: items})
	})
}

func writeDoorwayError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrDoorwayQuestionInvalid), errors.Is(err, domain.ErrUnsafeProfile):
		writeError(w, r, http.StatusUnprocessableEntity, APIError{
			Code:    "validation_failed",
			Message: "The question must be 1-60 safe characters.",
		})
	case errors.Is(err, application.ErrDoorwayQuestionMissing):
		writeError(w, r, http.StatusNotFound, APIError{
			Code:    "doorway_question_not_found",
			Message: "No doorway question is set.",
		})
	case errors.Is(err, domain.ErrVaultFull):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "vault_full",
			Message: "The photo vault is full.",
		})
	case errors.Is(err, application.ErrVaultItemConflict):
		writeError(w, r, http.StatusConflict, APIError{
			Code:    "vault_position_taken",
			Message: "That vault position is already taken.",
		})
	default:
		writeError(w, r, http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "The request could not be completed.",
		})
	}
}
