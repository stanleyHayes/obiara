// Package application exposes the authorization boundary used by inbound
// adapters and use cases. Deny-by-default is enforced by the domain policy;
// this layer turns decisions into errors and carries actor context.
package application

import (
	"errors"

	"github.com/stanleyHayes/obiara/services/api/internal/authz/domain"
)

// ErrForbidden is returned whenever the policy denies an action. Inbound
// adapters map it to the stable "forbidden" machine code without exposing
// policy internals.
var ErrForbidden = errors.New("action is not permitted")

type Authorizer struct{}

func NewAuthorizer() Authorizer { return Authorizer{} }

// Require returns ErrForbidden unless the policy explicitly grants the
// action. Callers must treat any error as terminal for the request.
func (Authorizer) Require(subject domain.Subject, action string, resource domain.Resource) error {
	if decision := domain.Evaluate(subject, action, resource); !decision.Allowed {
		return ErrForbidden
	}
	return nil
}
