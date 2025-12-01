package repository

import (
	"fmt"
	"sync"

	"github.com/avakumov/metrics/internal/models"
)

type MemoryRepository struct {
	metrics map[string]models.Metric
	mu      sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		metrics: make(map[string]models.Metric),
		mu:      sync.RWMutex{},
	}
}

func (r *MemoryRepository) GetMetricById(id string) (models.Metric, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	metric, exist := r.metrics[id]
	if !exist {
		return models.Metric{}, fmt.Errorf("Не найдена метрика: %s", id)
	}
	return metric, nil
}

func (r *MemoryRepository) SaveMetric(metric models.Metric) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics[metric.Id] = metric
	return nil
}

func (r *MemoryRepository) DeleteMetricById(id string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	metric, exist := r.metrics[id]
	if !exist {
		return fmt.Errorf("Удаление не выполнено. Не найдена метрика: %s", id)
	}
	r.metrics[id] = metric
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
