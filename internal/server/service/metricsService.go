package service

import (
	"context"
	"time"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/repository"
	"go.uber.org/zap"
)

type MetricService struct {
	metricsRepo   repository.Repository
	storeRepo     repository.Repository
	storeInterval int
}

func NewMetricService(repo repository.Repository, storeRepository repository.Repository, storeInterval int) MetricService {
	return MetricService{metricsRepo: repo, storeRepo: storeRepository, storeInterval: storeInterval}
}

func (s *MetricService) Init() {
	if s.storeInterval > 0 {
		go s.saveMetricsWithPeriod()
	}
}

func (s *MetricService) Ping(ctx context.Context) error {
	return s.storeRepo.Ping(ctx)
}

func (s *MetricService) SaveMetric(metric models.Metric) error {
	existMetric, _ := s.metricsRepo.GetMetricByID(metric.ID)
	if existMetric != nil {
		if metric.MType == models.Counter {
			if existMetric.Delta != nil {
				*metric.Delta += *existMetric.Delta
			}

		}
	}
	err := s.metricsRepo.SaveMetric(metric)
	if err != nil {
		return err
	}
	//сохраняем синхронно в файл если не задан интервал сохранения
	if s.storeInterval == 0 {
		err = s.store()
		if err != nil {
			logger.Log.Error("store data error:", zap.Error(err))
		}
	}

	return nil
}

func (s *MetricService) GetMetric(id string) (*models.Metric, error) {
	return s.metricsRepo.GetMetricByID(id)
}

func (s *MetricService) GetAllMetric() ([]models.Metric, error) {
	return s.metricsRepo.GetAll()
}

func (s *MetricService) RemoveMetric(id string) error {
	return s.metricsRepo.DeleteMetricByID(id)
}

func (s *MetricService) saveMetricsWithPeriod() {

	ticker := time.NewTicker(time.Duration(s.storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := s.store()
		if err != nil {
			logger.Log.Error("store data error:", zap.Error(err))
		}

	}
}

func (s *MetricService) store() error {
	metrics, err := s.metricsRepo.GetAll()
	if err != nil {
		return err
	}
	err = s.storeRepo.SaveMetrics(metrics)
	if err != nil {
		return err
	}
	logger.Log.Debug("store metrics is success")
	return nil
}
