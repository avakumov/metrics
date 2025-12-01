package agent

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/utils"
	"github.com/go-resty/resty/v2"
)

// MemStatsCollector собирает и управляет метриками
type MemStatsCollector struct {
	metricsLock sync.Mutex
	metrics     []models.Metric
	count       int
	RandomValue int64
	restyClient *resty.Client
}

// NewMemStatsCollector создает новый сборщик метрик
func NewMemStatsCollector() *MemStatsCollector {
	return &MemStatsCollector{
		restyClient: resty.New(),
	}
}

// Collect собирает все метрики памяти
func (c *MemStatsCollector) Collect() []models.Metric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.metricsLock.Lock()

	c.metrics = []models.Metric{
		{ID: models.Alloc, Value: utils.Float64Ptr(m.Alloc)},
		{ID: "BuckHashSys", Value: utils.Float64Ptr(m.BuckHashSys)},
		{ID: "Frees", Value: utils.Float64Ptr(m.Frees)},
		{ID: "GCCPUFraction", Value: utils.Float64Ptr(m.GCCPUFraction)},
		{ID: "GCSys", Value: utils.Float64Ptr(m.GCSys)},
		{ID: "HeapAlloc", Value: utils.Float64Ptr(m.HeapAlloc)},
		{ID: "HeapIdle", Value: utils.Float64Ptr(m.HeapIdle)},
		{ID: "HeapInuse", Value: utils.Float64Ptr(m.HeapInuse)},
		{ID: "HeapObjects", Value: utils.Float64Ptr(m.HeapObjects)},
		{ID: "HeapReleased", Value: utils.Float64Ptr(m.HeapReleased)},
		{ID: "HeapSys", Value: utils.Float64Ptr(m.HeapSys)},
		{ID: "LastGC", Value: utils.Float64Ptr(m.LastGC)},
		{ID: "Lookups", Value: utils.Float64Ptr(m.Lookups)},
		{ID: "MCacheInuse", Value: utils.Float64Ptr(m.MCacheInuse)},
		{ID: "MCacheSys", Value: utils.Float64Ptr(m.MCacheSys)},
		{ID: "MSpanInuse", Value: utils.Float64Ptr(m.MSpanInuse)},
		{ID: "MSpanSys", Value: utils.Float64Ptr(m.MSpanSys)},
		{ID: "Mallocs", Value: utils.Float64Ptr(m.Mallocs)},
		{ID: "NextGC", Value: utils.Float64Ptr(m.NextGC)},
		{ID: "NumForcedGC", Value: utils.Float64Ptr(m.NumForcedGC)},
		{ID: "NumGC", Value: utils.Float64Ptr(m.NumGC)},
		{ID: "OtherSys", Value: utils.Float64Ptr(m.OtherSys)},
		{ID: "PauseTotalNs", Value: utils.Float64Ptr(m.PauseTotalNs)},
		{ID: "StackInuse", Value: utils.Float64Ptr(m.StackInuse)},
		{ID: "StackSys", Value: utils.Float64Ptr(m.StackSys)},
		{ID: "Sys", Value: utils.Float64Ptr(m.Sys)},
		{ID: "TotalAlloc", Value: utils.Float64Ptr(m.TotalAlloc)},
	}
	c.RandomValue = rand.Int64()
	c.count++

	c.metricsLock.Unlock()
	return c.metrics
}

func (c *MemStatsCollector) SendMetrics() {
	c.metricsLock.Lock()

	client := c.restyClient

	for _, metric := range c.metrics {
		params := map[string]string{
			"typeMetric":  "gauge",
			"metricID":    metric.ID,
			"metricValue": fmt.Sprintf("%f", *metric.Value),
		}

		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			SetPathParams(params).
			Post("http://localhost:8080/update/{typeMetric}/{metricID}/{metricValue}")

		if err != nil {
			fmt.Printf("▶️  REQUEST ERROR: %v\n", err)
		}
		fmt.Printf("url: %s, code: %d\n", resp.Request.URL, resp.StatusCode())
	}

	c.metricsLock.Unlock()
}
