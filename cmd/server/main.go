package main

import (
	"fmt"
	"net/http"

	router "github.com/avakumov/metrics/internal/server"
	"github.com/avakumov/metrics/internal/server/config"
)

func main() {
	options := config.GetOptions()
	port := fmt.Sprintf(":%d", options.Port)
	http.ListenAndServe(port, router.MetricsRouter())
}
