// Package application runs export, deletion and legal-hold workflows
// (E03-S10). Deletion completes with cryptographic erasure of voice and
// biometric blobs, account closure and session revocation; legal holds
// block destruction while preserving review paths.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/stanleyHayes/obiara/internal/privacy/domain"
)

var (
	ErrRequestNotFound   = errors.New("privacy request not found")
	ErrOpenRequestExists = errors.New("an open request of this kind already exists")
	// ErrHoldNotFound reports no active legal hold for an account.
	ErrHoldNotFound = errors.New("no active legal hold")
)

// RequestRepository persists privacy requests.
type RequestRepository interface {
	Create(context.Context, domain.PrivacyRequest) error
	FindByID(context.Context, string) (domain.PrivacyRequest, error)
	FindOpenByAccountAndKind(context.Context, string, domain.Kind) (domain.PrivacyRequest, error)
	Update(context.Context, domain.PrivacyRequest) error
	// NextExecutable returns open, unblocked requests past which execution
	// may begin, oldest first.
	NextExecutable(context.Context, int) ([]domain.PrivacyRequest, error)
}

// LegalHoldRepository persists legal holds.
type LegalHoldRepository interface {
	Place(context.Context, domain.LegalHold) error
	ActiveFor(context.Context, string) (domain.LegalHold, error)
	Lift(context.Context, string, time.Time) error
}

// ExportAssembler collects a member's machine-readable archive (FR-106).
// Implementations gather across contexts; the archive never contains
// another member's data.
type ExportAssembler interface {
	Assemble(ctx context.Context, requestID, accountID string) (archiveRef string, err error)
}

// ErasureRunner cryptographically erases voice/biometric blobs and marks
// the account deleted (FR-106, Doc 09 retention table).
type ErasureRunner interface {
	Erase(ctx context.Context, requestID, accountID string) error
}

// SessionRevoker is the identity-context port for closing sessions on
// deletion.
type SessionRevoker interface {
	RevokeMemberSessions(ctx context.Context, memberID string) error
}

// PrivacyService handles member-facing requests.
type PrivacyService struct {
	requests RequestRepository
	holds    LegalHoldRepository
	now      func() time.Time
	newID    func() string
}

func NewPrivacyService(requests RequestRepository, holds LegalHoldRepository, now func() time.Time, newID func() string) PrivacyService {
	return PrivacyService{requests: requests, holds: holds, now: now, newID: newID}
}

// RequestExport opens an export request (due within 72 hours).
func (service PrivacyService) RequestExport(ctx context.Context, accountID string) (domain.PrivacyRequest, error) {
	return service.open(ctx, accountID, domain.KindExport)
}

// RequestDeletion opens a deletion request (due within 30 days) unless a
// legal hold blocks it.
func (service PrivacyService) RequestDeletion(ctx context.Context, accountID string) (domain.PrivacyRequest, error) {
	if hold, err := service.holds.ActiveFor(ctx, accountID); err == nil && hold.Active() {
		return domain.PrivacyRequest{}, domain.ErrLegalHoldActive
	} else if err != nil && !errors.Is(err, ErrHoldNotFound) {
		return domain.PrivacyRequest{}, err
	}
	return service.open(ctx, accountID, domain.KindDeletion)
}

func (service PrivacyService) open(ctx context.Context, accountID string, kind domain.Kind) (domain.PrivacyRequest, error) {
	if _, err := service.requests.FindOpenByAccountAndKind(ctx, accountID, kind); err == nil {
		return domain.PrivacyRequest{}, ErrOpenRequestExists
	} else if !errors.Is(err, ErrRequestNotFound) {
		return domain.PrivacyRequest{}, err
	}
	request, err := domain.NewRequest(service.newID(), accountID, kind, service.now())
	if err != nil {
		return domain.PrivacyRequest{}, err
	}
	if err := service.requests.Create(ctx, request); err != nil {
		return domain.PrivacyRequest{}, err
	}
	return request, nil
}

// Status returns one request for the member-facing status view.
func (service PrivacyService) Status(ctx context.Context, requestID string) (domain.PrivacyRequest, error) {
	return service.requests.FindByID(ctx, requestID)
}

// PlaceHold preserves an account's data; active holds block deletion.
func (service PrivacyService) PlaceHold(ctx context.Context, accountID, reason, actorID string) error {
	hold, err := domain.NewLegalHold(accountID, reason, actorID, service.now())
	if err != nil {
		return err
	}
	if err := service.holds.Place(ctx, hold); err != nil {
		return err
	}
	// An open deletion request moves behind the hold.
	if request, err := service.requests.FindOpenByAccountAndKind(ctx, accountID, domain.KindDeletion); err == nil {
		_ = request.Block(service.now())
		return service.requests.Update(ctx, request)
	}
	return nil
}

// LiftHold releases the hold and unblocks any held deletion request.
func (service PrivacyService) LiftHold(ctx context.Context, accountID string) error {
	if err := service.holds.Lift(ctx, accountID, service.now()); err != nil {
		return err
	}
	if request, err := service.requests.FindOpenByAccountAndKind(ctx, accountID, domain.KindDeletion); err == nil {
		_ = request.Unblock()
		return service.requests.Update(ctx, request)
	}
	return nil
}

// Processor executes requests (driven by the worker scheduler).
type Processor struct {
	requests  RequestRepository
	assembler ExportAssembler
	erasure   ErasureRunner
	sessions  SessionRevoker
	now       func() time.Time
}

func NewProcessor(requests RequestRepository, assembler ExportAssembler, erasure ErasureRunner, sessions SessionRevoker, now func() time.Time) Processor {
	return Processor{requests: requests, assembler: assembler, erasure: erasure, sessions: sessions, now: now}
}

// RunBatch executes up to limit open requests.
func (processor Processor) RunBatch(ctx context.Context, limit int) error {
	batch, err := processor.requests.NextExecutable(ctx, limit)
	if err != nil {
		return err
	}
	for _, request := range batch {
		if err := processor.execute(ctx, request); err != nil {
			return err
		}
	}
	return nil
}

func (processor Processor) execute(ctx context.Context, request domain.PrivacyRequest) error {
	if request.Status() == domain.StatusRequested {
		if err := request.StartProcessing(); err != nil {
			return err
		}
		if err := processor.requests.Update(ctx, request); err != nil {
			return err
		}
	} else if request.Status() != domain.StatusProcessing {
		return domain.ErrRequestNotOpen
	}

	switch request.Kind() {
	case domain.KindExport:
		if _, err := processor.assembler.Assemble(ctx, request.ID(), request.AccountID()); err != nil {
			return err
		}
	case domain.KindDeletion:
		if err := processor.erasure.Erase(ctx, request.ID(), request.AccountID()); err != nil {
			return err
		}
		if err := processor.sessions.RevokeMemberSessions(ctx, request.AccountID()); err != nil {
			return err
		}
	}

	if err := request.Complete(processor.now()); err != nil {
		return err
	}
	return processor.requests.Update(ctx, request)
}
