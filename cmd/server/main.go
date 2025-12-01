package main

import (
	"net/http"

	"github.com/avakumov/metrics/internal/handlers"
	"github.com/avakumov/metrics/internal/repository"
	"github.com/avakumov/metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo)
	metricHandler := handlers.NewMetricHandler(metricService)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", metricHandler.GetAllHandler)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetricHandler)
		r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetricHandler)
	})

	http.ListenAndServe(":8080", r)
}
