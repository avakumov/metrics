package handlers

import (
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

func NewMetricHandler(metricService service.MetricService) *MetricHandler {
	return &MetricHandler{metricService: metricService}
}

func (h *MetricHandler) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {

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

func (h *MetricHandler) GetAllHandler(rw http.ResponseWriter, r *http.Request) {
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

func (h *MetricHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {

	metricType := strings.ToLower(chi.URLParam(r, "metricType"))
	metricName := strings.ToLower(chi.URLParam(r, "metricName"))
	metricValue := strings.ToLower(chi.URLParam(r, "metricValue"))

	if metricType != models.Counter && metricType != models.Gauge {
		http.Error(w, "Metric types is gauge or counter", http.StatusBadRequest)
		return
	}

	if metricName == "" {
		http.Error(w, "Not found type, name, value", http.StatusNotFound)
		return
	}

	if metricValue == "" {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}

	switch metricType {
	case "counter":
		_, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "invalid counter value", http.StatusBadRequest)
			return
		}

	case "gauge":
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "invalid guege value", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	value, _ := strconv.ParseFloat(metricValue, 64)
	if metricType == models.Counter {
		existedMetric, err := h.metricService.GetMetric(metricName)
		if err == nil {
			value += *existedMetric.Value
		}

	}
	metric := models.Metric{
		ID:    metricName,
		MType: metricType,
		Value: &value,
	}
	err := h.metricService.SaveMetric(metric)
	if err != nil {
		logger.Log.Error("error on save metric", zap.Error(err))
	} else {
		logger.Log.Debug("update metric",
			zap.String("ID", metricName),
			zap.String("Type", metricType),
			zap.Float64("Value", value))
	}

	w.WriteHeader(http.StatusOK)

}

func (h *MetricHandler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest) // 400
}
