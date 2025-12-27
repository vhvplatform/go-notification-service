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

// NewMongoClient creates a new MongoDB client with optimized connection pooling
func NewMongoClient(uri, database string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure connection pool for better performance
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).       // Maximum connections in the pool
		SetMinPoolSize(10).        // Minimum connections to maintain
		SetMaxConnIdleTime(30 * time.Second). // Close idle connections after 30s
		SetConnectTimeout(10 * time.Second).  // Connection timeout
		SetServerSelectionTimeout(10 * time.Second) // Server selection timeout

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

// CreateIndexes creates indexes for a collection
func (c *MongoClient) CreateIndexes(ctx context.Context, collectionName string, indexes []mongo.IndexModel) error {
	collection := c.database.Collection(collectionName)
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}
