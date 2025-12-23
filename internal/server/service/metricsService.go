package service

import (
	"encoding/json"
	"os"
	"time"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/repository"
	"go.uber.org/zap"
)

type MetricService struct {
	metricsRepo   repository.Repository
	storeInterval int
	storeFilepath string
}

func NewMetricService(repo repository.Repository, storeFilepath string, storeInterval int) MetricService {
	return MetricService{metricsRepo: repo, storeInterval: storeInterval, storeFilepath: storeFilepath}
}

func (s *MetricService) Init() {
	if s.storeInterval > 0 {
		go s.saveMetricsWithPeriod()
	}
}

func (s *MetricService) SaveMetric(metric models.Metric) error {
	err := s.metricsRepo.SaveMetric(metric)
	if err != nil {
		return err
	}

	err = s.saveMetricInFile()
	if err != nil {
		return err
	}
	return nil
	//save metrics to file
}

func (s *MetricService) GetMetric(id string) (models.Metric, error) {
	return s.metricsRepo.GetMetricByID(id)
}

func (s *MetricService) GetAllMetric() ([]models.Metric, error) {
	return s.metricsRepo.FindAll()
}

func (s *MetricService) RemoveMetric(id string) error {
	return s.metricsRepo.DeleteMetricByID(id)
}

func (s *MetricService) saveMetricsWithPeriod() {

	ticker := time.NewTicker(time.Duration(s.storeInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := s.saveMetricInFile()
		if err != nil {
			logger.Log.Error("error on save to file", zap.Error(err))
		}
	}
}

func (s *MetricService) saveMetricInFile() error {

	logger.Log.Debug("Auto-saving metrics...")
	metrics, err := s.metricsRepo.FindAll()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}
	err = os.WriteFile(s.storeFilepath, data, 0666)
	if err != nil {
		return err
	}
	logger.Log.Debug("Auto-save completed")
	return nil
}
