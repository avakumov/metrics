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

	//fmt.Printf("metric name: %s\n", metricName)
	//fmt.Printf("metric type: %s\n", metricType)

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
		io.WriteString(rw, strconv.FormatFloat(*metric.Value, 'f', -1, 64))
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
			//remove zeros
			valueWithoutZeros := strconv.FormatFloat(*m.Value, 'f', -1, 64)
			list = append(list, fmt.Sprintf("%s = %s", m.ID, valueWithoutZeros))
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
		fmt.Printf("Counter: %s = %d \n", metricValue, value)

	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "invalid guege value", http.StatusBadRequest)
			return
		}
		fmt.Printf("Gauge: %s = %g\n", metricName, value)
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

	h.metricService.SaveMetric(models.Metric{
		ID:    metricName,
		MType: metricType,
		Value: &value,
	})

	w.WriteHeader(http.StatusOK)

}
