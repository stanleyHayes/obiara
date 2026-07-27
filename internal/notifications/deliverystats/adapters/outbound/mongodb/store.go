// Package mongodb answers delivery-statistics queries across channel logs.
package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Store struct {
	database *mongo.Database
}

func NewStore(database *mongo.Database) *Store {
	return &Store{database: database}
}

func (store *Store) CountByStatus(ctx context.Context, collection string, since time.Time) (map[string]int, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{"at": bson.M{"$gte": since.UTC()}}},
		bson.M{"$group": bson.M{"_id": "$status", "count": bson.M{"$sum": 1}}},
	}
	cursor, err := store.database.Collection(collection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	counts := map[string]int{}
	for cursor.Next(ctx) {
		var row struct {
			Status string `bson:"_id"`
			Count  int    `bson:"count"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, err
		}
		counts[row.Status] = row.Count
	}
	return counts, cursor.Err()
}
