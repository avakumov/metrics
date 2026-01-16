package main

import (
	"net/http"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/server/config"
	"github.com/avakumov/metrics/internal/server/database"
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

	//database init
	db, err := database.Connect(options.DSN)
	if err != nil {
		logger.Log.Error("failed to connect db", zap.String("DSN", options.DSN), zap.Error(err))
	}

	//apply migrations
	err = database.RunMigrations(db.SQLDB)
	if err != nil {
		logger.Log.Error("failed migrations db", zap.Error(err))
	}

	metricsRepo := repository.NewMemoryRepository()
	//restore from file
	if options.Restore {
		metricsRepo.Restore(options.FileStoragePath)
	}
	metricService := service.NewMetricService(metricsRepo, options.FileStoragePath, options.StoreInterval, db)
	metricService.Init()
	metricHandler := handlers.NewMetricsHandler(metricService)
	err = http.ListenAndServe(options.Address, router.MetricsRouter(metricHandler))
	if err != nil {
		logger.Log.Error("failed to start metrics app server", zap.String("address", options.Address), zap.Error(err))
	}
}
