package application

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

type artifactRepositoryStub struct {
	values []Artifact
	err    error
}

func (stub *artifactRepositoryStub) SaveArtifact(_ context.Context, artifact Artifact) error {
	stub.values = append(stub.values, artifact)
	return stub.err
}

type artifactSealerStub struct{}

func (artifactSealerStub) Seal(value []byte) ([]byte, []byte, error) {
	return append([]byte("sealed:"), value...), []byte("nonce"), nil
}

type artifactKeyerStub struct{}

func (artifactKeyerStub) Key(string) (string, error) { return "subject-key", nil }

type artifactIDs struct{ next int }

func (ids *artifactIDs) NewID() string {
	ids.next++
	return "artifact-" + string(rune('0'+ids.next))
}

func TestArtifactServiceEncryptsAndExpiresCaptures(t *testing.T) {
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	repository := &artifactRepositoryStub{}
	service := NewArtifactService(
		repository, artifactSealerStub{}, artifactKeyerStub{}, &artifactIDs{},
		func() time.Time { return now },
	)
	result, err := service.Upload(context.Background(), ArtifactUploadRequest{
		SubjectID: "member-1", VoiceMediaType: "audio/webm",
		VoiceBase64:   base64.StdEncoding.EncodeToString([]byte("voice")),
		FaceMediaType: "image/jpeg",
		FaceBase64:    base64.StdEncoding.EncodeToString([]byte("face")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.VoiceArtifactRef == result.FaceArtifactRef || len(repository.values) != 2 {
		t.Fatalf("result=%+v saved=%d", result, len(repository.values))
	}
	if string(repository.values[0].Ciphertext) == "voice" ||
		!repository.values[0].ExpiresAt.Equal(now.Add(24*time.Hour)) {
		t.Fatalf("artifact=%+v", repository.values[0])
	}
	if repository.values[0].SubjectKey != "subject-key" {
		t.Fatalf("subject key=%q", repository.values[0].SubjectKey)
	}
}

func TestArtifactServiceRejectsUnsupportedOrOversizedCapture(t *testing.T) {
	service := NewArtifactService(
		&artifactRepositoryStub{}, artifactSealerStub{}, artifactKeyerStub{},
		&artifactIDs{}, time.Now,
	)
	_, err := service.Upload(context.Background(), ArtifactUploadRequest{
		SubjectID: "member-1", VoiceMediaType: "text/plain",
		VoiceBase64:   base64.StdEncoding.EncodeToString([]byte("voice")),
		FaceMediaType: "image/jpeg",
		FaceBase64:    base64.StdEncoding.EncodeToString([]byte("face")),
	})
	if !errors.Is(err, ErrInvalidArtifact) {
		t.Fatalf("error=%v", err)
	}
}
