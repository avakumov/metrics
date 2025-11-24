package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateMetricHandler)
	http.ListenAndServe(":8080", mux)
}

func updateMetricHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf(" contentType: %s\n", r.Header.Get("Content-Type"))
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
	for i, part := range pathParts {
		fmt.Printf("part %d is %s\n", i, part)
	}

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

	if len(pathParts) == 3 {
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
	err := processMetric(metricType, metricName, metricValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func processMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value: %s", metricValue)
		}
		fmt.Printf("Counter: %s = %d \n", metricValue, value)

	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value: %s", metricValue)
		}
		fmt.Printf("Gauge: %s = %f\n", metricName, value)
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
	return nil
}
