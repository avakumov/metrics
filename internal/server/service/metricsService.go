package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/database"
	"github.com/avakumov/metrics/internal/server/repository"
	"go.uber.org/zap"
)

type MetricService struct {
	metricsRepo   repository.Repository
	storeInterval int
	storeFilepath string
	DB            *database.Database
}

func NewMetricService(repo repository.Repository, storeFilepath string, storeInterval int, db *database.Database) MetricService {
	return MetricService{metricsRepo: repo, storeInterval: storeInterval, storeFilepath: storeFilepath, DB: db}
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
	//сохраняем синхронно в файл если не задан интервал сохранения
	if s.storeInterval == 0 {
		err = s.store()
		if err != nil {
			return err
		}
	}

	return nil
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
		err := s.store()
		if err != nil {
			logger.Log.Error("store data error:", zap.Error(err))
		}

	}
}

func (s *MetricService) store() error {
	metrics, err := s.metricsRepo.FindAll()
	if err != nil {
		return err
	}
	//store in database
	if s.DB != nil && s.DB.Pool != nil {
		err = s.storeInDB(metrics)
		if err != nil {
			return err
		}
		return nil
	}
	//store in file
	if s.storeFilepath != "" {
		err = s.storeInFile(metrics)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("store not executed")
}
func (s *MetricService) storeInFile(metrics []models.Metric) error {

	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}
	err = os.WriteFile(s.storeFilepath, data, 0666)
	if err != nil {
		return err
	}
	logger.Log.Debug("Auto-save in file completed")
	return nil
}

func (s *MetricService) storeInDB(metrics []models.Metric) error {
	query := `
	INSERT INTO metrics (id, m_type, delta, value)
	VALUES ($1, $2, $3, $4)
  ON CONFLICT (id, m_type) 
  DO UPDATE SET 
    delta = EXCLUDED.delta,
    value = EXCLUDED.value,
    hash = EXCLUDED.hash,
    updated_at = CURRENT_TIMESTAMP
	`
	pool := s.DB.Pool
	ctx := context.Background()
	for _, metric := range metrics {

		_, err := pool.Exec(ctx, query, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err != nil {
			return fmt.Errorf("failed to save metric %s: %w", metric.ID, err)
		}
	}

	logger.Log.Debug("Auto-saving metrics in DB completed")
	return nil
}
