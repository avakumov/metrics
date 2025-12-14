package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strings"

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

	metricType := strings.ToLower(chi.URLParam(r, "metricType"))
	metricName := strings.ToLower(chi.URLParam(r, "metricName"))

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
	metricString := strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	_, err = io.WriteString(rw, metricString)
	if err != nil {
		logger.Log.Error("write response error", zap.String("metricValue", metricString), zap.Error(err))
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

		data.Metrics = append(data.Metrics, Metric{
			Name:  m.ID,
			Value: strconv.FormatFloat(*m.Value, 'f', -1, 64),
		})
	}

	sort.Slice(data.Metrics, func(i, j int) bool {
		return data.Metrics[i].Name < data.Metrics[j].Name
	})

	tmpl := template.Must(template.ParseFiles("../../internal/server/templates/allMetrics.html"))
	//Execute добавляет статус 200
	err = tmpl.Execute(rw, data)
	if err != nil {
		logger.Log.Error("error write data in parsed html file", zap.Error(err))
	}
}

func (h *MetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {

	metric, err := getMetricFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if metric.MType != models.Counter && metric.MType != models.Gauge {
		http.Error(w, "Metric types is gauge or counter", http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		http.Error(w, "Not found type, name, value", http.StatusNotFound)
		return
	}

	if metric.MType == models.Counter {
		existedMetric, err := h.metricService.GetMetric(metric.ID)
		if err == nil {
			*metric.Value += *existedMetric.Value
		}

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
		zap.Float64("Value", *metric.Value))

	w.WriteHeader(http.StatusOK)

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

	contentLen := r.ContentLength
	if contentLen > 0 {
		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			return metric, errors.New("parsing error")
		}
		logger.Log.Debug("length body is ", zap.Int("body length", int(contentLen)))
		logger.Log.Sugar().Debugf("metric recived %+v", metric)
		logger.Log.Sugar().Debugf("value %f", *metric.Value)
	} else {
		metricName := strings.ToLower(chi.URLParam(r, "metricName"))
		metricType := strings.ToLower(chi.URLParam(r, "metricType"))
		metricValue := strings.ToLower(chi.URLParam(r, "metricValue"))

		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return metric, errors.New("invalid parse digit value")
		}
		metric.ID = metricName
		metric.MType = metricType
		metric.Value = &value
	}
	return metric, nil

}
