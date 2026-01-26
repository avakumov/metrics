package repository

import (
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/database"

	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBRepository struct {
	Pool *pgxpool.Pool
}

func NewDBRepository(DSN string) (*DBRepository, error) {
	logger.Log.Debug("starting new DB repository")
	var db = &DBRepository{
		Pool: nil,
	}
	if len(DSN) == 0 {
		return db, fmt.Errorf("DSN string is empty: %s", DSN)
	}

	config, err := pgxpool.ParseConfig(DSN)
	if err != nil {
		return db, fmt.Errorf("failed to parse connection string: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return db, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return db, fmt.Errorf("failed to ping database: %w", err)
	}
	db.Pool = pool
	logger.Log.Info("✅ Connected to PostgreSQL by Pool")

	//TODO:перенести куда-нибудь
	err = database.RunMigrations(DSN)
	if err != nil {
		logger.Log.Error("migration error:", zap.Error(err))
	}

	return db, nil
}

func (db *DBRepository) GetAll() ([]models.Metric, error) {
	pool := db.Pool
	if pool == nil {
		return nil, fmt.Errorf("pool of database repository is nil")
	}
	ctx := context.Background()

	query := `
	SELECT id, m_type, delta, value FROM metrics
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()
	var metrics []models.Metric
	for rows.Next() {
		var m models.Metric
		var delta sql.NullInt64
		var value sql.NullFloat64
		err := rows.Scan(
			&m.ID,
			&m.MType,
			&delta,
			&value,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}
		if delta.Valid {
			m.Delta = new(int64)
			*m.Delta = delta.Int64
		}
		if value.Valid {
			m.Value = new(float64)
			*m.Value = value.Float64
		}

		metrics = append(metrics, m)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return metrics, nil
}

func (db *DBRepository) GetMetricByID(id string) (*models.Metric, error) {
	pool := db.Pool
	if pool == nil {
		return nil, fmt.Errorf("pool of database repository is nil")
	}

	ctx := context.Background()

	query := `
	SELECT id, m_type, delta, value FROM metrics WHERE id = $1
	`
	row := pool.QueryRow(ctx, query, id)

	var m models.Metric
	var delta sql.NullInt64
	var value sql.NullFloat64

	err := row.Scan(
		&m.ID,
		&m.MType,
		&delta,
		&value,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("metric with id %s not found", id)
		}
		return nil, fmt.Errorf("failed to get metric row: %w", err)
	}
	if delta.Valid {
		m.Delta = new(int64)
		*m.Delta = delta.Int64
	}
	if value.Valid {
		m.Value = new(float64)
		*m.Value = value.Float64
	}
	return &m, nil
}

func (db *DBRepository) SaveMetric(metric models.Metric) error {
	pool := db.Pool
	if pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}

	ctx := context.Background()

	query := `
	INSERT INTO metrics (id, m_type, delta, value)
	VALUES ($1, $2, $3, $4)
  ON CONFLICT (id, m_type) 
  DO UPDATE SET 
    delta = EXCLUDED.delta,
    value = EXCLUDED.value,
    hash = EXCLUDED.hash,
    updated_at = CURRENT_TIMESTAMP
	`

	_, err := pool.Exec(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		return fmt.Errorf("failed to save metric %s: %w", metric.ID, err)
	}
	return nil
}

func (db *DBRepository) SaveMetrics(metrics []models.Metric) error {

	if db.Pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback(ctx) //nolint:errcheck
		}
	}()

	query := `
	INSERT INTO metrics (id, m_type, delta, value, hash)
	VALUES ($1, $2, $3, $4, $5)
  ON CONFLICT (id, m_type) 
  DO UPDATE SET 
    delta = EXCLUDED.delta,
    value = EXCLUDED.value,
    hash = EXCLUDED.hash,
    updated_at = CURRENT_TIMESTAMP
	`

	batch := &pgx.Batch{}
	for _, m := range metrics {
		batch.Queue(query, m.ID, m.MType, m.Delta, m.Value, m.Hash)
	}
	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	// Проверяем результаты
	for range metrics {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("batch insert error: %w", err)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("close batch: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)

	}
	tx = nil
	return nil

}

func (db *DBRepository) DeleteMetricByID(id string) error {
	pool := db.Pool

	if pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}
	ctx := context.Background()

	query := `
	DELETE FROM metrics WHERE id = $1
	`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete metric %s: %w", id, err)
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}
	return nil
}

func (db *DBRepository) Ping(ctx context.Context) error {

	if db.Pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}
	return db.Pool.Ping(ctx)
}

func (db *DBRepository) Close() {

	if db.Pool == nil {
		return
	}
	db.Pool.Close()
	logger.Log.Info("✅ Closed to PostgreSQL")
}
