package agent

import (
	"math/rand"
	"runtime"
	"strconv"
	"sync"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/utils"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// MemStatsCollector собирает и управляет метриками
type MemStatsCollector struct {
	mu          sync.Mutex
	metrics     []models.Metric
	restyClient *resty.Client
}

// NewMemStatsCollector создает новый сборщик метрик
func NewMemStatsCollector(url string) *MemStatsCollector {
	client := resty.New()
	client.SetBaseURL(url)
	return &MemStatsCollector{
		restyClient: client,
	}
}

// Collect собирает все метрики памяти
func (c *MemStatsCollector) Collect() []models.Metric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := []models.Metric{
		{ID: models.Alloc, MType: "gauge", Value: utils.Float64Ptr(m.Alloc)},
		{ID: "BuckHashSys", MType: "gauge", Value: utils.Float64Ptr(m.BuckHashSys)},
		{ID: "Frees", MType: "gauge", Value: utils.Float64Ptr(m.Frees)},
		{ID: "GCCPUFraction", MType: "gauge", Value: utils.Float64Ptr(m.GCCPUFraction)},
		{ID: "GCSys", MType: "gauge", Value: utils.Float64Ptr(m.GCSys)},
		{ID: "HeapAlloc", MType: "gauge", Value: utils.Float64Ptr(m.HeapAlloc)},
		{ID: "HeapIdle", MType: "gauge", Value: utils.Float64Ptr(m.HeapIdle)},
		{ID: "HeapInuse", MType: "gauge", Value: utils.Float64Ptr(m.HeapInuse)},
		{ID: "HeapObjects", MType: "gauge", Value: utils.Float64Ptr(m.HeapObjects)},
		{ID: "HeapReleased", MType: "gauge", Value: utils.Float64Ptr(m.HeapReleased)},
		{ID: "HeapSys", MType: "gauge", Value: utils.Float64Ptr(m.HeapSys)},
		{ID: "LastGC", MType: "gauge", Value: utils.Float64Ptr(m.LastGC)},
		{ID: "Lookups", MType: "gauge", Value: utils.Float64Ptr(m.Lookups)},
		{ID: "MCacheInuse", MType: "gauge", Value: utils.Float64Ptr(m.MCacheInuse)},
		{ID: "MCacheSys", MType: "gauge", Value: utils.Float64Ptr(m.MCacheSys)},
		{ID: "MSpanInuse", MType: "gauge", Value: utils.Float64Ptr(m.MSpanInuse)},
		{ID: "MSpanSys", MType: "gauge", Value: utils.Float64Ptr(m.MSpanSys)},
		{ID: "Mallocs", MType: "gauge", Value: utils.Float64Ptr(m.Mallocs)},
		{ID: "NextGC", MType: "gauge", Value: utils.Float64Ptr(m.NextGC)},
		{ID: "NumForcedGC", MType: "gauge", Value: utils.Float64Ptr(m.NumForcedGC)},
		{ID: "NumGC", MType: "gauge", Value: utils.Float64Ptr(m.NumGC)},
		{ID: "OtherSys", MType: "gauge", Value: utils.Float64Ptr(m.OtherSys)},
		{ID: "PauseTotalNs", MType: "gauge", Value: utils.Float64Ptr(m.PauseTotalNs)},
		{ID: "StackInuse", MType: "gauge", Value: utils.Float64Ptr(m.StackInuse)},
		{ID: "StackSys", MType: "gauge", Value: utils.Float64Ptr(m.StackSys)},
		{ID: "Sys", MType: "gauge", Value: utils.Float64Ptr(m.Sys)},
		{ID: "TotalAlloc", MType: "gauge", Value: utils.Float64Ptr(m.TotalAlloc)},
		{ID: "RandomValue", MType: "gauge", Value: utils.Float64Ptr(rand.Float64() * 1000.0)},
	}
	setCounter(&metrics)

	c.mu.Lock()
	c.metrics = metrics
	c.mu.Unlock()

	return c.metrics
}

func (c *MemStatsCollector) SendMetrics() {

	client := c.restyClient

	c.mu.Lock()
	metrics := make([]models.Metric, len(c.metrics))
	copy(metrics, c.metrics)
	c.mu.Unlock()

	for _, metric := range metrics {
		metricValue := strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		params := map[string]string{
			"typeMetric":  metric.MType,
			"metricID":    metric.ID,
			"metricValue": metricValue,
		}

		resp, err := client.R().
			SetHeader("Content-Type", "text/plain").
			SetPathParams(params).
			Post("/update/{typeMetric}/{metricID}/{metricValue}")

		if err != nil {
			logger.Log.Error("request error", zap.Error(err))
		}
		logger.Log.Info("SEND METRIC", zap.String("url", resp.Request.URL), zap.Int("code", resp.StatusCode()))
	}

}

func setCounter(metrics *[]models.Metric) {
	for i := range *metrics {
		if (*metrics)[i].ID == "pollcount" {
			*(*metrics)[i].Value += 1.0
			return
		}
	}
	*metrics = append(*metrics, models.Metric{
		ID:    "pollcount",
		MType: "counter",
		Value: utils.Float64Ptr(1.0),
	})
}
