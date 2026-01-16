package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type MemoryRepository struct {
	metrics map[string]models.Metric
	mu      sync.Mutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		metrics: make(map[string]models.Metric),
		mu:      sync.Mutex{},
	}

}

func (r *MemoryRepository) RestoreFromFile(filepath string) error {

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		logger.Log.Info("Storage file does not exist, starting fresh")
		return fmt.Errorf("file with metrics is not exist")
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		logger.Log.Error("read file error, starting fresh", zap.Error(err))
		return err
	}
	var metrics []models.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return err
	}
	err = r.SaveMetrics(metrics)
	if err != nil {
		return err
	}

	logger.Log.Sugar().Infof("succesfully restored %d metrics from file", len(metrics))
	return nil
}

func (r *MemoryRepository) RestoreFromDB(pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("database connection pool is nil")
	}

	query := `SELECT id, m_type, delta, value FROM metrics ORDER BY id`
	ctx := context.Background()
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var metrics []models.Metric
	for rows.Next() {
		var m models.Metric
		err := rows.Scan(
			&m.ID,
			&m.MType,
			&m.Delta,
			&m.Value,
		)
		if err != nil {
			return err
		}
		metrics = append(metrics, m)
	}

	if err := rows.Err(); err != nil {
		return err
	}
	err = r.SaveMetrics(metrics)
	if err != nil {
		return err
	}
	logger.Log.Sugar().Infof("succesfully restored %d metrics from database", len(metrics))
	return nil
}

func (r *MemoryRepository) GetMetricByID(id string) (models.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metric, ok := r.metrics[id]
	if !ok {
		return models.Metric{}, fmt.Errorf("not found metric: %s", id)
	}
	return metric, nil
}

func (r *MemoryRepository) SaveMetric(metric models.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if metric.MType == models.Counter {
		existMetric, ok := r.metrics[metric.ID]
		if ok {
			if existMetric.Delta != nil {
				*metric.Delta += *existMetric.Delta
			}
		}
	}
	r.metrics[metric.ID] = metric
	return nil
}

func (r *MemoryRepository) SaveMetrics(metrics []models.Metric) error {
	for _, metric := range metrics {
		err := r.SaveMetric(metric)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MemoryRepository) DeleteMetricByID(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.metrics[id]
	if !ok {
		return fmt.Errorf("delete with error. Not found: %s", id)
	}
	delete(r.metrics, id)
	return nil
}

func (r *MemoryRepository) FindAll() ([]models.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metrics := make([]models.Metric, 0, len(r.metrics))
	for _, metric := range r.metrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil

}
