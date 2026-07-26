package application

import (
	"context"
	"time"
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application

type LocaleSource interface {
	CurrentLocale(context.Context) (string, error)
}

type LocaleCatalog interface {
	Resolve(context.Context, string) (LocaleReview, error)
}

type MediaInspector interface {
	Inspect(context.Context, string) (MediaMetadata, error)
}

// Advisor is a provider-neutral evidence port. Implementations may compose
// through the AI gateway, but this package has no vendor dependency.
type Advisor interface {
	Screen(context.Context, ScreeningInput) (Advisory, error)
}

// Adjudicator is separate from Advisor so provider/model output can never be
// the final decision.
type Adjudicator interface {
	Decide(context.Context, ScreeningInput, Advisory) (Adjudication, error)
}

type HumanReview interface {
	Route(context.Context, ReviewCase) (string, error)
}

type IDSource interface {
	NewID() string
}

type LocaleReview struct {
	Tag        string
	Version    uint64
	Reviewed   bool
	ReviewedAt time.Time
}

type MediaMetadata struct {
	MIME       string
	Bytes      int64
	DurationMs int64
}
