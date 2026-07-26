package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
)

var (
	ErrTranscriptionFailed    = errors.New("voice introduction transcription failed")
	ErrTranscriptionUncertain = errors.New("voice introduction transcription uncertain")
)

type Service struct {
	store       Store
	consent     ConsentGate
	media       MediaManager
	transcriber Transcriber
	keyer       Keyer
	ids         IDSource
	now         func() time.Time
}

func NewService(store Store, consent ConsentGate, media MediaManager, transcriber Transcriber, keyer Keyer, ids IDSource, now func() time.Time) Service {
	if now == nil {
		now = time.Now
	}
	return Service{
		store: store, consent: consent, media: media, transcriber: transcriber,
		keyer: keyer, ids: ids, now: now,
	}
}

type BeginUploadRequest struct {
	CommandID      string
	OwnerID        string
	PurposeID      string
	PurposeVersion uint64
	ContentType    string
	RetentionUntil time.Time
	LegalHold      bool
}

type BeginUploadResult struct {
	Introduction domain.Introduction
	Access       UploadAccess
	Replayed     bool
}

func (service Service) BeginUpload(ctx context.Context, request BeginUploadRequest) (BeginUploadResult, error) {
	if !service.ready() {
		return BeginUploadResult{}, ErrDependencyUnavailable
	}
	request.CommandID = strings.TrimSpace(request.CommandID)
	request.OwnerID = strings.TrimSpace(request.OwnerID)
	request.PurposeID = strings.TrimSpace(request.PurposeID)
	request.ContentType = strings.TrimSpace(request.ContentType)
	if err := service.requireConsent(ctx, request.OwnerID, request.PurposeID, request.PurposeVersion); err != nil {
		return BeginUploadResult{}, err
	}
	now := service.now().UTC()
	consent, err := domain.NewConsentSnapshot(request.PurposeID, request.PurposeVersion, now)
	if err != nil {
		return BeginUploadResult{}, err
	}
	assetID := service.ids.NewID("intro_asset")
	media, err := domain.NewMediaRef(assetID, request.ContentType, 0, 0, "")
	if err != nil {
		return BeginUploadResult{}, err
	}
	createCommand, err := service.command(request.CommandID+".create", fmt.Sprintf(
		"%s|%s|%d|%s|%s|%t", request.OwnerID, request.PurposeID,
		request.PurposeVersion, request.ContentType, request.RetentionUntil.UTC().Format(time.RFC3339Nano),
		request.LegalHold,
	), now)
	if err != nil {
		return BeginUploadResult{}, err
	}
	introduction, err := domain.New(
		service.ids.NewID("introduction"), request.OwnerID, consent, media,
		domain.NewRetention(request.RetentionUntil, request.LegalHold), createCommand,
	)
	if err != nil {
		return BeginUploadResult{}, err
	}
	introduction, replayed, err := service.store.Create(ctx, introduction)
	if err != nil {
		if errors.Is(err, ErrCommandAlreadyUsed) {
			return BeginUploadResult{}, domain.ErrCommandConflict
		}
		return BeginUploadResult{}, ErrDependencyUnavailable
	}
	if introduction.OwnerID() != request.OwnerID ||
		introduction.Consent().PurposeID() != request.PurposeID ||
		introduction.Consent().Version() != request.PurposeVersion ||
		!introduction.MatchesCommand(createCommand, domain.ActionCreated) {
		return BeginUploadResult{}, domain.ErrCommandConflict
	}
	if introduction.Status() == domain.StatusDraft {
		authorizeCommand, commandErr := service.command(
			request.CommandID+".upload", introduction.Media().AssetID(), now,
		)
		if commandErr != nil {
			return BeginUploadResult{}, commandErr
		}
		updated, transitionErr := introduction.AuthorizeUpload(authorizeCommand, introduction.Version())
		if transitionErr != nil {
			return BeginUploadResult{}, transitionErr
		}
		if updateErr := service.store.Update(
			ctx, updated, introduction.Version(), authorizeCommand.ID,
		); updateErr != nil {
			current, findErr := service.store.FindByID(ctx, introduction.ID())
			if findErr != nil || current.Status() == domain.StatusDraft {
				return BeginUploadResult{}, ErrDependencyUnavailable
			}
			introduction = current
			replayed = true
		} else {
			introduction = updated
		}
	}
	if introduction.Status() != domain.StatusUploadAuthorized {
		return BeginUploadResult{}, domain.ErrInvalidTransition
	}
	access, err := service.media.AuthorizeUpload(
		ctx, request.OwnerID, introduction.ID(), introduction.Media().AssetID(),
	)
	if err != nil {
		return BeginUploadResult{Introduction: introduction, Replayed: replayed}, ErrDependencyUnavailable
	}
	return BeginUploadResult{Introduction: introduction, Access: access, Replayed: replayed}, nil
}

func (service Service) ConfirmUpload(ctx context.Context, introductionID, commandID string) (domain.Introduction, error) {
	introduction, err := service.find(ctx, introductionID)
	if err != nil {
		return domain.Introduction{}, err
	}
	if err := service.requireSnapshotConsent(ctx, introduction); err != nil {
		return domain.Introduction{}, err
	}
	if introduction.HasCommand(strings.TrimSpace(commandID) + ".confirmed") {
		return introduction, nil
	}
	media, err := service.media.Inspect(ctx, introduction.Media().AssetID())
	if err != nil {
		return domain.Introduction{}, ErrDependencyUnavailable
	}
	command, err := service.command(
		strings.TrimSpace(commandID)+".confirmed",
		fmt.Sprintf("%s|%s|%d|%d|%s", media.AssetID(), media.ContentType(), media.Size(), media.Duration(), media.Checksum()),
		service.now().UTC(),
	)
	if err != nil {
		return domain.Introduction{}, err
	}
	updated, err := introduction.ConfirmUpload(media, command, introduction.Version())
	if err != nil {
		return domain.Introduction{}, err
	}
	return service.update(ctx, introduction, updated, command.ID)
}

type TranscribeResult struct {
	Introduction domain.Introduction
	Replayed     bool
}

func (service Service) Transcribe(ctx context.Context, introductionID, commandID string) (TranscribeResult, error) {
	introduction, err := service.find(ctx, introductionID)
	if err != nil {
		return TranscribeResult{}, err
	}
	if err := service.requireSnapshotConsent(ctx, introduction); err != nil {
		return TranscribeResult{}, err
	}
	commandID = strings.TrimSpace(commandID)
	startID := commandID + ".start"
	if introduction.HasCommand(commandID+".ready") ||
		introduction.HasCommand(commandID+".uncertain") ||
		introduction.HasCommand(commandID+".failed") {
		return transcriptionResult(introduction, true)
	}
	if introduction.Status() != domain.StatusTranscribing {
		start, commandErr := service.command(startID, introduction.Media().AssetID(), service.now().UTC())
		if commandErr != nil {
			return TranscribeResult{}, commandErr
		}
		started, transitionErr := introduction.StartTranscription(start, introduction.Version())
		if transitionErr != nil {
			return TranscribeResult{}, transitionErr
		}
		started, err = service.update(ctx, introduction, started, start.ID)
		if err != nil {
			return TranscribeResult{}, err
		}
		introduction = started
	}
	result, providerErr := service.transcriber.Transcribe(ctx, TranscriptionRequest{
		CommandID: commandID, IntroductionID: introduction.ID(),
		AssetID:        introduction.Media().AssetID(),
		ConsentPurpose: introduction.Consent().PurposeID(),
		ConsentVersion: introduction.Consent().Version(),
	})
	now := service.now().UTC()
	var updated domain.Introduction
	var actionID string
	switch {
	case providerErr != nil:
		actionID = commandID + ".uncertain"
		command, commandErr := service.command(actionID, "provider_error", now)
		if commandErr != nil {
			return TranscribeResult{}, commandErr
		}
		updated, err = introduction.TranscriptionUncertain(command, introduction.Version())
	case result.Outcome == TranscriptionCompleted:
		actionID = commandID + ".ready"
		command, commandErr := service.command(actionID, result.Transcript.ID(), now)
		if commandErr != nil {
			return TranscribeResult{}, commandErr
		}
		updated, err = introduction.CompleteTranscription(result.Transcript, command, introduction.Version())
	case result.Outcome == TranscriptionFailed:
		actionID = commandID + ".failed"
		command, commandErr := service.command(actionID, "provider_failed", now)
		if commandErr != nil {
			return TranscribeResult{}, commandErr
		}
		updated, err = introduction.TranscriptionFailed(command, introduction.Version())
	default:
		actionID = commandID + ".uncertain"
		command, commandErr := service.command(actionID, "provider_uncertain", now)
		if commandErr != nil {
			return TranscribeResult{}, commandErr
		}
		updated, err = introduction.TranscriptionUncertain(command, introduction.Version())
	}
	if err != nil {
		return TranscribeResult{}, err
	}
	updated, err = service.update(ctx, introduction, updated, actionID)
	if err != nil {
		return TranscribeResult{}, err
	}
	return transcriptionResult(updated, false)
}

func (service Service) Cancel(ctx context.Context, introductionID, commandID string) (domain.Introduction, error) {
	return service.stop(ctx, introductionID, commandID, false)
}

func (service Service) Revoke(ctx context.Context, introductionID, commandID string) (domain.Introduction, error) {
	return service.stop(ctx, introductionID, commandID, true)
}

func (service Service) stop(ctx context.Context, introductionID, commandID string, revoked bool) (domain.Introduction, error) {
	introduction, err := service.find(ctx, introductionID)
	if err != nil {
		return domain.Introduction{}, err
	}
	action := "cancel"
	if revoked {
		action = "revoke"
	}
	command, err := service.command(strings.TrimSpace(commandID)+"."+action, action, service.now().UTC())
	if err != nil {
		return domain.Introduction{}, err
	}
	var updated domain.Introduction
	if revoked {
		updated, err = introduction.Revoke(command, introduction.Version())
	} else {
		updated, err = introduction.Cancel(command, introduction.Version())
	}
	if err != nil {
		if introduction.HasCommand(command.ID) {
			return introduction, nil
		}
		return domain.Introduction{}, err
	}
	updated, err = service.update(ctx, introduction, updated, command.ID)
	if err != nil {
		return domain.Introduction{}, err
	}
	if introduction.Status() == domain.StatusTranscribing {
		if cancelErr := service.transcriber.Cancel(ctx, introduction.ID()); cancelErr != nil {
			return updated, ErrDependencyUnavailable
		}
	}
	if !updated.Retention().LegalHold() && !service.now().UTC().Before(updated.DeletionDueAt()) {
		return service.Purge(ctx, updated.ID(), commandID+".cleanup")
	}
	return updated, nil
}

func (service Service) Purge(ctx context.Context, introductionID, commandID string) (domain.Introduction, error) {
	introduction, err := service.find(ctx, introductionID)
	if err != nil {
		return domain.Introduction{}, err
	}
	purgeCommandID := strings.TrimSpace(commandID) + ".purged"
	if introduction.HasCommand(purgeCommandID) && introduction.DataStatus() == domain.DataPurged {
		return introduction, nil
	}
	if introduction.DataStatus() != domain.DataPurgePending {
		return domain.Introduction{}, domain.ErrInvalidTransition
	}
	if introduction.Retention().LegalHold() {
		return domain.Introduction{}, domain.ErrLegalHold
	}
	if service.now().UTC().Before(introduction.DeletionDueAt()) {
		return domain.Introduction{}, domain.ErrRetentionActive
	}
	if err := service.media.Delete(ctx, introduction.Media().AssetID()); err != nil {
		return domain.Introduction{}, ErrDependencyUnavailable
	}
	if introduction.Transcript().ID() != "" {
		if err := service.transcriber.Delete(ctx, introduction.Transcript().ID()); err != nil {
			return domain.Introduction{}, ErrDependencyUnavailable
		}
	}
	command, err := service.command(purgeCommandID, "purged", service.now().UTC())
	if err != nil {
		return domain.Introduction{}, err
	}
	updated, err := introduction.MarkPurged(command, introduction.Version())
	if err != nil {
		return domain.Introduction{}, err
	}
	return service.update(ctx, introduction, updated, command.ID)
}

func (service Service) ready() bool {
	return service.store != nil && service.consent != nil && service.media != nil &&
		service.transcriber != nil && service.keyer != nil && service.ids != nil && service.now != nil
}

func (service Service) requireConsent(ctx context.Context, ownerID, purposeID string, version uint64) error {
	effective, err := service.consent.Effective(ctx, ownerID, purposeID, version)
	if err != nil {
		return ErrDependencyUnavailable
	}
	if !effective {
		return ErrConsentRequired
	}
	return nil
}

func (service Service) requireSnapshotConsent(ctx context.Context, introduction domain.Introduction) error {
	return service.requireConsent(
		ctx, introduction.OwnerID(), introduction.Consent().PurposeID(),
		introduction.Consent().Version(),
	)
}

func (service Service) find(ctx context.Context, id string) (domain.Introduction, error) {
	if !service.ready() {
		return domain.Introduction{}, ErrDependencyUnavailable
	}
	introduction, err := service.store.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return domain.Introduction{}, ErrNotFound
		}
		return domain.Introduction{}, ErrDependencyUnavailable
	}
	return introduction, nil
}

func (service Service) update(ctx context.Context, before, after domain.Introduction, commandID string) (domain.Introduction, error) {
	err := service.store.Update(ctx, after, before.Version(), commandID)
	if err == nil {
		return after, nil
	}
	if errors.Is(err, ErrOptimisticConflict) || errors.Is(err, ErrCommandAlreadyUsed) {
		current, findErr := service.store.FindByID(ctx, before.ID())
		if findErr == nil && current.SameCommand(after, commandID) {
			return current, nil
		}
		if findErr == nil && current.HasCommand(commandID) {
			return domain.Introduction{}, domain.ErrCommandConflict
		}
	}
	return domain.Introduction{}, ErrDependencyUnavailable
}

func (service Service) command(id, payload string, at time.Time) (domain.Command, error) {
	fingerprint, err := service.keyer.Key(payload)
	if err != nil {
		return domain.Command{}, ErrDependencyUnavailable
	}
	return domain.Command{ID: strings.TrimSpace(id), Fingerprint: fingerprint, At: at.UTC()}, nil
}

func transcriptionResult(introduction domain.Introduction, replayed bool) (TranscribeResult, error) {
	result := TranscribeResult{Introduction: introduction, Replayed: replayed}
	switch introduction.Status() {
	case domain.StatusReady:
		return result, nil
	case domain.StatusTranscriptionFailed:
		return result, ErrTranscriptionFailed
	default:
		return result, ErrTranscriptionUncertain
	}
}
