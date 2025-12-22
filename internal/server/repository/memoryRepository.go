package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
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

func (r *MemoryRepository) Restore(filepath string) {

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		logger.Log.Info("Storage file does not exist, starting fresh")
		return
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		logger.Log.Error("read file error, starting fresh", zap.Error(err))
		return
	}
	var metrics []models.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		logger.Log.Error("unmarshal json error, starting fresh", zap.Error(err))
		return
	}
	err = r.SaveMetrics(metrics)
	if err != nil {
		logger.Log.Error("restore metrics from file error", zap.Error(err))
		return
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
