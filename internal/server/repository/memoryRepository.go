package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/avakumov/metrics/internal/models"
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

func (r *MemoryRepository) GetMetricByID(id string) (*models.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metric, ok := r.metrics[id]
	if !ok {
		return nil, fmt.Errorf("not found metric: %s", id)
	}
	return &metric, nil
}

func (r *MemoryRepository) SaveMetric(metric models.Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *MemoryRepository) GetAll() ([]models.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metrics := make([]models.Metric, 0, len(r.metrics))
	for _, metric := range r.metrics {
		metrics = append(metrics, metric)
	}
	return metrics, nil

}

func (r *MemoryRepository) Ping(ctx context.Context) error {
	return fmt.Errorf("ping is not available. The memory repository is being used")
}

func (r *MemoryRepository) Close() {

}
