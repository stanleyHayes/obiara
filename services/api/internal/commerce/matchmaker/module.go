package matchmaker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/adapters/outbound/mongodb"
	"github.com/stanleyHayes/obiara/services/api/internal/commerce/matchmaker/application"
)

type Module struct {
	Engagements application.Service
	Catalog     *mongodb.Catalog
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

func (clock) Now() time.Time { return time.Now().UTC() }

func NewModule(ctx context.Context, database *mongo.Database) (Module, error) {
	repository := mongodb.New(database)
	if err := repository.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	catalog := mongodb.NewCatalog(database)
	if err := catalog.EnsureIndexes(ctx); err != nil {
		return Module{}, err
	}
	return Module{
		Engagements: application.New(repository, catalog, ids{}, clock{}),
		Catalog:     catalog,
	}, nil
}
