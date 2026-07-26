package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// WithTransaction runs fn inside a MongoDB transaction (agent_plan.md §7.4:
// transactions guard multi-document invariants where required). fn receives
// the session context and must use it for every operation that belongs to
// the transaction. The transaction commits when fn returns nil and aborts
// otherwise. Single-document writes do not need this helper.
func WithTransaction(ctx context.Context, client *mongo.Client, fn func(context.Context) error) error {
	session, err := client.StartSession()
	if err != nil {
		return fmt.Errorf("start mongo session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc context.Context) (any, error) {
		return nil, fn(sc)
	})
	if err != nil {
		return fmt.Errorf("mongo transaction: %w", err)
	}
	return nil
}
