package drrehearsal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongoReplicaSetIsolatedRestorePreservesSourceAndDetectsCorruption(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := testmongodb.Run(ctx, "mongo:8.0.13", testmongodb.WithReplicaSet("rs0"))
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Skipf("Docker unavailable: %v", err)
	}
	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	separator := "?"
	if strings.Contains(uri, "?") {
		separator = "&"
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri + separator + "directConnection=true"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	sourceName, targetName := "staging_source", "isolated_dr_20260727"
	source := client.Database(sourceName)
	if _, err := source.Collection("ledger").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "entry", Value: 1}}, Options: options.Index().SetUnique(true)}); err != nil {
		t.Fatal(err)
	}
	session, err := client.StartSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx context.Context) (any, error) {
		if _, insertErr := source.Collection("ledger").InsertMany(tx, []any{
			bson.D{{Key: "_id", Value: "l-1"}, {Key: "entry", Value: 1}, {Key: "amount", Value: 40}},
			bson.D{{Key: "_id", Value: "l-2"}, {Key: "entry", Value: 2}, {Key: "amount", Value: -40}},
		}); insertErr != nil {
			return nil, insertErr
		}
		_, insertErr := source.Collection("audit").InsertOne(tx, bson.D{{Key: "_id", Value: "a-1"}, {Key: "event", Value: "balanced"}})
		return nil, insertErr
	})
	if err != nil {
		t.Fatal(err)
	}
	watermark := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
	before, err := mongoSnapshot(ctx, source, watermark)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyDatabase(ctx, source, client.Database(targetName)); err != nil {
		t.Fatal(err)
	}
	restored, err := mongoSnapshot(ctx, client.Database(targetName), watermark)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(before, restored); err != nil {
		t.Fatalf("isolated restore verification: %v", err)
	}
	if _, err := client.Database(targetName).Collection("ledger").DeleteOne(ctx, bson.D{{Key: "_id", Value: "l-2"}}); err != nil {
		t.Fatal(err)
	}
	corrupt, err := mongoSnapshot(ctx, client.Database(targetName), watermark)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(before, corrupt); err == nil {
		t.Fatal("corrupt isolated restore passed verification")
	}
	after, err := mongoSnapshot(ctx, source, watermark)
	if err != nil {
		t.Fatal(err)
	}
	beforeDigest, _ := before.CanonicalDigest()
	afterDigest, _ := after.CanonicalDigest()
	if beforeDigest != afterDigest {
		t.Fatal("source changed during rehearsal")
	}
	if err := client.Database(targetName).Drop(ctx); err != nil {
		t.Fatal(err)
	}
	names, err := client.ListDatabaseNames(ctx, bson.D{{Key: "name", Value: sourceName}})
	if err != nil || len(names) != 1 {
		t.Fatalf("isolated cleanup affected source: names=%v err=%v", names, err)
	}
}

func copyDatabase(ctx context.Context, source, target *mongo.Database) error {
	names, err := source.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		cursor, err := source.Collection(name).Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
		if err != nil {
			return err
		}
		var docs []bson.Raw
		if err := cursor.All(ctx, &docs); err != nil {
			return err
		}
		values := make([]any, len(docs))
		for i := range docs {
			values[i] = docs[i]
		}
		if len(values) > 0 {
			if _, err := target.Collection(name).InsertMany(ctx, values); err != nil {
				return err
			}
		}
		indexCursor, err := source.Collection(name).Indexes().List(ctx)
		if err != nil {
			return err
		}
		var indexes []bson.M
		if err := indexCursor.All(ctx, &indexes); err != nil {
			return err
		}
		for _, index := range indexes {
			if index["name"] == "_id_" {
				continue
			}
			model := mongo.IndexModel{Keys: index["key"]}
			if unique, ok := index["unique"].(bool); ok {
				model.Options = options.Index().SetUnique(unique)
			}
			if _, err := target.Collection(name).Indexes().CreateOne(ctx, model); err != nil {
				return err
			}
		}
	}
	return nil
}

func mongoSnapshot(ctx context.Context, db *mongo.Database, watermark time.Time) (Snapshot, error) {
	names, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return Snapshot{}, err
	}
	sort.Strings(names)
	result := Snapshot{Watermark: watermark}
	for _, name := range names {
		cursor, err := db.Collection(name).Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
		if err != nil {
			return Snapshot{}, err
		}
		var docs []bson.Raw
		if err := cursor.All(ctx, &docs); err != nil {
			return Snapshot{}, err
		}
		hash := sha256.New()
		for _, doc := range docs {
			_, _ = hash.Write(doc)
		}
		indexCursor, err := db.Collection(name).Indexes().List(ctx)
		if err != nil {
			return Snapshot{}, err
		}
		var indexes []bson.M
		if err := indexCursor.All(ctx, &indexes); err != nil {
			return Snapshot{}, err
		}
		indexNames := make([]string, 0, len(indexes))
		for _, index := range indexes {
			if value, ok := index["name"].(string); ok {
				indexNames = append(indexNames, value)
			}
		}
		sort.Strings(indexNames)
		result.Collections = append(result.Collections, CollectionFact{
			Name: name, Count: int64(len(docs)), Digest: hex.EncodeToString(hash.Sum(nil)),
			Indexes: indexNames, Transaction: true, Audit: true,
		})
	}
	return result, nil
}
