package service

import (
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/repository"
)

type MetricService struct {
	metricsRepo repository.Repository
}

func NewMetricService(repo repository.Repository) MetricService {
	return MetricService{metricsRepo: repo}
}

func (s *MetricService) SaveMetric(metric models.Metric) error {
	return s.metricsRepo.SaveMetric(metric)
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
