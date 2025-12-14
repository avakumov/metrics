package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"sort"

	"strconv"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type MetricHandler struct {
	metricService service.MetricService
}

type Metric struct {
	Name  string
	Value string
}
type PageData struct {
	Metrics []Metric
	Header  string
	Title   string
}

func NewMetricsHandler(metricService service.MetricService) *MetricHandler {
	return &MetricHandler{metricService: metricService}
}

func (h *MetricHandler) GetMetric(rw http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metric, err := h.metricService.GetMetric(metricName)
	if err != nil {
		http.Error(rw, "Not found metric", http.StatusNotFound)
		return
	}

	if metricType != metric.MType {
		http.Error(rw, "Wrong type", http.StatusNotFound)
		return
	}
	rw.WriteHeader(http.StatusOK)
	if metric.MType == models.Gauge {
		metricString := strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		_, err = io.WriteString(rw, metricString)
		if err != nil {
			logger.Log.Error("write response error", zap.String("metricValue", metricString), zap.Error(err))
			http.Error(rw, "write error", http.StatusInternalServerError)
			return
		}

	}
	if metric.MType == models.Counter {
		metricString := strconv.FormatInt(*metric.Delta, 10)
		_, err = io.WriteString(rw, metricString)
		if err != nil {
			logger.Log.Error("write response error", zap.String("metricValue", metricString), zap.Error(err))
			http.Error(rw, "write error", http.StatusInternalServerError)
			return
		}

	}
}

func (h *MetricHandler) GetAll(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	metrics, err := h.metricService.GetAllMetric()
	if err != nil {
		http.Error(rw, "Not found metrics", http.StatusNotFound)
		return
	}

	data := PageData{
		Header: "Metrics",
		Title:  "All metrics",
	}
	for _, m := range metrics {
		if m.MType == models.Gauge {
			data.Metrics = append(data.Metrics, Metric{
				Name:  m.ID,
				Value: strconv.FormatFloat(*m.Value, 'f', -1, 64),
			})
		}
		if m.MType == models.Counter {
			data.Metrics = append(data.Metrics, Metric{
				Name:  m.ID,
				Value: strconv.FormatInt(*m.Delta, 10),
			})
		}
	}

	sort.Slice(data.Metrics, func(i, j int) bool {
		return data.Metrics[i].Name < data.Metrics[j].Name
	})
	tmpl := template.Must(template.ParseFiles("../../internal/server/templates/allMetrics.html"))
	//Execute добавляет статус 200
	err = tmpl.Execute(rw, data)
	if err != nil {
		logger.Log.Error("error write data in parsed html file", zap.Error(err))
		http.Error(rw, "template error", http.StatusInternalServerError)
		return
	}
}

func (h *MetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {

	metric, err := getMetricFromRequest(r)
	logger.Log.Sugar().Debugf("METRIC: %+v\n", metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if metric.ID == "" {
		http.Error(w, "Not found type, name, value", http.StatusNotFound)
		return
	}

	if metric.Value == nil && metric.MType == models.Gauge {
		http.Error(w, "requered value not nil", http.StatusBadRequest)
		return
	}

	if metric.Delta == nil && metric.MType == models.Counter {
		http.Error(w, "requered delta not nil", http.StatusBadRequest)
		return
	}

	err = h.metricService.SaveMetric(metric)
	if err != nil {
		logger.Log.Error("error on save metric", zap.Error(err))
		http.Error(w, "error on save metric", http.StatusInternalServerError)
		return
	}
	logger.Log.Debug("update metric",
		zap.String("ID", metric.ID),
		zap.String("Type", metric.MType),
		//zap.Int64("Value", *metric.Delta),
		//zap.Float64("Value", *metric.Value),
	)

	w.WriteHeader(http.StatusOK)

}

func (h *MetricHandler) GetMetricValues(w http.ResponseWriter, r *http.Request) {
	sM := models.Metric{}
	if err := json.NewDecoder(r.Body).Decode(&sM); err != nil {
		http.Error(w, "parsing json error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	logger.Log.Sugar().Debugf("receive metrics: %+v", sM)

	//получаем метрику из хранилаща
	metric, err := h.metricService.GetMetric(sM.ID)
	if err != nil {
		http.Error(w, "not found metric", http.StatusNotFound)
		logger.Log.Warn("get metric by id", zap.Error(err))
		return
	}
	if metric.MType != sM.MType {
		http.Error(w, "not found metric", http.StatusNotFound)
		return
	}
	// Устанавливаем заголовки и отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		logger.Log.Error("failed to encode response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusBadRequest)
		return
	}
}

func (h *MetricHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound) // 404
}
func (h *MetricHandler) BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest) // 400
}

// для извлеыения обновления метрики из тела запроса json или из url
func getMetricFromRequest(r *http.Request) (models.Metric, error) {
	metric := models.Metric{}

	//метрика из json
	contentLen := r.ContentLength
	if contentLen > 0 {
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			return metric, err
		}
		logger.Log.Debug("length body is ", zap.Int("body length", int(contentLen)))
		logger.Log.Sugar().Debugf("metric recived %+v", metric)
		return metric, nil
	}
	//метрика из url
	metricName := chi.URLParam(r, "metricName")
	metricType := chi.URLParam(r, "metricType")
	metricValue := chi.URLParam(r, "metricValue")
	switch metricType {
	case models.Counter:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		metric.Value = nil
		metric.Delta = &delta
		metric.MType = models.Counter
		metric.ID = metricName
		if err != nil {
			return metric, err
		}
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		metric.Value = &value
		metric.Delta = nil
		metric.MType = models.Gauge
		metric.ID = metricName
		if err != nil {
			return metric, err
		}
	default:
		return metric, errors.New("error type of metric: counter, gauge")
	}
	return metric, nil

}
