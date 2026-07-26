package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

const (
	MinSignedTTL = time.Minute
	MaxSignedTTL = 15 * time.Minute
)

var (
	ErrAccessDenied       = errors.New("media access denied")
	ErrInvalidRequest     = errors.New("invalid media access request")
	ErrStorageUnavailable = errors.New("media storage unavailable")
)

type AccessService struct {
	assets AssetRepository
	policy AccessPolicy
	signer Signer
	now    func() time.Time
}

func NewAccessService(assets AssetRepository, policy AccessPolicy, signer Signer, now func() time.Time) AccessService {
	return AccessService{assets: assets, policy: policy, signer: signer, now: now}
}

type UploadRequest struct {
	SubjectID string
	Purpose   string
	Asset     domain.Asset
	TTL       time.Duration
}

func (service AccessService) RequestUpload(ctx context.Context, request UploadRequest) (SignedAccess, error) {
	if !validIdentityAndPurpose(request.SubjectID, request.Purpose) || !validTTL(request.TTL) ||
		request.Asset.ID() == "" || request.Asset.IsDeleted() {
		return SignedAccess{}, ErrInvalidRequest
	}
	if err := service.policy.Authorize(ctx, AccessDecision{
		SubjectID: strings.TrimSpace(request.SubjectID),
		OwnerID:   request.Asset.OwnerID(),
		AssetID:   request.Asset.ID(),
		Purpose:   strings.TrimSpace(request.Purpose),
		Action:    ActionUpload,
	}); err != nil {
		return SignedAccess{}, ErrAccessDenied
	}
	now := service.now().UTC()
	if !request.Asset.AvailableAt(now) {
		return SignedAccess{}, domain.ErrAssetUnavailable
	}
	expiresAt := now.Add(request.TTL)
	signed, err := service.signer.SignUpload(ctx, UploadSigningRequest{
		ObjectKey:   request.Asset.ObjectKey(),
		ContentType: request.Asset.ContentType(),
		Size:        request.Asset.Size(),
		Checksum:    request.Asset.Checksum(),
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return SignedAccess{}, ErrStorageUnavailable
	}
	return normalizeSignedAccess(signed, expiresAt)
}

type ReadRequest struct {
	SubjectID string
	AssetID   string
	Purpose   string
	TTL       time.Duration
}

func (service AccessService) RequestRead(ctx context.Context, request ReadRequest) (SignedAccess, error) {
	if !validIdentityAndPurpose(request.SubjectID, request.Purpose) ||
		strings.TrimSpace(request.AssetID) == "" || !validTTL(request.TTL) {
		return SignedAccess{}, ErrInvalidRequest
	}
	asset, err := service.assets.FindByID(ctx, strings.TrimSpace(request.AssetID))
	if err != nil || asset.ID() == "" {
		// Repository details and asset existence are deliberately hidden.
		return SignedAccess{}, domain.ErrAssetUnavailable
	}
	now := service.now().UTC()
	if !asset.AvailableAt(now) {
		return SignedAccess{}, domain.ErrAssetUnavailable
	}
	if err := service.policy.Authorize(ctx, AccessDecision{
		SubjectID: strings.TrimSpace(request.SubjectID),
		OwnerID:   asset.OwnerID(),
		AssetID:   asset.ID(),
		Purpose:   strings.TrimSpace(request.Purpose),
		Action:    ActionRead,
	}); err != nil {
		return SignedAccess{}, ErrAccessDenied
	}
	expiresAt := now.Add(request.TTL)
	signed, err := service.signer.SignRead(ctx, ReadSigningRequest{
		ObjectKey: asset.ObjectKey(),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return SignedAccess{}, ErrStorageUnavailable
	}
	return normalizeSignedAccess(signed, expiresAt)
}

func validTTL(ttl time.Duration) bool {
	return ttl >= MinSignedTTL && ttl <= MaxSignedTTL
}

func validIdentityAndPurpose(subjectID, purpose string) bool {
	return strings.TrimSpace(subjectID) != "" && strings.TrimSpace(purpose) != ""
}

func normalizeSignedAccess(signed SignedAccess, expectedExpiry time.Time) (SignedAccess, error) {
	if strings.TrimSpace(signed.URL) == "" {
		return SignedAccess{}, ErrStorageUnavailable
	}
	// The domain owns expiry. Provider adapters cannot silently extend access.
	signed.ExpiresAt = expectedExpiry
	return signed, nil
}
