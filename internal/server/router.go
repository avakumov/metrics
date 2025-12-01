package router

import (
	"github.com/avakumov/metrics/internal/handlers"
	"github.com/avakumov/metrics/internal/repository"
	"github.com/avakumov/metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func MetricsRouter() chi.Router {

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
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest) // 400
			w.Write([]byte("Bad Request"))
		})
	})
	return r
}
