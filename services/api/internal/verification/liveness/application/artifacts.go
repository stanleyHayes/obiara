package application

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

const (
	MaxVoiceArtifactBytes = 2 << 20
	MaxFaceArtifactBytes  = 1 << 20
)

var (
	ErrInvalidArtifact = errors.New("invalid liveness artifact")
	ErrArtifactStore   = errors.New("liveness artifact store unavailable")
)

type Artifact struct {
	ID         string
	SubjectKey string
	Kind       string
	MediaType  string
	Ciphertext []byte
	Nonce      []byte
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

type ArtifactRepository interface {
	SaveArtifact(context.Context, Artifact) error
}

type ArtifactSealer interface {
	Seal([]byte) (ciphertext []byte, nonce []byte, err error)
}

type ArtifactUploadRequest struct {
	SubjectID      string
	VoiceMediaType string
	VoiceBase64    string
	FaceMediaType  string
	FaceBase64     string
}

type ArtifactUploadResult struct {
	VoiceArtifactRef string
	FaceArtifactRef  string
	ExpiresAt        time.Time
}

type ArtifactService struct {
	repository ArtifactRepository
	sealer     ArtifactSealer
	keyer      Keyer
	ids        IDSource
	now        func() time.Time
}

func NewArtifactService(
	repository ArtifactRepository,
	sealer ArtifactSealer,
	keyer Keyer,
	ids IDSource,
	now func() time.Time,
) ArtifactService {
	return ArtifactService{repository: repository, sealer: sealer, keyer: keyer, ids: ids, now: now}
}

func (service ArtifactService) Upload(ctx context.Context, request ArtifactUploadRequest) (ArtifactUploadResult, error) {
	if service.repository == nil || service.sealer == nil || service.keyer == nil ||
		service.ids == nil || service.now == nil || strings.TrimSpace(request.SubjectID) == "" {
		return ArtifactUploadResult{}, ErrInvalidArtifact
	}
	voice, err := decodeArtifact(request.VoiceBase64, request.VoiceMediaType, "audio/", MaxVoiceArtifactBytes)
	if err != nil {
		return ArtifactUploadResult{}, err
	}
	face, err := decodeArtifact(request.FaceBase64, request.FaceMediaType, "image/jpeg", MaxFaceArtifactBytes)
	if err != nil {
		return ArtifactUploadResult{}, err
	}
	subjectKey, err := service.keyer.Key("subject:" + request.SubjectID)
	if err != nil {
		return ArtifactUploadResult{}, ErrInvalidArtifact
	}
	now := service.now().UTC()
	expiresAt := now.Add(24 * time.Hour)
	voiceRef, err := service.store(ctx, subjectKey, "voice", request.VoiceMediaType, voice, now, expiresAt)
	if err != nil {
		return ArtifactUploadResult{}, err
	}
	faceRef, err := service.store(ctx, subjectKey, "face", request.FaceMediaType, face, now, expiresAt)
	if err != nil {
		return ArtifactUploadResult{}, err
	}
	return ArtifactUploadResult{
		VoiceArtifactRef: voiceRef,
		FaceArtifactRef:  faceRef,
		ExpiresAt:        expiresAt,
	}, nil
}

func decodeArtifact(encoded, mediaType, allowedType string, maxBytes int) ([]byte, error) {
	mediaType = strings.TrimSpace(strings.ToLower(mediaType))
	if mediaType == "" ||
		(allowedType == "audio/" && !strings.HasPrefix(mediaType, allowedType)) ||
		(allowedType != "audio/" && mediaType != allowedType) {
		return nil, ErrInvalidArtifact
	}
	value, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(value) == 0 || len(value) > maxBytes {
		return nil, ErrInvalidArtifact
	}
	return value, nil
}

func (service ArtifactService) store(
	ctx context.Context,
	subjectKey, kind, mediaType string,
	plaintext []byte,
	createdAt, expiresAt time.Time,
) (string, error) {
	ciphertext, nonce, err := service.sealer.Seal(plaintext)
	if err != nil {
		return "", ErrArtifactStore
	}
	id := service.ids.NewID()
	if err := service.repository.SaveArtifact(ctx, Artifact{
		ID: id, SubjectKey: subjectKey, Kind: kind, MediaType: mediaType,
		Ciphertext: ciphertext, Nonce: nonce, CreatedAt: createdAt, ExpiresAt: expiresAt,
	}); err != nil {
		return "", ErrArtifactStore
	}
	return id, nil
}
