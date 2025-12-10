package handlers

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"strconv"

	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/server/service"
	"github.com/go-chi/chi/v5"
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

	_, err = io.WriteString(rw, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
	if err != nil {
		log.Printf("error WriteString in response %f", *metric.Value)
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
		log.Printf("error write data in parsed doc")
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
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "invalid counter value", http.StatusBadRequest)
			return
		}
		log.Printf("Counter: %s = %d \n", metricValue, value)

	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "invalid guege value", http.StatusBadRequest)
			return
		}
		log.Printf("Gauge: %s = %g\n", metricName, value)
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
		log.Printf("error on save metric: %+v", metric)
	}

	w.WriteHeader(http.StatusOK)

}

func (h *MetricHandler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest) // 400
}
