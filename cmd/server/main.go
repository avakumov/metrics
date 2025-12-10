package main

import (
	"log"
	"net/http"

	"github.com/avakumov/metrics/internal/server/config"
	"github.com/avakumov/metrics/internal/server/handlers"
	"github.com/avakumov/metrics/internal/server/logger"
	"github.com/avakumov/metrics/internal/server/repository"
	"github.com/avakumov/metrics/internal/server/router"
	"github.com/avakumov/metrics/internal/server/service"
)

func main() {
	options := config.GetOptions()
	logger.InitLogger()

	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo)
	metricHandler := handlers.NewMetricHandler(metricService)

	err := http.ListenAndServe(options.Address, router.MetricsRouter(metricHandler))
	if err != nil {
		log.Printf("Failed to start metrics server on address %s: %v", options.Address, err)
	}
}
