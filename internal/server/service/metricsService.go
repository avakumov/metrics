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
}

func NewMetricService(
	repo repository.Repository,
	storeRepository repository.Repository,
	options config.Options) MetricService {
	return MetricService{
		metricsRepo:   repo,
		storeRepo:     storeRepository,
		storeInterval: options.StoreInterval,
		isRestore:     options.Restore,
	}
}

func (s *MetricService) Init(ctx context.Context) {
	//restore data from store
	if s.isRestore {
		err := s.restoreMetrics(ctx)
		if err != nil {
			logger.Log.Error("restore error:", zap.Error(err))
		}
	}

	if s.storeInterval > 0 {
		go s.saveMetricsWithPeriod(ctx)
	}
}

func (s *MetricService) Ping(ctx context.Context) error {
	return s.storeRepo.Ping(ctx)
}

func (s *MetricService) SaveMetric(ctx context.Context, metric models.Metric) error {
	existMetric, _ := s.metricsRepo.GetMetricByID(ctx, metric.ID)
	if existMetric != nil {
		if metric.MType == models.Counter {
			if existMetric.Delta != nil {
				*metric.Delta += *existMetric.Delta
			}

		}
	}
	err := s.metricsRepo.SaveMetric(ctx, metric)
	if err != nil {
		return err
	}
	//сохраняем только одну запись
	if s.storeInterval == 0 {
		err = s.storeMetric(ctx, metric)
		if err != nil {
			logger.Log.Error("store data error:", zap.Error(err))
		}
	}

	return nil
}

func (s *MetricService) SaveMetrics(ctx context.Context, metrics []models.Metric) error {
	if metrics == nil {
		return fmt.Errorf("metrics is nil")
	}
	if len(metrics) == 0 {
		return fmt.Errorf("metrics is empty")
	}
	for _, m := range metrics {
		err := s.SaveMetric(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricService) GetMetric(ctx context.Context, id string) (*models.Metric, error) {
	return s.metricsRepo.GetMetricByID(ctx, id)
}

func (s *MetricService) GetAllMetric(ctx context.Context) ([]models.Metric, error) {
	return s.metricsRepo.GetAll(ctx)
}

func (s *MetricService) RemoveMetric(ctx context.Context, id string) error {
	return s.metricsRepo.DeleteMetricByID(ctx, id)
}

func (s *MetricService) saveMetricsWithPeriod(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.storeInterval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			err := s.storeMetrics(ctx)
			if err != nil {
				logger.Log.Error("store data error:", zap.Error(err))
			}
			logger.Log.Info("shutdown on save metric with period success")
			return
		case <-ticker.C:
			err := s.storeMetrics(ctx)
			if err != nil {
				logger.Log.Error("store data error:", zap.Error(err))
			}
		}
	}
}

func (s *MetricService) restoreMetrics(ctx context.Context) error {
	if s.metricsRepo == nil {
		return fmt.Errorf("metric repo is nil")
	}
	if s.storeRepo == nil {
		return fmt.Errorf("store repo is nil")
	}
	metrics, err := s.storeRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	err = s.metricsRepo.SaveMetrics(ctx, metrics)
	if err != nil {
		return err
	}

	logger.Log.Debug("restore metric is success")
	return nil
}

func (s *MetricService) storeMetric(ctx context.Context, metric models.Metric) error {
	if s.metricsRepo == nil {
		return fmt.Errorf("metric repo is nil")
	}
	if s.storeRepo == nil {
		return fmt.Errorf("store repo is nil")
	}
	err := s.storeRepo.SaveMetric(ctx, metric)
	if err != nil {
		return err
	}
	logger.Log.Debug("store metric is success: ", zap.String("ID", metric.ID))
	return nil
}

func (s *MetricService) storeMetrics(ctx context.Context) error {
	if s.metricsRepo == nil {
		return fmt.Errorf("metric repo is nil")
	}
	if s.storeRepo == nil {
		return fmt.Errorf("store repo is nil")
	}
	metrics, err := s.metricsRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	err = s.storeRepo.SaveMetrics(ctx, metrics)
	if err != nil {
		return err
	}
	logger.Log.Debug("store metrics is success")
	return nil
}
