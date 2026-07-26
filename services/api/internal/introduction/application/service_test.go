package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/introduction/domain"
	"go.uber.org/mock/gomock"
)

var applicationTime = time.Date(2026, 7, 26, 21, 0, 0, 0, time.UTC)

func digest(character string) string { return strings.Repeat(character, 64) }

func appCommand(id string, at time.Time) domain.Command {
	return domain.Command{ID: id, Fingerprint: digest("a"), At: at}
}

func applicationIntroduction(t *testing.T, status domain.Status, retention domain.Retention) domain.Introduction {
	t.Helper()
	consent, _ := domain.NewConsentSnapshot("voice.introduction", 2, applicationTime)
	media, _ := domain.NewMediaRef("asset:1", "audio/ogg", 0, 0, "")
	introduction, err := domain.New(
		"introduction:1", "member:1", consent, media, retention,
		appCommand("command:create", applicationTime),
	)
	if err != nil {
		t.Fatal(err)
	}
	if status == domain.StatusDraft {
		return introduction
	}
	introduction, _ = introduction.AuthorizeUpload(appCommand("command:upload", applicationTime), 1)
	if status == domain.StatusUploadAuthorized {
		return introduction
	}
	complete, _ := domain.NewMediaRef(
		"asset:1", "audio/ogg", 2048, time.Minute, digest("b"),
	)
	introduction, _ = introduction.ConfirmUpload(complete, appCommand("command:confirm", applicationTime), 2)
	if status == domain.StatusUploaded {
		return introduction
	}
	introduction, _ = introduction.StartTranscription(appCommand("command:start", applicationTime), 3)
	return introduction
}

func TestBeginUploadRequiresExactVersionedConsent(t *testing.T) {
	controller := gomock.NewController(t)
	consent := NewMockConsentGate(controller)
	consent.EXPECT().Effective(gomock.Any(), "member:1", "voice.introduction", uint64(2)).
		Return(false, nil)
	service := NewService(
		NewMockStore(controller), consent, NewMockMediaManager(controller),
		NewMockTranscriber(controller), NewMockKeyer(controller), NewMockIDSource(controller),
		func() time.Time { return applicationTime },
	)
	_, err := service.BeginUpload(context.Background(), BeginUploadRequest{
		CommandID: "command:1", OwnerID: "member:1",
		PurposeID: "voice.introduction", PurposeVersion: 2, ContentType: "audio/ogg",
	})
	if !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("expected consent required, got %v", err)
	}
}

func TestBeginUploadPersistsConsentBeforeSigning(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockStore(controller)
	consent := NewMockConsentGate(controller)
	media := NewMockMediaManager(controller)
	keyer := NewMockKeyer(controller)
	ids := NewMockIDSource(controller)
	consent.EXPECT().Effective(gomock.Any(), "member:1", "voice.introduction", uint64(2)).
		Return(true, nil)
	ids.EXPECT().NewID("intro_asset").Return("asset:1")
	keyer.EXPECT().Key(gomock.Any()).Return(digest("a"), nil)
	ids.EXPECT().NewID("introduction").Return("introduction:1")
	store.EXPECT().Create(gomock.Any(), gomock.Cond(func(value domain.Introduction) bool {
		return value.Status() == domain.StatusDraft && value.Consent().Version() == 2 &&
			value.Media().AssetID() == "asset:1"
	})).DoAndReturn(func(_ context.Context, value domain.Introduction) (domain.Introduction, bool, error) {
		return value, false, nil
	})
	keyer.EXPECT().Key("asset:1").Return(digest("b"), nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(value domain.Introduction) bool {
		return value.Status() == domain.StatusUploadAuthorized
	}), uint64(1), "command:1.upload").Return(nil)
	media.EXPECT().AuthorizeUpload(gomock.Any(), "member:1", "introduction:1", "asset:1").
		Return(UploadAccess{URL: "https://upload.invalid", ExpiresAt: applicationTime.Add(time.Minute)}, nil)

	result, err := NewService(
		store, consent, media, NewMockTranscriber(controller), keyer, ids,
		func() time.Time { return applicationTime },
	).BeginUpload(context.Background(), BeginUploadRequest{
		CommandID: "command:1", OwnerID: "member:1",
		PurposeID: "voice.introduction", PurposeVersion: 2, ContentType: "audio/ogg",
	})
	if err != nil || result.Introduction.Status() != domain.StatusUploadAuthorized ||
		result.Access.URL == "" {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
}

func TestConfirmUploadRechecksConsentAndUsesAuthoritativeMetadata(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockStore(controller)
	consent := NewMockConsentGate(controller)
	media := NewMockMediaManager(controller)
	keyer := NewMockKeyer(controller)
	introduction := applicationIntroduction(t, domain.StatusUploadAuthorized, domain.NewRetention(time.Time{}, false))
	store.EXPECT().FindByID(gomock.Any(), "introduction:1").Return(introduction, nil)
	consent.EXPECT().Effective(gomock.Any(), "member:1", "voice.introduction", uint64(2)).Return(true, nil)
	complete, _ := domain.NewMediaRef("asset:1", "audio/ogg", 2048, time.Minute, digest("b"))
	media.EXPECT().Inspect(gomock.Any(), "asset:1").Return(complete, nil)
	keyer.EXPECT().Key(gomock.Any()).Return(digest("c"), nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(value domain.Introduction) bool {
		return value.Media().Complete() && value.Status() == domain.StatusUploaded
	}), uint64(2), "command:confirm.confirmed").Return(nil)

	updated, err := NewService(
		store, consent, media, NewMockTranscriber(controller), keyer,
		NewMockIDSource(controller), func() time.Time { return applicationTime },
	).ConfirmUpload(context.Background(), "introduction:1", "command:confirm")
	if err != nil || !updated.Media().Complete() {
		t.Fatalf("updated=%+v, err=%v", updated, err)
	}
}

func TestTranscriptionOutcomesNeverStoreRawText(t *testing.T) {
	tests := []struct {
		name        string
		provider    TranscriptionResult
		providerErr error
		want        domain.Status
		wantErr     error
	}{
		{"ready", TranscriptionResult{
			Outcome: TranscriptionCompleted,
			Transcript: func() domain.TranscriptRef {
				ref, _ := domain.NewTranscriptRef("transcript:1", "en", 90)
				return ref
			}(),
		}, nil, domain.StatusReady, nil},
		{"uncertain", TranscriptionResult{Outcome: TranscriptionUncertain}, nil, domain.StatusTranscriptionUncertain, ErrTranscriptionUncertain},
		{"failed", TranscriptionResult{Outcome: TranscriptionFailed}, nil, domain.StatusTranscriptionFailed, ErrTranscriptionFailed},
		{"outage", TranscriptionResult{}, errors.New("provider down"), domain.StatusTranscriptionUncertain, ErrTranscriptionUncertain},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			store := NewMockStore(controller)
			consent := NewMockConsentGate(controller)
			transcriber := NewMockTranscriber(controller)
			keyer := NewMockKeyer(controller)
			introduction := applicationIntroduction(t, domain.StatusUploaded, domain.NewRetention(time.Time{}, false))
			store.EXPECT().FindByID(gomock.Any(), "introduction:1").Return(introduction, nil)
			consent.EXPECT().Effective(gomock.Any(), "member:1", "voice.introduction", uint64(2)).Return(true, nil)
			keyer.EXPECT().Key("asset:1").Return(digest("c"), nil)
			store.EXPECT().Update(gomock.Any(), gomock.Any(), uint64(3), "command:tx.start").
				DoAndReturn(func(_ context.Context, value domain.Introduction, _ uint64, _ string) error {
					introduction = value
					return nil
				})
			transcriber.EXPECT().Transcribe(gomock.Any(), gomock.Cond(func(request TranscriptionRequest) bool {
				return request.ConsentVersion == 2 && request.AssetID == "asset:1"
			})).Return(test.provider, test.providerErr)
			keyer.EXPECT().Key(gomock.Any()).Return(digest("d"), nil)
			store.EXPECT().Update(gomock.Any(), gomock.Cond(func(value domain.Introduction) bool {
				return value.Status() == test.want
			}), uint64(4), gomock.Any()).Return(nil)

			result, err := NewService(
				store, consent, NewMockMediaManager(controller), transcriber, keyer,
				NewMockIDSource(controller), func() time.Time { return applicationTime },
			).Transcribe(context.Background(), "introduction:1", "command:tx")
			if result.Introduction.Status() != test.want || !errors.Is(err, test.wantErr) {
				t.Fatalf("status=%q, err=%v", result.Introduction.Status(), err)
			}
		})
	}
}

func TestRevokeCancelsActiveTranscriptionAndPurgesImmediately(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockStore(controller)
	transcriber := NewMockTranscriber(controller)
	media := NewMockMediaManager(controller)
	keyer := NewMockKeyer(controller)
	introduction := applicationIntroduction(t, domain.StatusTranscribing, domain.NewRetention(time.Time{}, false))
	store.EXPECT().FindByID(gomock.Any(), "introduction:1").Return(introduction, nil)
	keyer.EXPECT().Key("revoke").Return(digest("c"), nil)
	store.EXPECT().Update(gomock.Any(), gomock.Any(), uint64(4), "command:revoke.revoke").
		DoAndReturn(func(_ context.Context, value domain.Introduction, _ uint64, _ string) error {
			introduction = value
			return nil
		})
	transcriber.EXPECT().Cancel(gomock.Any(), "introduction:1").Return(nil)
	store.EXPECT().FindByID(gomock.Any(), "introduction:1").
		DoAndReturn(func(context.Context, string) (domain.Introduction, error) { return introduction, nil })
	media.EXPECT().Delete(gomock.Any(), "asset:1").Return(nil)
	keyer.EXPECT().Key("purged").Return(digest("d"), nil)
	store.EXPECT().Update(gomock.Any(), gomock.Cond(func(value domain.Introduction) bool {
		return value.Status() == domain.StatusRevoked && value.DataStatus() == domain.DataPurged &&
			value.Media().AssetID() == ""
	}), uint64(5), "command:revoke.cleanup.purged").Return(nil)

	result, err := NewService(
		store, NewMockConsentGate(controller), media, transcriber, keyer,
		NewMockIDSource(controller), func() time.Time { return applicationTime },
	).Revoke(context.Background(), "introduction:1", "command:revoke")
	if err != nil || result.DataStatus() != domain.DataPurged {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
}

func TestPurgeDoesNotDeleteActiveOrRetainedData(t *testing.T) {
	controller := gomock.NewController(t)
	store := NewMockStore(controller)
	active := applicationIntroduction(t, domain.StatusUploaded, domain.NewRetention(time.Time{}, false))
	store.EXPECT().FindByID(gomock.Any(), "introduction:1").Return(active, nil)
	_, err := NewService(
		store, NewMockConsentGate(controller), NewMockMediaManager(controller),
		NewMockTranscriber(controller), NewMockKeyer(controller), NewMockIDSource(controller),
		func() time.Time { return applicationTime },
	).Purge(context.Background(), "introduction:1", "command:purge")
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("active data purge = %v", err)
	}
}
