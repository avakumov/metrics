package main

import (
	"net/http"

	router "github.com/avakumov/metrics/internal/server"
)

func main() {
	http.ListenAndServe(":8080", router.MetricsRouter())
}
