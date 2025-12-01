package repository

import (
	"github.com/avakumov/metrics/internal/models"
)

type Repository interface {
	GetMetricById(id string) (models.Metric, error)
	SaveMetric(metric models.Metric) error
	DeleteMetricById(id string) error
	FindAll() ([]models.Metric, error)
}
