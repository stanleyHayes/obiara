// Package authz is the composition root of the authorization kernel.
// Role assignments are asserted by the identity/admin contexts (E16); this
// module only evaluates them.
package authz

import (
	"github.com/stanleyHayes/obiara/services/api/internal/authz/application"
)

type Module struct {
	Authorizer application.Authorizer
}

func NewModule() Module {
	return Module{Authorizer: application.NewAuthorizer()}
}
