package mongodb

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoClient wraps the MongoDB client
type MongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

// validateMongoURI performs basic validation on MongoDB URI to prevent injection attacks
func validateMongoURI(uri string) error {
	if uri == "" {
		return errors.New("mongodb URI cannot be empty")
	}

	// Parse the URI to validate format
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid mongodb URI format: %w", err)
	}

	// Ensure scheme is mongodb or mongodb+srv
	scheme := parsedURI.Scheme
	if scheme != "mongodb" && scheme != "mongodb+srv" {
		return fmt.Errorf("invalid mongodb URI scheme: %s (must be mongodb or mongodb+srv)", scheme)
	}

	// Validate that host is present
	if parsedURI.Host == "" {
		return errors.New("mongodb URI must contain a host")
	}

	return nil
}

// NewMongoClient creates a new MongoDB client with optimized connection pooling and security features
func NewMongoClient(uri, database string) (*MongoClient, error) {
	// Validate URI before connecting to prevent injection attacks
	if err := validateMongoURI(uri); err != nil {
		return nil, fmt.Errorf("mongodb URI validation failed: %w", err)
	}

	// Validate database name
	if database == "" {
		return nil, errors.New("database name cannot be empty")
	}
	if strings.ContainsAny(database, "/\\. \"$*<>:|?") {
		return nil, errors.New("database name contains invalid characters")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configure connection pool for better performance and security
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).       // Maximum connections in the pool
		SetMinPoolSize(10).        // Minimum connections to maintain
		SetMaxConnIdleTime(30 * time.Second). // Close idle connections after 30s
		SetConnectTimeout(10 * time.Second).  // Connection timeout
		SetServerSelectionTimeout(10 * time.Second). // Server selection timeout
		SetRetryWrites(true).      // Enable retryable writes for better reliability
		SetRetryReads(true)        // Enable retryable reads for better reliability

	// Enable TLS for production environments (if URI uses mongodb+srv or has tls=true)
	if strings.Contains(uri, "mongodb+srv://") || strings.Contains(uri, "tls=true") || strings.Contains(uri, "ssl=true") {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12, // Enforce minimum TLS 1.2
		}
		clientOptions.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	// Ping the database with read preference to verify connection
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
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
