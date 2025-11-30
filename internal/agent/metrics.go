package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/avakumov/metrics/internal/models"

	"github.com/avakumov/metrics/internal/utils"
)

// MemStatsCollector собирает и управляет метриками
type MemStatsCollector struct {
	metricsLock sync.Mutex
	metrics     []models.Metric
	count       int
	RandomValue int64
}

// NewMemStatsCollector создает новый сборщик метрик
func NewMemStatsCollector() *MemStatsCollector {
	return &MemStatsCollector{}
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

	endpoint := "http://localhost:8080"
	client := &http.Client{}

	for _, metric := range c.metrics {

		url := fmt.Sprintf("%s/%s/%s/%s/%f", endpoint, "update", "gauge", metric.Id, *metric.Value)
		fmt.Println(url)
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			panic(err)
		}
		request.Header.Add("Content-Type", "text/plain")
		response, err := client.Do(request)
		if err != nil {
			panic(err)
		}
		// выводим код ответа
		fmt.Println("Статус-код ", response.Status)
		defer response.Body.Close()
		// читаем поток из тела ответа
		body, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		// и печатаем его
		fmt.Println(string(body))
	}

	c.metricsLock.Unlock()
}

// GetMetric возвращает конкретную метрику по имени
func (c *MemStatsCollector) GetMetric(name string) (models.Metric, bool) {
	for _, metric := range c.metrics {
		if metric.Id == name {
			return metric, true
		}
	}
	return models.Metric{}, false
}

// PrintMetrics печатает все метрики в читаемом формате
func (c *MemStatsCollector) PrintMetrics() {
	fmt.Printf("=== Memory Metrics at %s ===\n", time.Now().Format(time.RFC3339))
	for _, metric := range c.metrics {
		fmt.Printf("%-20s: %.2f\n", metric.Id, *metric.Value)
	}
	fmt.Println("=====================================")
}

// ToJSON возвращает метрики в формате JSON
func (c *MemStatsCollector) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c.metrics, "", "  ")
}
