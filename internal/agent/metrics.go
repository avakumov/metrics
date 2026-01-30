package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	agenterrors "github.com/avakumov/metrics/internal/agent/errors"
	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/utils"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// MetricsCollector собирает и управляет метриками
type MetricsCollector struct {
	mu             sync.Mutex
	metrics        []models.Metric
	restyClient    *resty.Client
	retryDurations []time.Duration
}

// NewMetricsCollector создает новый сборщик метрик
func NewMetricsCollector(url string) *MetricsCollector {
	client := resty.New()
	client.SetBaseURL(url)

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		return nil
	})
	client.OnAfterResponse(
		func(c *resty.Client, resp *resty.Response) error {
			logger.Log.Info("REQUEST: ", zap.String("url", resp.Request.URL), zap.Int("code", resp.StatusCode()))
			return nil
		})
	return &MetricsCollector{
		restyClient:    client,
		retryDurations: []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second},
	}
}

// Collect собирает все метрики памяти
func (c *MetricsCollector) Collect() {
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
}

// отправка метрик по одной
func (c *MetricsCollector) PostMetricsByJSON() {
	metrics := c.getMetrics()
	for _, metric := range metrics {
		// Конвертируем в JSON
		jsonData, err := json.Marshal(metric)
		if err != nil {
			logger.Log.Error("json error", zap.Error(err))
			continue
		}

		// Сжимаем если большой
		body := jsonData
		compressed, err := compressGzip(jsonData)
		if err == nil {
			body = compressed
		}
		update := func() (*resty.Response, error) {
			return c.restyClient.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Content-Encoding", "gzip").
				SetBody(body).
				Post("/update/")
		}
		_, err = retry(update, c.retryDurations)

		if err != nil {
			logger.Log.Error("request error", zap.Error(err))
		}
	}
}

// отправка метрик пачкой
func (c *MetricsCollector) PostMetrics() error {
	metrics := c.getMetrics()
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("json error", zap.Error(err))
		return fmt.Errorf("json error: %w", err)
	}
	body, err := compressGzip(jsonData)
	if err != nil {
		logger.Log.Error("compress with error", zap.Error(err))
		return err
	}
	update := func() (*resty.Response, error) {
		return c.restyClient.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(body).
			Post("/updates/")
	}

	resp, err := retry(update, c.retryDurations)

	if resp != nil {
		logger.Log.Info("RESPONSE", zap.Int("STATUS", resp.StatusCode()))
	}
	if err != nil {
		logger.Log.Error("request error", zap.Error(err))
		return fmt.Errorf("request error: %w", err)
	}
	return nil
}

func (c *MetricsCollector) PostMetricsByURL() {
	metrics := c.getMetrics()
	for _, metric := range metrics {
		var metricValue string
		if metric.MType == models.Gauge {
			metricValue = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
		}
		if metric.MType == models.Counter {
			metricValue = strconv.FormatInt(*metric.Delta, 10)
		}
		params := map[string]string{
			"typeMetric":  metric.MType,
			"metricID":    metric.ID,
			"metricValue": metricValue,
		}
		update := func() (*resty.Response, error) {
			return c.restyClient.R().
				SetHeader("Content-Type", "text/plain").
				SetPathParams(params).
				Post("/update/{typeMetric}/{metricID}/{metricValue}")
		}

		_, err := retry(update, c.retryDurations)

		if err != nil {
			logger.Log.Error("request error", zap.Error(err))
		}
	}
}

func (c *MetricsCollector) getMetrics() []models.Metric {
	c.mu.Lock()
	metrics := make([]models.Metric, len(c.metrics))
	copy(metrics, c.metrics)
	c.mu.Unlock()

	return metrics
}

func setCounter(metrics *[]models.Metric) {
	for i := range *metrics {
		if (*metrics)[i].ID == "PollCount" {
			*(*metrics)[i].Delta += 1
			return
		}
	}
	var startCounter int64 = 1
	*metrics = append(*metrics, models.Metric{
		ID:    "PollCount",
		MType: "counter",
		Delta: &startCounter,
	})
}

func compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	gz.Close()
	return buf.Bytes(), nil
}

func retry(f func() (*resty.Response, error), durations []time.Duration) (*resty.Response, error) {
	var err error
	var resp *resty.Response
	for i, duration := range durations {
		resp, err = f()
		if agenterrors.IsRetryableError(resp, err) {
			logger.Log.Info("request error. Try again after",
				zap.Duration("duration", duration),
				zap.Int("attempt", i+1),
				zap.Int("total attempt", len(durations)),
				zap.Error(err))
			time.Sleep(duration)
			continue
		}
		return resp, err
	}
	return resp, err
}
