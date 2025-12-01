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
	"github.com/avakumov/metrics/internal/utils"
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
	io.WriteString(rw, fmt.Sprintf("%f", *metric.Value))
}

func (h *MetricHandler) GetAllHandler(rw http.ResponseWriter, r *http.Request) {
	metrics, err := h.metricService.GetAllMetric()
	if err != nil {
		http.Error(rw, "Not found metrics", http.StatusNotFound)
		return
	}

	list := make([]string, 0)
	for _, m := range metrics {
		line := fmt.Sprintf("%s = %f", m.Id, *m.Value)
		list = append(list, line)
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
	//	fmt.Println(r.URL.Path)

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
	err := utils.CheckMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//TODO переписать конвертер
	value, err := strconv.ParseFloat(metricValue, 64)

	h.metricService.SaveMetric(models.Metric{
		Id:    metricName,
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
