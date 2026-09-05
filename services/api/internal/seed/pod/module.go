// Package pod is the composition root of the pod: a recording resting at a
// member's house front for named people to open.
//
// It is the last step of the sow. A sow that a reviewer released is delivered
// by being placed here; without this, releasing one marked it delivered and
// nobody received anything.
package pod

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/seed/pod/application"
)

type Module struct {
	Pods application.Service
}

// ErrDependenciesRequired reports a module built without the ports that
// decide who may place a pod, who may open one, and whether either of them
// still wants anything to do with the other.
var ErrDependenciesRequired = errors.New(
	"pod module requires an authorizer, an eligibility check, a media issuer and a keying secret",
)

// NewModule composes the pod against one database.
func NewModule(
	ctx context.Context,
	database *mongo.Database,
	authorizer application.Authorizer,
	eligibility application.PlaybackEligibility,
	issuer application.MediaIssuer,
	secret string,
) (Module, error) {
	if authorizer == nil || eligibility == nil || issuer == nil || secret == "" {
		return Module{}, ErrDependenciesRequired
	}
	keyer, err := privacy.NewKeyer([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	repository := mongodb.NewRepository(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Pods: application.NewService(repository, authorizer, eligibility, issuer, keyer, idSource{}, time.Now),
	}, nil
}

type idSource struct{}

func (idSource) NewID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return "pod_" + hex.EncodeToString(value)
}
