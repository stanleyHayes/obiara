package apihttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/liveness/application"
)

type Liveness interface {
	Submit(context.Context, application.SubmitRequest) (application.Result, error)
}

type LivenessArtifacts interface {
	Upload(context.Context, application.ArtifactUploadRequest) (application.ArtifactUploadResult, error)
}

func RegisterLivenessRoutes(
	mux *http.ServeMux,
	service Liveness,
	artifacts LivenessArtifacts,
	sessions SessionAuthenticator,
) {
	mux.Handle("POST /v1/verifications/liveness", livenessHandler(service, sessions))
	mux.Handle("POST /v1/verifications/liveness/artifacts", livenessArtifactHandler(artifacts, sessions))
}

type livenessRequest struct {
	VoiceArtifactRef string `json:"voiceArtifactRef"`
	FaceArtifactRef  string `json:"faceArtifactRef"`
}

type livenessResponse struct {
	AttemptID string `json:"attemptId"`
	Status    string `json:"status"`
	Replayed  bool   `json:"replayed"`
}

type livenessArtifactRequest struct {
	VoiceMediaType string `json:"voiceMediaType"`
	VoiceBase64    string `json:"voiceBase64"`
	FaceMediaType  string `json:"faceMediaType"`
	FaceBase64     string `json:"faceBase64"`
}

type livenessArtifactResponse struct {
	VoiceArtifactRef string `json:"voiceArtifactRef"`
	FaceArtifactRef  string `json:"faceArtifactRef"`
	ExpiresAt        string `json:"expiresAt"`
}

const maxLivenessArtifactRequestBytes = 5 << 20

func livenessArtifactHandler(artifacts LivenessArtifacts, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok || artifacts == nil || sessions == nil {
			writeError(writer, request, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		session, err := sessions.Authenticate(request.Context(), token)
		if err != nil {
			writeError(writer, request, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		if mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(writer, request, http.StatusUnsupportedMediaType, APIError{
				Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
			})
			return
		}
		request.Body = http.MaxBytesReader(writer, request.Body, maxLivenessArtifactRequestBytes)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		var body livenessArtifactRequest
		if err := decoder.Decode(&body); err != nil {
			writeError(writer, request, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			writeError(writer, request, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must contain one JSON object.",
			})
			return
		}
		result, err := artifacts.Upload(request.Context(), application.ArtifactUploadRequest{
			SubjectID: session.MemberID(), VoiceMediaType: body.VoiceMediaType,
			VoiceBase64: body.VoiceBase64, FaceMediaType: body.FaceMediaType,
			FaceBase64: body.FaceBase64,
		})
		if errors.Is(err, application.ErrInvalidArtifact) {
			writeError(writer, request, http.StatusUnprocessableEntity, APIError{
				Code: "validation_failed", Message: "Audio or face capture is missing, unsupported, or too large.",
			})
			return
		}
		if err != nil {
			writeError(writer, request, http.StatusServiceUnavailable, APIError{
				Code: "artifact_store_unavailable", Message: "Temporary secure storage is unavailable.",
			})
			return
		}
		writeSuccess(writer, request, http.StatusCreated, livenessArtifactResponse{
			VoiceArtifactRef: result.VoiceArtifactRef,
			FaceArtifactRef:  result.FaceArtifactRef,
			ExpiresAt:        result.ExpiresAt.Format(time.RFC3339),
		})
	})
}

func livenessHandler(service Liveness, sessions SessionAuthenticator) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok || service == nil || sessions == nil {
			writeError(writer, request, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		session, err := sessions.Authenticate(request.Context(), token)
		if err != nil {
			writeError(writer, request, http.StatusUnauthorized, APIError{
				Code: "authentication_required", Message: "A valid member session is required.",
			})
			return
		}
		if mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type")); err != nil || mediaType != "application/json" {
			writeError(writer, request, http.StatusUnsupportedMediaType, APIError{
				Code: "unsupported_media_type", Message: "Content-Type must be application/json.",
			})
			return
		}
		commandID := strings.TrimSpace(request.Header.Get("Idempotency-Key"))
		var body livenessRequest
		if decodeErr := decodeJSON(writer, request, &body); decodeErr != nil {
			writeError(writer, request, http.StatusBadRequest, APIError{
				Code: "invalid_json", Message: "The request body must be one valid JSON object.",
			})
			return
		}
		body.VoiceArtifactRef = strings.TrimSpace(body.VoiceArtifactRef)
		body.FaceArtifactRef = strings.TrimSpace(body.FaceArtifactRef)
		if !validIdentifier(commandID) || !validIdentifier(body.VoiceArtifactRef) ||
			!validIdentifier(body.FaceArtifactRef) {
			writeError(writer, request, http.StatusUnprocessableEntity, APIError{
				Code:    "validation_failed",
				Message: "Valid artifact references and Idempotency-Key are required.",
			})
			return
		}
		result, err := service.Submit(request.Context(), application.SubmitRequest{
			CommandID: commandID, SubjectID: session.MemberID(),
			VoiceArtifactRef: body.VoiceArtifactRef, FaceArtifactRef: body.FaceArtifactRef,
		})
		status := http.StatusCreated
		if errors.Is(err, application.ErrManualReviewNeeded) {
			status = http.StatusAccepted
		} else if errors.Is(err, application.ErrLivenessFailed) {
			writeError(writer, request, http.StatusUnprocessableEntity, APIError{
				Code: "liveness_failed", Message: "The liveness check could not be confirmed.",
			})
			return
		} else if errors.Is(err, application.ErrCommandConflict) ||
			errors.Is(err, application.ErrOptimisticConflict) {
			writeError(writer, request, http.StatusConflict, APIError{
				Code: "liveness_conflict", Message: "The liveness request conflicts with an earlier attempt.",
			})
			return
		} else if err != nil {
			writeError(writer, request, http.StatusServiceUnavailable, APIError{
				Code: "liveness_unavailable", Message: "The liveness service is temporarily unavailable.",
			})
			return
		}
		writeSuccess(writer, request, status, livenessResponse{
			AttemptID: result.Attempt.ID(), Status: string(result.Attempt.Status()),
			Replayed: result.Replayed,
		})
	})
}
