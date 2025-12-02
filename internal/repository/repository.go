package repository

import (
	"github.com/avakumov/metrics/internal/models"
)

type Repository interface {
	GetMetricByID(id string) (models.Metric, error)
	SaveMetric(metric models.Metric) error
	DeleteMetricByID(id string) error
	FindAll() ([]models.Metric, error)
}
