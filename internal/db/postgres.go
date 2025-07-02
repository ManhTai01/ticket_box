package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the database connection pool
type DB struct {
	pool *pgxpool.Pool
}

// NewDB initializes a new PostgreSQL connection pool
func NewDB() (*DB, error) {
	db := &DB{}
	dsn := db.ConfigDSN()
	log.Printf("Attempting to connect to PostgreSQL with DSN: %s", dsn)

	db, err := NewPostgresConn(dsn, 5, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL connection: %v", err)
	}
	return db, nil
}

// ConfigDSN returns the Data Source Name (DSN) for PostgreSQL
func (db *DB) ConfigDSN() string {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	// Kiểm tra và sử dụng giá trị mặc định nếu biến môi trường trống
	if dbHost == "" {
		dbHost = "localhost"
		log.Printf("Using default POSTGRES_HOST: %s", dbHost)
	}
	if dbPort == "" {
		dbPort = "5432"
		log.Printf("Using default POSTGRES_PORT: %s", dbPort)
	}
	if dbUser == "" {
		dbUser = "postgres"
		log.Printf("Using default POSTGRES_USER: %s", dbUser)
	}
	if dbPass == "" {
		log.Printf("POSTGRES_PASSWORD not set, using empty password")
	}
	if dbName == "" {
		dbName = "postgres" // Database mặc định nếu không có
		log.Printf("Using default POSTGRES_DB: %s", dbName)
	}

	// Tạo DSN theo định dạng postgres://
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)
	return dsn
}

// NewPostgresConn initializes a new PostgreSQL connection with retry logic
func NewPostgresConn(dsn string, maxRetries int, retryDelay time.Duration) (*DB, error) {
	var db *DB
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db, err = openConnection(dsn)
		if err == nil {
			if err = db.Ping(ctx); err == nil {
				log.Printf("Successfully connected to PostgreSQL on attempt %d", attempt)
				return db, nil
			}
			db.Close()
		}

		log.Printf("Failed to connect to PostgreSQL (attempt %d/%d): %v", attempt, maxRetries, err)
		if attempt == maxRetries {
			return nil, fmt.Errorf("failed to connect to PostgreSQL after %d attempts: %v", maxRetries, err)
		}

		// Exponential backoff
		backoff := retryDelay * time.Duration(1<<uint(attempt-1))
		log.Printf("Retrying in %v...", backoff)
		time.Sleep(backoff)
	}

	return nil, err
}

// openConnection opens a new PostgreSQL connection pool
func openConnection(dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN configuration: %v", err)
	}
	config.MaxConns = 10 // Cấu hình tối đa kết nối, có thể điều chỉnh
	config.MinConns = 1  // Tối thiểu kết nối

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL pool: %v", err)
	}

	return &DB{pool: pool}, nil
}

// Ping attempts to ping the database
func (db *DB) Ping(ctx context.Context) error {
	if db.pool == nil {
		return fmt.Errorf("database pool is nil")
	}
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}
	return nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}