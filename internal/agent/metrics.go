package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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
func NewMetricsCollector(url string, key string) *MetricsCollector {
	client := resty.New()
	client.SetBaseURL(url)

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		if key == "" {
			return nil
		}
		bodyBytes, err := getRequestBodyAsBytes(req)
		if err != nil {
			return fmt.Errorf("failed to get request body: %w", err)
		}
		hash, err := utils.HashHMACSHA256(bodyBytes, key)
		if err != nil {
			return fmt.Errorf("failed to compute hash: %w", err)
		}

		req.SetHeader("HashSHA256", hash)
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
	c.updateCounter()
	c.updateMetrics(metrics)
}

func (c *MetricsCollector) CollectSystemMetrics() {

	// Сбор информации о памяти
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Error("failed to get virtual memory stats:", zap.Error(err))
	}
	metrics := []models.Metric{
		{ID: "TotalMemory", MType: "gauge", Value: utils.Float64Ptr(vmStat.Total)},
		{ID: "FreeMemory", MType: "gauge", Value: utils.Float64Ptr(vmStat.Free)},
	}

	// Получаем процент использования для каждого CPU
	percentages, err := cpu.Percent(100*time.Millisecond, true) // true - для каждого CPU отдельно
	if err != nil {
		logger.Log.Error("failed to get cpu percentages", zap.Error(err))
	}
	for i := 0; i < len(percentages); i++ {
		metrics = append(metrics, models.Metric{ID: fmt.Sprintf("CPUutilization%d", i+1), MType: "gauge", Value: &percentages[i]})
	}
	c.updateMetrics(metrics)
}

func (c *MetricsCollector) updateMetric(metric models.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, m := range c.metrics {
		if metric.ID == m.ID {

			c.metrics[i] = metric
			return
		}
	}
	c.metrics = append(c.metrics, metric)
}

func (c *MetricsCollector) updateMetrics(metrics []models.Metric) {
	for _, m := range metrics {
		c.updateMetric(m)
	}

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
		r := c.restyClient.R()

		r.SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetBody(body)
		return r.Post("/updates/")
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

func (c *MetricsCollector) updateCounter() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.metrics {
		if c.metrics[i].ID == "PollCount" {
			// Работаем напрямую с элементом среза через индекс
			if c.metrics[i].Delta == nil {
				var start int64 = 1
				c.metrics[i].Delta = &start
			} else {
				*c.metrics[i].Delta++
			}
			return
		}
	}

	// Если не нашли, создаём новую метрику
	startCounter := int64(1)
	c.metrics = append(c.metrics, models.Metric{
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

// при retryable ошибке будет делать еще запросы
func retry(f func() (*resty.Response, error), durations []time.Duration) (*resty.Response, error) {
	var err error
	var resp *resty.Response
	for i, duration := range durations {
		resp, err = f()
		var statusCode int
		if resp != nil {
			statusCode = resp.StatusCode()
		}

		if agenterrors.IsRetryableError(statusCode, err) {
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

func getRequestBodyAsBytes(req *resty.Request) ([]byte, error) {
	if req.Body == nil {
		return []byte{}, nil
	}

	switch body := req.Body.(type) {
	case []byte:
		return body, nil
	case string:
		return []byte(body), nil
	case io.Reader:
		return io.ReadAll(body)
	default:
		//для структур/map - сериализуем в JSON
		return json.Marshal(body)
	}
}
