package membership

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/adapters/outbound/privacy"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/membership/application"
)

type Module struct {
	Membership application.Service
	Keyer      privacy.Keyer
}

type ids struct{}

func (ids) NewID() string {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value)
}

type clock struct{}

func (clock) Now() time.Time { return time.Now() }

func NewModule(ctx context.Context, database *mongo.Database, secret string) (Module, error) {
	repository := mongodb.New(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	keyer, err := privacy.New([]byte(secret))
	if err != nil {
		return Module{}, err
	}
	return Module{
		Membership: application.New(repository, nil, ids{}, clock{}),
		Keyer:      keyer,
	}, nil
}
