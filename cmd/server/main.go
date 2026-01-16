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
	var DB *database.Database

	logger.Init(options.LogLevel, "server")
	defer logger.Log.Sync() //nolint:errcheck
	logger.Log.Sugar().Infof("START OPTIONS: %+v", options)

	metricsRepo := repository.NewMemoryRepository()
	//without database
	if options.DSN == "" {
		logger.Log.Info("DSN is empty string. Start without DB")

		if options.Restore {
			//restore from file
			err := metricsRepo.RestoreFromFile(options.FileStoragePath)
			if err != nil {
				logger.Log.Error("failed to restore data from file", zap.Error(err))
			}
		}
		//with database
	} else {
		//database init
		db, err := database.Connect(options.DSN)
		if err != nil {
			logger.Log.Error("failed to connect db", zap.String("DSN", options.DSN), zap.Error(err))
		}
		DB = db
		//apply migrations
		err = database.RunMigrations(db.SQLDB)
		if err != nil {
			logger.Log.Error("failed to apply migrations db", zap.Error(err))
		}
		if options.Restore {

			//restore from db
			if db != nil && db.Pool != nil {
				err := metricsRepo.RestoreFromDB(db.Pool)
				if err != nil {
					logger.Log.Error("failed to restore data from database", zap.Error(err))
				}
			}
		}
	}

	metricService := service.NewMetricService(metricsRepo, options.FileStoragePath, options.StoreInterval, DB)
	metricService.Init()
	metricHandler := handlers.NewMetricsHandler(metricService)
	err := http.ListenAndServe(options.Address, router.MetricsRouter(metricHandler))
	if err != nil {
		logger.Log.Error("failed to start metrics app server", zap.String("address", options.Address), zap.Error(err))
	}
}
