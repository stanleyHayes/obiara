package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
	"go.uber.org/mock/gomock"
)

func TestRequestReadEnforcesOwnershipPurposeAndTTL(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	asset := applicationAsset(t, now, time.Time{}, domain.Retention{})

	for _, test := range []struct {
		name    string
		request ReadRequest
		wantErr error
	}{
		{name: "ttl below minimum", request: ReadRequest{SubjectID: "member-1", AssetID: "asset-1", Purpose: "profile", TTL: MinSignedTTL - 1}, wantErr: ErrInvalidRequest},
		{name: "ttl above maximum", request: ReadRequest{SubjectID: "member-1", AssetID: "asset-1", Purpose: "profile", TTL: MaxSignedTTL + 1}, wantErr: ErrInvalidRequest},
		{name: "purpose required", request: ReadRequest{SubjectID: "member-1", AssetID: "asset-1", TTL: MinSignedTTL}, wantErr: ErrInvalidRequest},
		{name: "subject required", request: ReadRequest{AssetID: "asset-1", Purpose: "profile", TTL: MinSignedTTL}, wantErr: ErrInvalidRequest},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			service := NewAccessService(NewMockAssetRepository(ctrl), NewMockAccessPolicy(ctrl), NewMockSigner(ctrl), func() time.Time { return now })
			_, err := service.RequestRead(context.Background(), test.request)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("RequestRead() error = %v, want %v", err, test.wantErr)
			}
		})
	}

	t.Run("authorization carries owner and purpose", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockAssetRepository(ctrl)
		policy := NewMockAccessPolicy(ctrl)
		signer := NewMockSigner(ctrl)
		repository.EXPECT().FindByID(gomock.Any(), "asset-1").Return(asset, nil)
		policy.EXPECT().Authorize(gomock.Any(), AccessDecision{
			SubjectID: "delegate-1", OwnerID: "member-1", AssetID: "asset-1",
			Purpose: "profile", Action: ActionRead,
		}).Return(nil)
		signer.EXPECT().SignRead(gomock.Any(), ReadSigningRequest{
			ObjectKey: "owners/member-1/asset-1", ExpiresAt: now.Add(5 * time.Minute),
		}).Return(SignedAccess{URL: "https://storage.invalid/read", ExpiresAt: now.Add(time.Hour)}, nil)

		service := NewAccessService(repository, policy, signer, func() time.Time { return now })
		signed, err := service.RequestRead(context.Background(), ReadRequest{
			SubjectID: "delegate-1", AssetID: "asset-1", Purpose: "profile", TTL: 5 * time.Minute,
		})
		if err != nil {
			t.Fatal(err)
		}
		if !signed.ExpiresAt.Equal(now.Add(5 * time.Minute)) {
			t.Fatalf("provider extended expiry to %v", signed.ExpiresAt)
		}
	})
}

func TestRequestReadDenySafeFailures(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	providerSecret := "provider request id secret-123"
	request := ReadRequest{SubjectID: "member-1", AssetID: "asset-1", Purpose: "profile", TTL: 5 * time.Minute}
	retainedDeleted := applicationAsset(t, now.Add(-2*time.Hour), time.Time{}, domain.NewRetention(now.Add(-time.Hour), false))
	retainedDeleted, err := retainedDeleted.MarkDeleted(now)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name      string
		asset     domain.Asset
		findErr   error
		policyErr error
		signerErr error
		wantErr   error
	}{
		{name: "unknown asset", findErr: errors.New("mongo: no documents"), wantErr: domain.ErrAssetUnavailable},
		{name: "expired asset", asset: applicationAsset(t, now.Add(-2*time.Hour), now.Add(-time.Hour), domain.Retention{}), wantErr: domain.ErrAssetUnavailable},
		{name: "retained metadata stays unavailable after deletion", asset: retainedDeleted, wantErr: domain.ErrAssetUnavailable},
		{name: "denied policy", asset: applicationAsset(t, now, time.Time{}, domain.Retention{}), policyErr: errors.New("owner mismatch"), wantErr: ErrAccessDenied},
		{name: "provider failure is redacted", asset: applicationAsset(t, now, time.Time{}, domain.Retention{}), signerErr: errors.New(providerSecret), wantErr: ErrStorageUnavailable},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			repository := NewMockAssetRepository(ctrl)
			policy := NewMockAccessPolicy(ctrl)
			signer := NewMockSigner(ctrl)
			repository.EXPECT().FindByID(gomock.Any(), "asset-1").Return(test.asset, test.findErr)
			if test.asset.ID() != "" && test.asset.AvailableAt(now) {
				policy.EXPECT().Authorize(gomock.Any(), gomock.Any()).Return(test.policyErr)
				if test.policyErr == nil {
					signer.EXPECT().SignRead(gomock.Any(), gomock.Any()).Return(SignedAccess{}, test.signerErr)
				}
			}
			service := NewAccessService(repository, policy, signer, func() time.Time { return now })
			_, err := service.RequestRead(context.Background(), request)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("RequestRead() error = %v, want %v", err, test.wantErr)
			}
			if strings.Contains(err.Error(), providerSecret) {
				t.Fatal("provider details leaked through application error")
			}
		})
	}
}

func TestRequestUploadIncludesChecksumMetadata(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	asset := applicationAsset(t, now, time.Time{}, domain.Retention{})
	ctrl := gomock.NewController(t)
	policy := NewMockAccessPolicy(ctrl)
	signer := NewMockSigner(ctrl)
	policy.EXPECT().Authorize(gomock.Any(), AccessDecision{
		SubjectID: "member-1", OwnerID: "member-1", AssetID: "asset-1",
		Purpose: "profile", Action: ActionUpload,
	}).Return(nil)
	signer.EXPECT().SignUpload(gomock.Any(), gomock.Cond(func(request UploadSigningRequest) bool {
		return request.Checksum.Algorithm() == "sha256" &&
			request.Checksum.Value() == "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" &&
			request.ContentType == "image/jpeg" && request.Size == 128 &&
			request.ExpiresAt.Equal(now.Add(5*time.Minute))
	})).Return(SignedAccess{URL: "https://storage.invalid/upload"}, nil)

	service := NewAccessService(NewMockAssetRepository(ctrl), policy, signer, func() time.Time { return now })
	_, err := service.RequestUpload(context.Background(), UploadRequest{
		SubjectID: "member-1", Purpose: "profile", Asset: asset, TTL: 5 * time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func applicationAsset(t *testing.T, createdAt, expiresAt time.Time, retention domain.Retention) domain.Asset {
	t.Helper()
	checksum, err := domain.NewChecksum("sha256", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	asset, err := domain.NewAsset(domain.NewAssetParams{
		ID: "asset-1", ObjectKey: "owners/member-1/asset-1", OwnerID: "member-1",
		ContentType: "image/jpeg", Size: 128, Checksum: checksum,
		CreatedAt: createdAt, ExpiresAt: expiresAt, Retention: retention,
	})
	if err != nil {
		t.Fatal(err)
	}
	return asset
}
