package repository

import (
	"context"

	"github.com/avakumov/metrics/internal/models"
)

type Repository interface {
	GetMetricByID(id string) (*models.Metric, error)
	GetAll() ([]models.Metric, error)
	SaveMetric(metric models.Metric) error
	SaveMetrics(metrics []models.Metric) error
	DeleteMetricByID(id string) error
	Ping(ctx context.Context) error
	Close()
}
