package handlers

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"strconv"

	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/service"
	"github.com/go-chi/chi/v5"
)

type MetricHandler struct {
	metricService service.MetricService
}

func NewMetricHandler(metricService service.MetricService) *MetricHandler {
	return &MetricHandler{metricService: metricService}
}

func (h *MetricHandler) GetMetricHandler(rw http.ResponseWriter, r *http.Request) {
	metricType := strings.ToLower(chi.URLParam(r, "metricType"))
	metricName := strings.ToLower(chi.URLParam(r, "metricName"))

	fmt.Printf("metric name: %s\n", metricName)
	fmt.Printf("metric type: %s\n", metricType)

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
	switch metric.MType {
	case "gauge":
		io.WriteString(rw, fmt.Sprintf("%f", *metric.Value))
	case "counter":
		io.WriteString(rw, fmt.Sprintf("%d", int64(*metric.Value)))

	}
}

func (h *MetricHandler) GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricService.GetAllMetric()
	if err != nil {
		http.Error(rw, "Not found metrics", http.StatusNotFound)
		return
	}

	list := make([]string, 0)
	for _, m := range metrics {
		if m.MType == "counter" {
			list = append(list, fmt.Sprintf("%s = %d", m.ID, int64(*m.Value)))
		}
		if m.MType == "gauge" {
			list = append(list, fmt.Sprintf("%s = %f", m.ID, *m.Value))
		}
	}

	sort.Strings(list)

	for i, l := range list {
		list[i] = "<div>" + l + "</div>"
	}
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(rw, strings.Join(list, "\n"))
}

func (h *MetricHandler) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	if http.MethodPost != r.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	fmt.Println(r.URL.Path)

	if len(pathParts) < 3 {
		http.Error(w, "wrong path", http.StatusBadRequest)
		return
	}

	if pathParts[1] != "update" {
		http.Error(w, "Invalid path format: use first /update/", http.StatusBadRequest)
		return
	}

	if pathParts[2] != "counter" && pathParts[2] != "gauge" {
		http.Error(w, "Invalid path format: use first /update/", http.StatusBadRequest)
		return
	}

	if len(pathParts) == 4 && pathParts[3] == "" {
		http.Error(w, "Not found type, name, value", http.StatusNotFound)
		return
	}

	if len(pathParts) != 5 {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	metricType := strings.ToLower(pathParts[2])
	metricName := strings.ToLower(pathParts[3])
	metricValue := strings.ToLower(pathParts[4])

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "invalid counter value", http.StatusBadRequest)
			return
		}
		fmt.Printf("Counter: %s = %d \n", metricValue, value)

	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "invalid guege value", http.StatusBadRequest)
			return
		}
		fmt.Printf("Gauge: %s = %f\n", metricName, value)
	default:
		http.Error(w, "unknown metric type", http.StatusBadRequest)
		return
	}

	//TODO переписать конвертер
	value, _ := strconv.ParseFloat(metricValue, 64)

	h.metricService.SaveMetric(models.Metric{
		ID:    metricName,
		MType: metricType,
		Value: &value,
		// Delta *int64   `json:"delta,omitempty"`
		// Hash  string   `json:"hash,omitempty"`
	})

	mets, err := h.metricService.GetAllMetric()
	if err != nil {
		fmt.Printf("Не могу прочитать данные %s\n", metricName)
	}
	for _, met := range mets {
		fmt.Printf("Cохранена метрика %v\n", met)
	}

	w.WriteHeader(http.StatusOK)

}
