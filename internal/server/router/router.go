package router

import (
	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/server/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func MetricsRouter(metricHandler *handlers.MetricHandler) chi.Router {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.LoggerMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Get("/", metricHandler.GetAll)
		r.Post("/update", metricHandler.UpdateMetric)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)
		r.Post("/update/gauge/", metricHandler.NotFound)
		r.Post("/update/counter/", metricHandler.NotFound)
		r.Post("/value", metricHandler.GetMetricValues)
		r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetric)

		//на все не найденные отвечать кодом 400
		r.NotFound(metricHandler.BadRequest)
	})
	return r
}
