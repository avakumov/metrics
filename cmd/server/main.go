package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	mainContext, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	memoRepo := repository.NewMemoryRepository()

	//репозиторий в который будут сохраняться данные
	storeRepository, err := getStore(options)
	if err != nil {
		logger.Log.Error("failed to create store repository:", zap.Any("Options", options), zap.Error(err))
	}

	metricService := service.NewMetricService(mainContext, memoRepo, storeRepository, options)
	metricService.Init()
	metricHandler := handlers.NewMetricsHandler(metricService)
	server := http.Server{
		Addr:    options.Address,
		Handler: router.MetricsRouter(metricHandler),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Log.Error("failed to start metrics app server", zap.String("address", options.Address), zap.Error(err))
			mainCancel()
		}
	}()

	<-stop
	logger.Log.Info("Shutdown signal received")
	mainCancel()
	time.Sleep(2 * time.Second)

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
