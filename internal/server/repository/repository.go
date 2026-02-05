package repository

import (
	"context"

	"github.com/avakumov/metrics/internal/models"
)

type Repository interface {
	GetMetricByID(ctx context.Context, id string) (*models.Metric, error)
	GetAll(ctx context.Context) ([]models.Metric, error)
	SaveMetric(ctx context.Context, metric models.Metric) error
	SaveMetrics(ctx context.Context, metrics []models.Metric) error
	DeleteMetricByID(ctx context.Context, id string) error
	Ping(ctx context.Context) error
	Close()
}
