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
		{Id: models.Alloc, Value: utils.Float64Ptr(m.Alloc)},
		{Id: "BuckHashSys", Value: utils.Float64Ptr(m.BuckHashSys)},
		{Id: "Frees", Value: utils.Float64Ptr(m.Frees)},
		{Id: "GCCPUFraction", Value: utils.Float64Ptr(m.GCCPUFraction)},
		{Id: "GCSys", Value: utils.Float64Ptr(m.GCSys)},
		{Id: "HeapAlloc", Value: utils.Float64Ptr(m.HeapAlloc)},
		{Id: "HeapIdle", Value: utils.Float64Ptr(m.HeapIdle)},
		{Id: "HeapInuse", Value: utils.Float64Ptr(m.HeapInuse)},
		{Id: "HeapObjects", Value: utils.Float64Ptr(m.HeapObjects)},
		{Id: "HeapReleased", Value: utils.Float64Ptr(m.HeapReleased)},
		{Id: "HeapSys", Value: utils.Float64Ptr(m.HeapSys)},
		{Id: "LastGC", Value: utils.Float64Ptr(m.LastGC)},
		{Id: "Lookups", Value: utils.Float64Ptr(m.Lookups)},
		{Id: "MCacheInuse", Value: utils.Float64Ptr(m.MCacheInuse)},
		{Id: "MCacheSys", Value: utils.Float64Ptr(m.MCacheSys)},
		{Id: "MSpanInuse", Value: utils.Float64Ptr(m.MSpanInuse)},
		{Id: "MSpanSys", Value: utils.Float64Ptr(m.MSpanSys)},
		{Id: "Mallocs", Value: utils.Float64Ptr(m.Mallocs)},
		{Id: "NextGC", Value: utils.Float64Ptr(m.NextGC)},
		{Id: "NumForcedGC", Value: utils.Float64Ptr(m.NumForcedGC)},
		{Id: "NumGC", Value: utils.Float64Ptr(m.NumGC)},
		{Id: "OtherSys", Value: utils.Float64Ptr(m.OtherSys)},
		{Id: "PauseTotalNs", Value: utils.Float64Ptr(m.PauseTotalNs)},
		{Id: "StackInuse", Value: utils.Float64Ptr(m.StackInuse)},
		{Id: "StackSys", Value: utils.Float64Ptr(m.StackSys)},
		{Id: "Sys", Value: utils.Float64Ptr(m.Sys)},
		{Id: "TotalAlloc", Value: utils.Float64Ptr(m.TotalAlloc)},
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
			"metricId":    metric.Id,
			"metricValue": fmt.Sprintf("%f", *metric.Value),
		}

		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			SetPathParams(params).
			Post("http://localhost:8080/update/{typeMetric}/{metricId}/{metricValue}")

		if err != nil {
			fmt.Printf("▶️  REQUEST ERROR: %v\n", err)
		}
		fmt.Printf("url: %s, code: %d\n", resp.Request.URL, resp.StatusCode())
	}

	c.metricsLock.Unlock()
}
