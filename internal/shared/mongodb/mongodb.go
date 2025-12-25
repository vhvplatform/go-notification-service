package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient wraps the MongoDB client
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(uri, database string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoClient{
		client:   client,
		database: client.Database(database),
	}, nil
}

// Collection returns a collection handle
func (c *MongoClient) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Disconnect closes the MongoDB connection
func (c *MongoClient) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Database returns the database handle
func (c *MongoClient) Database() *mongo.Database {
	return c.database
}
