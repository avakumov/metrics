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
	logger.Init()
	defer logger.Log.Sync()

	options := config.GetOptions()

	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo)
	metricHandler := handlers.NewMetricHandler(metricService)
	logger.Log.Info("metrics server app starting", zap.String("address", options.Address))
	err := http.ListenAndServe(options.Address, router.MetricsRouter(metricHandler))
	if err != nil {
		logger.Log.Error("failed to start metrics app server", zap.String("address", options.Address), zap.Error(err))
	}
}
