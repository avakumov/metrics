package main

import (
	"net/http"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/server/config"
	"github.com/avakumov/metrics/internal/server/handlers"
	"github.com/avakumov/metrics/internal/server/repository"
	"github.com/avakumov/metrics/internal/server/router"
	"github.com/avakumov/metrics/internal/server/service"
	"go.uber.org/zap"
)

func main() {
	options := config.GetOptions()

	logger.Init(options.LogLevel, "server")
	defer logger.Log.Sync() //nolint:errcheck
	logger.Log.Sugar().Infof("START OPTIONS: %+v", options)

	memoRepo := repository.NewMemoryRepository()

	//репозиторий в который будут сохраняться данные
	storeRepository, err := getStore(options)
	if err != nil {
		logger.Log.Error("failed to create store repository:", zap.Any("Options", options), zap.Error(err))
	}

	metricService := service.NewMetricService(memoRepo, storeRepository, options)
	metricService.Init()
	metricHandler := handlers.NewMetricsHandler(metricService)
	err = http.ListenAndServe(options.Address, router.MetricsRouter(metricHandler))
	if err != nil {
		logger.Log.Error("failed to start metrics app server", zap.String("address", options.Address), zap.Error(err))
	}
}

func getStore(options config.Options) (repository.Repository, error) {

	if options.DSN != "" {
		return repository.NewDBRepository(options.DSN)
	}
	if options.FileStoragePath != "" {
		return repository.NewFileRepository(options.FileStoragePath)
	}
	return nil, nil
}
