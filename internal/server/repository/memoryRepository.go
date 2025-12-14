package repository

import (
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
		existMetric, exists := r.metrics[metric.ID]
		if exists {
			if existMetric.Delta != nil {
				*metric.Delta += *existMetric.Delta
			}
		}
	}
	r.metrics[metric.ID] = metric
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
