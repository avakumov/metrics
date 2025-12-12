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
		r.Get("/", metricHandler.GetAllHandler)
		r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetricHandler)
		//для того чтобы отловить пустое значение метрики
		r.Post("/update/{metricType}/", metricHandler.UpdateMetricHandler)
		r.Get("/value/{metricType}/{metricName}", metricHandler.GetMetricHandler)

		//на все не найденные отвечать кодом 400
		r.NotFound(metricHandler.NotFoundHandler)
	})
	return r
}
