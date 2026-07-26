package application

import (
	"context"
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/sow/domain"
)

var (
	ErrUnavailable           = errors.New("sow service unavailable")
	ErrInsufficientAllowance = errors.New("insufficient weekly seed allowance")
)

//go:generate mockgen -source=ports.go -destination=mock_ports_test.go -package=application
type Screening interface {
	Screen(context.Context, string, []string) (ScreeningDecision, error)
}
type Acceptance interface {
	// Accept atomically stores the accepted sow and spends its allowance units.
	Accept(context.Context, domain.Sow) (domain.Sow, bool, error)
}
type Keyer interface {
	Key(namespace, value string) (string, error)
}
type IDSource interface{ NewID() string }

type ScreeningDecision struct {
	Approved  bool
	Reference string
}
