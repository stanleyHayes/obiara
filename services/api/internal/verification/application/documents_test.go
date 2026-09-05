package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/verification/domain"
)

type memoryDocuments struct {
	saved []Document
}

func (store *memoryDocuments) SaveDocument(_ context.Context, document Document) error {
	store.saved = append(store.saved, document)
	return nil
}

func (store *memoryDocuments) DocumentsForCase(_ context.Context, caseID string) ([]Document, error) {
	var found []Document
	for _, document := range store.saved {
		if document.CaseID == caseID {
			found = append(found, document)
		}
	}
	return found, nil
}

// plainSealer stands in for AES-GCM so the test asserts the flow, not crypto.
type plainSealer struct{}

func (plainSealer) Seal(plaintext []byte) ([]byte, []byte, error) {
	return plaintext, []byte("nonce"), nil
}

func (plainSealer) Open(ciphertext, _ []byte) ([]byte, error) { return ciphertext, nil }

type staticKeyer struct{}

func (staticKeyer) Key(value string) (string, error) { return "key:" + value, nil }

type recordingCases struct {
	created []domain.VerificationCase
}

func (store *recordingCases) Create(_ context.Context, value domain.VerificationCase) error {
	store.created = append(store.created, value)
	return nil
}
func (store *recordingCases) FindByID(context.Context, string) (domain.VerificationCase, error) {
	return domain.VerificationCase{}, ErrCaseNotFound
}
func (store *recordingCases) Update(context.Context, domain.VerificationCase) error { return nil }
func (store *recordingCases) NextQueued(context.Context, int) ([]domain.VerificationCase, error) {
	return nil, nil
}
func (store *recordingCases) ApprovedAccountByCardKey(context.Context, string) (string, error) {
	return "", ErrCaseNotFound
}
func (store *recordingCases) LatestByAccount(context.Context, string) (domain.VerificationCase, error) {
	return domain.VerificationCase{}, ErrCaseNotFound
}

// refusingProvider fails if it is ever called. The document path must not
// consult the issuer: that call is exactly the third party whose outage used
// to stop anyone from signing up.
type refusingProvider struct{ t *testing.T }

func (provider refusingProvider) Verify(context.Context, ProviderRequest) (ProviderResult, error) {
	provider.t.Fatal("the document path must not call the identity provider")
	return ProviderResult{}, nil
}

func documentService(t *testing.T, cases *recordingCases, documents *memoryDocuments) VerificationService {
	t.Helper()
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	counter := 0
	return NewVerificationService(
		cases, refusingProvider{t: t}, nil, staticKeyer{}, adultAgeGate{},
		func() time.Time { return now },
		func() string { counter++; return "vc_QWx0Z2ViZXJ0X21lbWJlcg" },
	).WithDocuments(documents, plainSealer{}, plainSealer{})
}

func validRequest() SubmitDocumentsRequest {
	// "AAAA" base64-decodes to three bytes; enough to be a non-empty image.
	return SubmitDocumentsRequest{
		AccountID: "member-1", CardNumber: "GHA-000000000-0",
		DateOfBirth:    time.Date(1996, time.March, 2, 0, 0, 0, 0, time.UTC),
		FrontMediaType: "image/jpeg", FrontBase64: "AAAA",
		BackMediaType: "image/png", BackBase64: "AAAA",
	}
}

func TestSubmitDocumentsQueuesForReviewWithoutTheProvider(t *testing.T) {
	cases := &recordingCases{}
	documents := &memoryDocuments{}
	service := documentService(t, cases, documents)

	result, err := service.SubmitDocuments(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("SubmitDocuments = %v", err)
	}
	if result.Status != string(domain.StatusQueuedManual) {
		t.Fatalf("status = %q, want queued_manual", result.Status)
	}
	if len(documents.saved) != 2 {
		t.Fatalf("stored %d sides, want 2", len(documents.saved))
	}
	sides := map[string]bool{}
	for _, document := range documents.saved {
		sides[document.Side] = true
		// The raw account id must never be the key a document is filed under.
		if strings.Contains(document.SubjectKey, "member-1") &&
			!strings.HasPrefix(document.SubjectKey, "key:") {
			t.Fatalf("subject stored unkeyed: %q", document.SubjectKey)
		}
	}
	if !sides["front"] || !sides["back"] {
		t.Fatalf("sides = %v, want both", sides)
	}
}

func TestSubmitDocumentsRefusesUnusableImages(t *testing.T) {
	for name, mutate := range map[string]func(*SubmitDocumentsRequest){
		"a PDF is not a photograph":   func(r *SubmitDocumentsRequest) { r.FrontMediaType = "application/pdf" },
		"an empty side":               func(r *SubmitDocumentsRequest) { r.BackBase64 = "" },
		"not base64 at all":           func(r *SubmitDocumentsRequest) { r.FrontBase64 = "!!!!" },
		"a card number nobody typed":  func(r *SubmitDocumentsRequest) { r.CardNumber = "  " },
		"a card number that is a nov": func(r *SubmitDocumentsRequest) { r.CardNumber = strings.Repeat("9", 33) },
	} {
		t.Run(name, func(t *testing.T) {
			cases := &recordingCases{}
			documents := &memoryDocuments{}
			service := documentService(t, cases, documents)
			request := validRequest()
			mutate(&request)
			if _, err := service.SubmitDocuments(context.Background(), request); err == nil {
				t.Fatal("expected a refusal")
			}
			// Nothing half-written: no case opened, no image stored.
			if len(cases.created) != 0 || len(documents.saved) != 0 {
				t.Fatalf("partial write: cases=%d documents=%d", len(cases.created), len(documents.saved))
			}
		})
	}
}

func TestOpenDocumentsReturnsBothSidesForAReviewer(t *testing.T) {
	cases := &recordingCases{}
	documents := &memoryDocuments{}
	service := documentService(t, cases, documents)
	result, err := service.SubmitDocuments(context.Background(), validRequest())
	if err != nil {
		t.Fatal(err)
	}
	images, err := service.OpenDocuments(context.Background(), result.CaseID)
	if err != nil || len(images) != 2 {
		t.Fatalf("images=%d err=%v", len(images), err)
	}
	if _, err := service.OpenDocuments(context.Background(), "vc_nothing"); err != ErrDocumentNotFound {
		t.Fatalf("missing case = %v, want ErrDocumentNotFound", err)
	}
}
