package service

import (
	"context"
	"fmt"
	"time"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/config"
	"github.com/avakumov/metrics/internal/server/repository"
	"go.uber.org/zap"
)

type MetricService struct {
	metricsRepo   repository.Repository
	storeRepo     repository.Repository
	storeInterval int
	isRestore     bool
	ctx           context.Context
}

func NewMetricService(
	ctx context.Context,
	repo repository.Repository,
	storeRepository repository.Repository,
	options config.Options) MetricService {
	return MetricService{
		metricsRepo:   repo,
		storeRepo:     storeRepository,
		storeInterval: options.StoreInterval,
		isRestore:     options.Restore,
		ctx:           ctx}
}

func (s *MetricService) Init() {
	//restore data from store
	if s.isRestore {
		err := s.restore()
		if err != nil {
			logger.Log.Error("restore error:", zap.Error(err))
		}
	}

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

func (s *MetricService) SaveMetrics(metrics []models.Metric) error {
	if metrics == nil {
		return fmt.Errorf("metrics is nil")
	}
	if len(metrics) == 0 {
		return fmt.Errorf("metrics is empty")
	}
	for _, m := range metrics {
		err := s.SaveMetric(m)
		if err != nil {
			return err
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
	for {
		select {
		case <-s.ctx.Done():
			err := s.store()
			if err != nil {
				logger.Log.Error("store data error:", zap.Error(err))
			}
			logger.Log.Info("shutdown on save metric with period success")
			return
		case <-ticker.C:
			err := s.store()
			if err != nil {
				logger.Log.Error("store data error:", zap.Error(err))
			}
		}
	}
}

func (s *MetricService) restore() error {
	if s.metricsRepo == nil {
		return fmt.Errorf("metric repo is nil")
	}
	if s.storeRepo == nil {
		return fmt.Errorf("store repo is nil")
	}
	metrics, err := s.storeRepo.GetAll()
	if err != nil {
		return err
	}
	err = s.metricsRepo.SaveMetrics(metrics)
	if err != nil {
		return err
	}
	return nil
}

func (s *MetricService) store() error {
	if s.metricsRepo == nil {
		return fmt.Errorf("metric repo is nil")
	}
	if s.storeRepo == nil {
		return fmt.Errorf("store repo is nil")
	}
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
