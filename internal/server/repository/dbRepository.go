package repository

import (
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/database"
	"github.com/avakumov/metrics/internal/server/pgerrors"

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
	Pool           *pgxpool.Pool
	RetryDurations []time.Duration
}

func NewDBRepository(DSN string) (*DBRepository, error) {
	logger.Log.Debug("starting new DB repository")
	var db = &DBRepository{
		Pool:           nil,
		RetryDurations: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
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

func (db *DBRepository) GetAllWithoutRetry(ctx context.Context) ([]models.Metric, error) {
	pool := db.Pool
	if pool == nil {
		return nil, fmt.Errorf("pool of database repository is nil")
	}

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

func (db *DBRepository) GetMetricByIDWithoutRetry(ctx context.Context, id string) (*models.Metric, error) {
	pool := db.Pool
	if pool == nil {
		return nil, fmt.Errorf("pool of database repository is nil")
	}

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

func (db *DBRepository) SaveMetricWithouRetry(ctx context.Context, metric models.Metric) error {
	pool := db.Pool
	if pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}

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

func (db *DBRepository) SaveMetricsWithoutRetry(ctx context.Context, metrics []models.Metric) error {

	if db.Pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}

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

// delete metric without retry
func (db *DBRepository) DeleteMetricByIDWithoutRetry(ctx context.Context, id string) error {
	pool := db.Pool

	if pool == nil {
		return fmt.Errorf("pool of database repository is nil")
	}

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

func (db *DBRepository) SaveMetrics(ctx context.Context, metrics []models.Metric) error {
	fn := func() error {
		return db.SaveMetricsWithoutRetry(ctx, metrics)
	}
	return withRetry(fn, db.RetryDurations)
}

func (db *DBRepository) SaveMetric(ctx context.Context, metric models.Metric) error {
	fn := func() error { return db.SaveMetricWithouRetry(ctx, metric) }
	return withRetry(fn, db.RetryDurations)
}

func (db *DBRepository) GetMetricByID(ctx context.Context, id string) (*models.Metric, error) {
	fn := func() (*models.Metric, error) { return db.GetMetricByIDWithoutRetry(ctx, id) }
	return withRetryWithResult(fn, db.RetryDurations)
}

func (db *DBRepository) GetAll(ctx context.Context) ([]models.Metric, error) {
	fn := func() ([]models.Metric, error) { return db.GetAllWithoutRetry(ctx) }
	return withRetryWithResult(fn, db.RetryDurations)
}

func (db *DBRepository) DeleteMetricByID(ctx context.Context, id string) error {
	fn := func() error { return db.DeleteMetricByIDWithoutRetry(ctx, id) }
	return withRetry(fn, db.RetryDurations)
}

// при retryable ошибке будет делать еще запросы
func withRetry(f func() error, durations []time.Duration) error {
	pgErrorClassifier := pgerrors.NewPostgresErrorClassifier()
	var err error
	for i, duration := range durations {
		err = f()
		switch pgErrorClassifier.Classify(err) {
		case pgerrors.NonRetriable:
			return err
		case pgerrors.Retriable:
			logger.Log.Info("Request retriable error",
				zap.Duration("duration", duration),
				zap.Int("attempt", i+1),
				zap.Int("total attempt", len(durations)),
				zap.Error(err))
			time.Sleep(duration)
		default: // No error
			return nil
		}
	}
	return err
}

// при retryable ошибке будет делать еще запросы
func withRetryWithResult[T any](fn func() (T, error), durations []time.Duration,
) (T, error) {

	pgErrorClassifier := pgerrors.NewPostgresErrorClassifier()
	var zero T
	var err error

	for i, duration := range durations {
		result, err := fn()
		switch pgErrorClassifier.Classify(err) {
		case pgerrors.NonRetriable:
			return zero, err
		case pgerrors.Retriable:
			logger.Log.Info("Request retriable error",
				zap.Duration("duration", duration),
				zap.Int("attempt", i+1),
				zap.Int("total attempt", len(durations)),
				zap.Error(err))
			time.Sleep(duration)
		default: // No error
			return result, nil
		}
	}
	return zero, err
}
