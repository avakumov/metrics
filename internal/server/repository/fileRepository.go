package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
)

type FileRepository struct {
	file *os.File
	mu   sync.Mutex
}

func (fr *FileRepository) Close() {

}

func NewFileRepository(filepath string) (*FileRepository, error) {
	logger.Log.Debug("starting new file repository")
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}
	return &FileRepository{file: file}, nil
}

func (fr *FileRepository) GetAll(ctx context.Context) ([]models.Metric, error) {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	// ВСЕГДА переходим в начало файла перед чтением!
	if _, err := fr.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek error: %v", err)
	}

	file := fr.file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	// Если файл пустой, возвращаем пустой массив
	if len(data) == 0 {
		return []models.Metric{}, nil
	}
	var metrics []models.Metric
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (fr *FileRepository) GetMetricByID(ctx context.Context, id string) (*models.Metric, error) {
	metrics, err := fr.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range metrics {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("not found metric with id %s", id)
}

func (fr *FileRepository) SaveMetric(ctx context.Context, metric models.Metric) error {
	metrics, err := fr.GetAll(ctx)
	if err != nil {
		return err
	}
	isExists := false
	for i, m := range metrics {
		if m.ID == metric.ID {
			metrics[i] = metric
			isExists = true
		}
	}
	if !isExists {
		metrics = append(metrics, metric)
	}
	return fr.write(metrics)
}

func (fr *FileRepository) SaveMetrics(ctx context.Context, metrics []models.Metric) error {
	for _, m := range metrics {
		err := fr.SaveMetric(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fr *FileRepository) DeleteMetricByID(ctx context.Context, id string) error {

	metrics, err := fr.GetAll(ctx)
	if err != nil {
		return err
	}
	for i, m := range metrics {
		if m.ID == id {
			metrics[i] = metrics[len(metrics)-1]
			metrics = metrics[:len(metrics)-1]
		}
	}
	return fr.write(metrics)
}

func (fr *FileRepository) write(metrics []models.Metric) error {
	data, err := json.MarshalIndent(metrics, "", "   ")
	if err != nil {
		return err
	}

	fr.mu.Lock()
	defer fr.mu.Unlock()
	// 1. Обрезаем файл
	err = fr.file.Truncate(0)
	if err != nil {
		return err
	}
	// 2. Переходим в начало
	if _, err := fr.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	_, err = fr.file.Write(data)
	if err != nil {
		return err
	}
	return nil

}

func (fr *FileRepository) Ping(ctx context.Context) error {
	return fmt.Errorf("ping is not available. The file repository is being used")
}
