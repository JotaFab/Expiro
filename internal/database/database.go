package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"Expiro/internal/database/sqlc"

	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	Health() map[string]string

	// Close terminates the database connection.
	Close() error

	// Queries gives access to generated sqlc queries.
	Queries() *sqlc.Queries
}

type service struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		username, password, host, port, database, schema,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create db pool: %v", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}

	q := sqlc.New(pool)

	dbInstance = &service{
		pool:    pool,
		queries: q,
	}
	return dbInstance
}

func (s *service) Queries() *sqlc.Queries {
	return s.queries
}

// Health checks the health of the database connection by pinging the database.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	if err := s.pool.Ping(ctx); err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Pool stats (available in pgxpool)
	ps := s.pool.Stat()
	stats["acquired"] = fmt.Sprintf("%d", ps.AcquiredConns())
	stats["idle"] = fmt.Sprintf("%d", ps.IdleConns())
	stats["total"] = fmt.Sprintf("%d", ps.TotalConns())
	stats["acquire_count"] = fmt.Sprintf("%d", ps.AcquireCount())
	stats["cancel_count"] = fmt.Sprintf("%d", ps.CanceledAcquireCount())

	return stats
}

// Close closes the database pool.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	s.pool.Close()
	return nil
}
