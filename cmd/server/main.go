package main

import (
	"net/http"
	"github.com/avakumov/metrics/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.UpdateMetricHandler)
	http.ListenAndServe(":8080", mux)
}
