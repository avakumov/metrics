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

	metricsRepo := repository.NewMemoryRepository()
	if options.Restore {
		metricsRepo.Restore(options.FileStoragePath)
	}
	metricService := service.NewMetricService(metricsRepo, options.FileStoragePath, options.StoreInterval)
	metricService.Init()
	metricHandler := handlers.NewMetricsHandler(metricService)
	err := http.ListenAndServe(options.Address, router.MetricsRouter(metricHandler))
	if err != nil {
		logger.Log.Error("failed to start metrics app server", zap.String("address", options.Address), zap.Error(err))
	}
}
