package redis

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

// Redis holds the Redis client
type Redis struct {
	client *redis.Client
	mu     sync.Mutex
}

// NewRedis initializes a new Redis client
func NewRedis() (*Redis, error) {
	r := &Redis{}
	dsn := r.ConfigDSN()
	log.Printf("Initializing Redis with DSN: %s", dsn)

	// Parse DSN to create Options
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		log.Printf("Failed to parse Redis DSN: %v", err)
		return nil, fmt.Errorf("failed to parse Redis DSN: %v", err)
	}

	// Create Redis client
	r.mu.Lock()
	defer r.mu.Unlock()
	r.client = redis.NewClient(opt)
	if err := r.client.Ping(context.Background()).Err(); err != nil {
		log.Printf("Failed to ping Redis: %v", err)
		r.client = nil // Ensure client is nil on failure
		return nil, fmt.Errorf("failed to ping Redis: %v", err)
	}

	log.Printf("Successfully connected to Redis. Client: %p", r.client)
	return r, nil
}

// ConfigDSN returns the Data Source Name (DSN) for Redis
func (r *Redis) ConfigDSN() string {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	if host == "" || port == "" {
		log.Printf("Redis environment variables not set, using default: localhost:6379")
		return "redis://localhost:6379/0"
	}
	return fmt.Sprintf("redis://:%s@%s:%s/0", password, host, port)
}

// Close shuts down the Redis client
func (r *Redis) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client != nil {
		log.Printf("Closing Redis client: %p", r.client)
		err := r.client.Close()
		r.client = nil // Set to nil after closing
		return err
	}
	log.Printf("No Redis client to close")
	return nil
}

// GetClient returns the underlying redis.Client for custom operations
func (r *Redis) GetClient() *redis.Client {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.client == nil {
		log.Printf("Error: Redis client is nil")
		return nil
	}
	return r.client
}