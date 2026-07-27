//go:build integration

package performance

import (
	"context"
	"fmt"
	apimongo "github.com/stanleyHayes/obiara/internal/platform/mongo"
	testmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"testing"
	"time"
)

func TestMongoBoundedLoad(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	box, e := testmongodb.Run(ctx, "mongo:8.0.13")
	if e != nil {
		t.Fatal(e)
	}
	defer box.Terminate(context.Background())
	uri, _ := box.ConnectionString(ctx)
	client, e := apimongo.Connect(ctx, uri)
	if e != nil {
		t.Fatal(e)
	}
	defer client.Disconnect(context.Background())
	collection := client.Database("performance_test").Collection("synthetic_load")
	if _, e = collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "key", Value: 1}}, Options: options.Index().SetUnique(true)}); e != nil {
		t.Fatal(e)
	}
	profile := Profile{Name: "mongodb_insert_read", Requests: 800, Concurrency: 16, MaxP95: 500 * time.Millisecond, MaxErrorRate: 0}
	result, e := Run(ctx, profile, func(call context.Context, index int) error {
		key := fmt.Sprintf("synthetic-%04d", index)
		if _, e := collection.InsertOne(call, bson.M{"key": key, "value": index}); e != nil {
			return e
		}
		var found struct {
			Key string `bson:"key"`
		}
		if e := collection.FindOne(call, bson.M{"key": key}).Decode(&found); e != nil {
			return e
		}
		if found.Key != key {
			return fmt.Errorf("wrong synthetic row")
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if !result.Within(profile) {
		t.Fatalf("load budget failed: %#v", result)
	}
	raw, _ := result.JSON()
	t.Logf("performance_evidence=%s", raw)
	stats := client.Database("admin").RunCommand(ctx, bson.D{{Key: "serverStatus", Value: 1}})
	if stats.Err() != nil {
		t.Fatal(stats.Err())
	}
}
