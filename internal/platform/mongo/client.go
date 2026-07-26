// Package mongo holds the platform-level MongoDB conventions for the api
// service (agent_plan.md §7.4, §8): one client setup path, transactions for
// multi-document invariants, duplicate-key translation, and an idempotent
// migration runner. Module adapters (e.g. member/adapters/outbound/mongodb)
// build repositories on top of these helpers; provider/driver types never
// cross into domain code.
package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Connect establishes a MongoDB client and verifies it with a ping. The
// caller controls the deadline via ctx (see platform/config timeouts).
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo (start a local MongoDB or set MONGODB_URI): %w", err)
	}
	return client, nil
}
