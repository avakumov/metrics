package router

import (
	"github.com/avakumov/metrics/internal/server/handlers"
	"github.com/avakumov/metrics/internal/server/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func MetricsRouter(metricHandler *handlers.MetricHandler) chi.Router {

	r := chi.NewRouter()

	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", logger.WithLogging(metricHandler.GetAllHandler))
		r.Post("/update/{metricType}/{metricName}/{metricValue}", logger.WithLogging(metricHandler.UpdateMetricHandler))
		//для того чтобы отловить пустое значение метрики
		r.Post("/update/{metricType}/", logger.WithLogging(metricHandler.UpdateMetricHandler))
		r.Get("/value/{metricType}/{metricName}", logger.WithLogging(metricHandler.GetMetricHandler))

		//на все не найденные отвечать кодом 400
		r.NotFound(logger.WithLogging(metricHandler.NotFoundHandler))
	})
	return r
}
