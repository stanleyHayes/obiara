// Package secrets validates provider-neutral runtime secret metadata without
// reading, persisting, hashing or logging secret values.
package secrets

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var (
	ErrMissing          = errors.New("required runtime secret is missing")
	ErrRotationMetadata = errors.New("secret rotation metadata is invalid")
	ErrStale            = errors.New("runtime secret exceeds maximum rotation age")
)

type Service string

const (
	API    Service = "api"
	Worker Service = "worker"
)

type Definition struct {
	Name              string
	RotatedAtVariable string
	Services          []Service
	MaxAge            time.Duration
}

// Inventory contains names and policy metadata only. Values remain in the
// approved runtime secret store and are deliberately never accepted here.
func Inventory() []Definition {
	return []Definition{
		{Name: "MONGODB_URI", RotatedAtVariable: "MONGODB_URI_ROTATED_AT", Services: []Service{API, Worker}, MaxAge: 90 * 24 * time.Hour},
		{Name: "RESEND_WEBHOOK_SECRET", RotatedAtVariable: "RESEND_WEBHOOK_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "LIVENESS_HMAC_SECRET", RotatedAtVariable: "LIVENESS_HMAC_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "COMMERCE_HMAC_SECRET", RotatedAtVariable: "COMMERCE_HMAC_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "ADMIN_HMAC_SECRET", RotatedAtVariable: "ADMIN_HMAC_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "NNOBOA_INVITE_SECRET", RotatedAtVariable: "NNOBOA_INVITE_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "SEED_HMAC_SECRET", RotatedAtVariable: "SEED_HMAC_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "SAFEGUARDING_HMAC_SECRET", RotatedAtVariable: "SAFEGUARDING_HMAC_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
		{Name: "CIRCLE_HMAC_SECRET", RotatedAtVariable: "CIRCLE_HMAC_SECRET_ROTATED_AT", Services: []Service{API}, MaxAge: 90 * 24 * time.Hour},
	}
}

// ValidateRuntime fails closed outside development/test when a service's
// required secret or its RFC3339 rotation evidence is absent, future-dated or
// stale. Errors contain variable names only, never values.
func ValidateRuntime(service Service, environment string, getenv func(string) string, now time.Time) error {
	if environment == "development" || environment == "test" || environment == "local" {
		return nil
	}
	if service != API && service != Worker {
		return fmt.Errorf("unknown service %q", service)
	}
	if now.IsZero() {
		return ErrRotationMetadata
	}
	for _, definition := range Inventory() {
		if !slices.Contains(definition.Services, service) {
			continue
		}
		if strings.TrimSpace(getenv(definition.Name)) == "" {
			return fmt.Errorf("%w: %s", ErrMissing, definition.Name)
		}
		rotatedAt, e := time.Parse(time.RFC3339, strings.TrimSpace(getenv(definition.RotatedAtVariable)))
		if e != nil || rotatedAt.After(now.Add(5*time.Minute)) {
			return fmt.Errorf("%w: %s", ErrRotationMetadata, definition.RotatedAtVariable)
		}
		if now.Sub(rotatedAt) > definition.MaxAge {
			return fmt.Errorf("%w: %s", ErrStale, definition.Name)
		}
	}
	return nil
}
