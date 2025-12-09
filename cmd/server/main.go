package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/avakumov/metrics/internal/handlers"
	"github.com/avakumov/metrics/internal/repository"
	router "github.com/avakumov/metrics/internal/server"
	"github.com/avakumov/metrics/internal/server/config"
	"github.com/avakumov/metrics/internal/service"
)

func main() {
	options := config.GetOptions()
	port := fmt.Sprintf(":%d", options.Port)

	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo)
	metricHandler := handlers.NewMetricHandler(metricService)

	err := http.ListenAndServe(port, router.MetricsRouter(metricHandler))
	if err != nil {
		log.Printf("Failed to start metrics server on port %s: %v", port, err)
	}
}
