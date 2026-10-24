// Package database manages the MongoDB client lifecycle.
// All other packages that need database access receive a *mongo.Client or
// a *mongo.Collection through dependency injection so that this package
// remains the single place responsible for connection setup and teardown.
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client wraps the raw mongo.Client together with the target database name
// so callers can obtain collections without repeating the DB name.
type Client struct {
	raw    *mongo.Client
	dbName string
}

// Connect creates a new MongoDB client, establishes a connection,
// and pings the primary to confirm it is reachable.
//
// The caller is responsible for calling Disconnect when the program exits.
func Connect(ctx context.Context, uri, dbName string) (*Client, error) {
	// Build client options: 10-second connection timeout, sensible pool size.
	opts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetMaxPoolSize(10)

	raw, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect: %w", err)
	}

	// Ping the primary server to verify the connection is truly up.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err = raw.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}

	log.Printf("[DB] Connected to MongoDB — database: %q", dbName)
	return &Client{raw: raw, dbName: dbName}, nil
}

// Disconnect cleanly closes the MongoDB connection.
// Should be deferred right after Connect in main.
func (c *Client) Disconnect(ctx context.Context) {
	if err := c.raw.Disconnect(ctx); err != nil {
		log.Printf("[DB] Error while disconnecting: %v", err)
		return
	}
	log.Println("[DB] MongoDB connection closed")
}

// Collection returns a handle to the named collection inside the configured database.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.raw.Database(c.dbName).Collection(name)
}

// EnsureIndexes creates all indexes required by the application.
// It is safe to call on every start-up – MongoDB will skip already-existing indexes.
func (c *Client) EnsureIndexes(ctx context.Context) error {
	coll := c.Collection("registrations")

	// Unique index on email to prevent duplicate registrations.
	emailIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("unique_email"),
	}

	// Index on submittedAt for fast chronological export queries.
	timeIdx := mongo.IndexModel{
		Keys:    bson.D{{Key: "submittedAt", Value: -1}},
		Options: options.Index().SetName("idx_submitted_at"),
	}

	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{emailIdx, timeIdx})
	if err != nil {
		return fmt.Errorf("EnsureIndexes: %w", err)
	}

	log.Println("[DB] Indexes verified")
	return nil
}
