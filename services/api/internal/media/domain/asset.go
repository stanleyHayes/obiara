package domain

import (
	"errors"
	"mime"
	"strings"
	"time"
)

var (
	ErrInvalidAsset     = errors.New("invalid media asset")
	ErrInvalidChecksum  = errors.New("invalid media checksum")
	ErrAssetUnavailable = errors.New("media asset unavailable")
	ErrRetentionActive  = errors.New("media retention policy is active")
	ErrLegalHoldActive  = errors.New("media legal hold is active")
	ErrInvalidDeletion  = errors.New("invalid media deletion time")
)

// Checksum describes the digest that storage must verify when accepting bytes.
type Checksum struct {
	algorithm string
	value     string
}

func NewChecksum(algorithm, value string) (Checksum, error) {
	algorithm = strings.ToLower(strings.TrimSpace(algorithm))
	value = strings.ToLower(strings.TrimSpace(value))
	if algorithm != "sha256" || len(value) != 64 {
		return Checksum{}, ErrInvalidChecksum
	}
	for _, character := range value {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return Checksum{}, ErrInvalidChecksum
		}
	}
	return Checksum{algorithm: algorithm, value: value}, nil
}

func (checksum Checksum) Algorithm() string { return checksum.algorithm }
func (checksum Checksum) Value() string     { return checksum.value }

// Retention records a compliance boundary. A legal hold always overrides the
// timestamp and prevents deletion until explicitly released by an authorized
// compliance workflow.
type Retention struct {
	until     time.Time
	legalHold bool
}

func NewRetention(until time.Time, legalHold bool) Retention {
	return Retention{until: until.UTC(), legalHold: legalHold}
}

func (retention Retention) Until() time.Time { return retention.until }
func (retention Retention) LegalHold() bool  { return retention.legalHold }

// Asset is immutable media metadata. State changes return a new value so
// callers cannot mutate persisted or authorization-checked metadata in place.
type Asset struct {
	id          string
	objectKey   string
	ownerID     string
	contentType string
	size        int64
	checksum    Checksum
	// duration is zero for media that has none (an image). For audio it is
	// the only account of length the server will trust: the listening gate
	// that Sow enforces counts against this, and a client-declared number
	// would let a member claim an answer they never recorded.
	duration  time.Duration
	createdAt time.Time
	expiresAt time.Time
	retention Retention
	deletedAt time.Time
}

type NewAssetParams struct {
	ID          string
	ObjectKey   string
	OwnerID     string
	ContentType string
	Size        int64
	Checksum    Checksum
	Duration    time.Duration
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Retention   Retention
}

func NewAsset(params NewAssetParams) (Asset, error) {
	params.ID = strings.TrimSpace(params.ID)
	params.ObjectKey = strings.TrimSpace(params.ObjectKey)
	params.OwnerID = strings.TrimSpace(params.OwnerID)
	params.ContentType = strings.ToLower(strings.TrimSpace(params.ContentType))
	if params.ID == "" || params.ObjectKey == "" || params.OwnerID == "" || params.Size <= 0 ||
		params.Checksum.algorithm == "" || params.CreatedAt.IsZero() || params.Duration < 0 {
		return Asset{}, ErrInvalidAsset
	}
	if parsed, _, err := mime.ParseMediaType(params.ContentType); err != nil || parsed == "" {
		return Asset{}, ErrInvalidAsset
	}
	createdAt := params.CreatedAt.UTC()
	expiresAt := params.ExpiresAt.UTC()
	if !expiresAt.IsZero() && !expiresAt.After(createdAt) {
		return Asset{}, ErrInvalidAsset
	}
	if !params.Retention.until.IsZero() && params.Retention.until.Before(createdAt) {
		return Asset{}, ErrInvalidAsset
	}
	return Asset{
		id:          params.ID,
		objectKey:   params.ObjectKey,
		ownerID:     params.OwnerID,
		contentType: params.ContentType,
		size:        params.Size,
		checksum:    params.Checksum,
		duration:    params.Duration,
		createdAt:   createdAt,
		expiresAt:   expiresAt,
		retention:   params.Retention,
	}, nil
}

func (asset Asset) ID() string              { return asset.id }
func (asset Asset) ObjectKey() string       { return asset.objectKey }
func (asset Asset) OwnerID() string         { return asset.ownerID }
func (asset Asset) ContentType() string     { return asset.contentType }
func (asset Asset) Size() int64             { return asset.size }
func (asset Asset) Checksum() Checksum      { return asset.checksum }
func (asset Asset) Duration() time.Duration { return asset.duration }
func (asset Asset) CreatedAt() time.Time    { return asset.createdAt }
func (asset Asset) ExpiresAt() time.Time    { return asset.expiresAt }
func (asset Asset) Retention() Retention    { return asset.retention }
func (asset Asset) DeletedAt() time.Time    { return asset.deletedAt }
func (asset Asset) IsDeleted() bool         { return !asset.deletedAt.IsZero() }
func (asset Asset) AvailableAt(now time.Time) bool {
	return !asset.IsDeleted() && (asset.expiresAt.IsZero() || now.Before(asset.expiresAt))
}

func (asset Asset) MarkDeleted(now time.Time) (Asset, error) {
	if now.IsZero() {
		return Asset{}, ErrInvalidDeletion
	}
	now = now.UTC()
	if now.Before(asset.createdAt) {
		return Asset{}, ErrInvalidDeletion
	}
	if asset.retention.legalHold {
		return Asset{}, ErrLegalHoldActive
	}
	if !asset.retention.until.IsZero() && now.Before(asset.retention.until) {
		return Asset{}, ErrRetentionActive
	}
	if asset.IsDeleted() {
		return asset, nil
	}
	asset.deletedAt = now
	return asset, nil
}
