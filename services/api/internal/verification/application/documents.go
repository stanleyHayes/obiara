package application

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

// MaxCardImageBytes bounds one side of a Ghana Card. A phone photograph of a
// card is well under this; anything larger is a mistake or an attempt to fill
// the store, and is refused before it is decoded.
const MaxCardImageBytes = 4 << 20

var (
	ErrInvalidDocument  = errors.New("invalid identity document")
	ErrDocumentStore    = errors.New("identity document store unavailable")
	ErrDocumentNotFound = errors.New("identity document not found")
)

// Document is one encrypted side of a submitted card.
//
// The bytes never rest in the clear and the subject is stored keyed, so a
// dump of this collection identifies nobody on its own. Unlike the liveness
// captures next to it these do not expire on a timer: a reviewer may not get
// to the queue for days, and a document that vanished mid-review would leave
// a member permanently unverifiable with no way to tell why.
type Document struct {
	ID         string
	CaseID     string
	SubjectKey string
	Side       string
	MediaType  string
	Ciphertext []byte
	Nonce      []byte
	CreatedAt  time.Time
}

// DocumentRepository persists encrypted card images.
type DocumentRepository interface {
	SaveDocument(context.Context, Document) error
	DocumentsForCase(ctx context.Context, caseID string) ([]Document, error)
}

// DocumentSealer encrypts a card image at the application boundary.
type DocumentSealer interface {
	Seal([]byte) (ciphertext []byte, nonce []byte, err error)
}

// DocumentOpener decrypts one for a reviewer who is allowed to see it.
type DocumentOpener interface {
	Open(ciphertext, nonce []byte) ([]byte, error)
}

type SubmitDocumentsRequest struct {
	AccountID      string
	CardNumber     string
	DateOfBirth    time.Time
	FrontMediaType string
	FrontBase64    string
	BackMediaType  string
	BackBase64     string
}

type SubmitDocumentsResult struct {
	CaseID string
	Status string
}

// SubmitDocuments opens a card case that goes straight to a human.
//
// This is the whole identity path now. The automated issuer check used to run
// inside signing up, so while the provider was unreachable nobody could get an
// account at all. A member uploads both sides after they are already in, a
// reviewer looks at them, and the outcome decides a badge rather than a door.
// No provider is called here: there is nothing to be uncertain about, so the
// case is queued for review the moment it is created.
func (service VerificationService) SubmitDocuments(
	ctx context.Context,
	request SubmitDocumentsRequest,
) (SubmitDocumentsResult, error) {
	if service.documents == nil || service.sealer == nil {
		return SubmitDocumentsResult{}, ErrDocumentStore
	}
	accountID := strings.TrimSpace(request.AccountID)
	cardNumber := strings.TrimSpace(request.CardNumber)
	if accountID == "" || cardNumber == "" || len(cardNumber) > 32 {
		return SubmitDocumentsResult{}, ErrInvalidDocument
	}
	front, err := decodeCardImage(request.FrontBase64, request.FrontMediaType)
	if err != nil {
		return SubmitDocumentsResult{}, err
	}
	back, err := decodeCardImage(request.BackBase64, request.BackMediaType)
	if err != nil {
		return SubmitDocumentsResult{}, err
	}

	cardKey, err := service.keyer.Key(cardNumber)
	if err != nil {
		return SubmitDocumentsResult{}, ErrInvalidDocument
	}
	// Same gate as the plain card path, for the same reason: this is where a
	// real date of birth arrives, so this is where it is checked — before the
	// case, and before two photographs of the card are sealed and stored.
	caseID := service.newID()
	if err := service.assessAge(ctx, accountID, caseID, request.DateOfBirth); err != nil {
		return SubmitDocumentsResult{}, err
	}

	now := service.now()
	verificationCase, err := domain.NewCase(
		caseID, accountID, cardKey, maskCard(cardNumber), request.DateOfBirth, now,
	)
	if err != nil {
		return SubmitDocumentsResult{}, ErrInvalidDocument
	}
	if err := verificationCase.QueueForManualReview("documents_submitted", now); err != nil {
		return SubmitDocumentsResult{}, ErrInvalidDocument
	}
	if err := service.cases.Create(ctx, verificationCase); err != nil {
		return SubmitDocumentsResult{}, err
	}

	subjectKey, err := service.keyer.Key("subject:" + accountID)
	if err != nil {
		return SubmitDocumentsResult{}, ErrInvalidDocument
	}
	for _, side := range []struct {
		name      string
		mediaType string
		bytes     []byte
	}{
		{"front", request.FrontMediaType, front},
		{"back", request.BackMediaType, back},
	} {
		ciphertext, nonce, sealErr := service.sealer.Seal(side.bytes)
		if sealErr != nil {
			return SubmitDocumentsResult{}, ErrDocumentStore
		}
		if saveErr := service.documents.SaveDocument(ctx, Document{
			ID: service.newID(), CaseID: verificationCase.ID(), SubjectKey: subjectKey,
			Side: side.name, MediaType: strings.ToLower(strings.TrimSpace(side.mediaType)),
			Ciphertext: ciphertext, Nonce: nonce, CreatedAt: now,
		}); saveErr != nil {
			return SubmitDocumentsResult{}, ErrDocumentStore
		}
	}

	return SubmitDocumentsResult{
		CaseID: verificationCase.ID(),
		Status: string(verificationCase.Status()),
	}, nil
}

// CardImage is one decrypted side, handed to a reviewer.
type CardImage struct {
	Side      string
	MediaType string
	Bytes     []byte
}

// OpenDocuments decrypts a case's images for review.
//
// Authorisation is the caller's job — this is reached only from the admin
// surface, behind the verification-review capability and a stepped-up session.
func (service VerificationService) OpenDocuments(ctx context.Context, caseID string) ([]CardImage, error) {
	if service.documents == nil || service.opener == nil {
		return nil, ErrDocumentStore
	}
	stored, err := service.documents.DocumentsForCase(ctx, strings.TrimSpace(caseID))
	if err != nil {
		return nil, err
	}
	if len(stored) == 0 {
		return nil, ErrDocumentNotFound
	}
	images := make([]CardImage, 0, len(stored))
	for _, document := range stored {
		plaintext, openErr := service.opener.Open(document.Ciphertext, document.Nonce)
		if openErr != nil {
			return nil, ErrDocumentStore
		}
		images = append(images, CardImage{
			Side: document.Side, MediaType: document.MediaType, Bytes: plaintext,
		})
	}
	return images, nil
}

// decodeCardImage accepts only the photograph formats a phone produces.
func decodeCardImage(encoded, mediaType string) ([]byte, error) {
	normalised := strings.ToLower(strings.TrimSpace(mediaType))
	if normalised != "image/jpeg" && normalised != "image/png" && normalised != "image/webp" {
		return nil, ErrInvalidDocument
	}
	value, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil || len(value) == 0 || len(value) > MaxCardImageBytes {
		return nil, ErrInvalidDocument
	}
	return value, nil
}
