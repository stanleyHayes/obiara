package application

import (
	"context"
	"time"

	"github.com/stanleyHayes/obiara/services/api/internal/media/domain"
)

type Action string

const (
	ActionUpload Action = "upload"
	ActionRead   Action = "read"
)

// AssetRepository persists authoritative immutable metadata.
type AssetRepository interface {
	FindByID(context.Context, string) (domain.Asset, error)
}

// AccessPolicy is the server-side authorization boundary. Implementations
// decide whether a subject may act for the asset owner and declared purpose.
type AccessPolicy interface {
	Authorize(context.Context, AccessDecision) error
}

type AccessDecision struct {
	SubjectID string
	OwnerID   string
	AssetID   string
	Purpose   string
	Action    Action
}

// Signer is a narrow provider-neutral port. Adapters translate these requests
// into vendor-specific signed URLs without leaking provider types.
type Signer interface {
	SignUpload(context.Context, UploadSigningRequest) (SignedAccess, error)
	SignRead(context.Context, ReadSigningRequest) (SignedAccess, error)
}

type UploadSigningRequest struct {
	ObjectKey   string
	ContentType string
	Size        int64
	Checksum    domain.Checksum
	ExpiresAt   time.Time
}

type ReadSigningRequest struct {
	ObjectKey string
	ExpiresAt time.Time
}

type SignedAccess struct {
	URL       string
	ExpiresAt time.Time
}
