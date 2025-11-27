package handlers

import (
	"github.com/avakumov/metrics/internal/utils"
	"fmt"
	"net/http"
	"strings"
)

func UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
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

	metricType := pathParts[2]
	metricName := pathParts[3]
	metricValue := pathParts[4]

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := utils.ParseMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)

}
