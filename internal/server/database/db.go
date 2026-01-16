package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool  *pgxpool.Pool
	SQLDB *sql.DB
}

func Connect(DSN string) (*Database, error) {

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("✅ Connected to PostgreSQL by Pool")

	// Создаём sql.DB для миграций
	sqlDB, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("✅ Connected to PostgreSQL")

	return &Database{Pool: pool, SQLDB: sqlDB}, nil
}

func (db *Database) Close() {
	logger.Log.Info("✅ Closed to PostgreSQL")
	db.Pool.Close()
}
