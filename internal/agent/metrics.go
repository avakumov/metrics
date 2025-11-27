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
)

// GaugeMetric представляет метрику типа gauge
type GaugeMetric struct {
	Name  string
	Value float64
}

// MemStatsCollector собирает и управляет метриками
type MemStatsCollector struct {
	metricsLock sync.Mutex
	metrics     []GaugeMetric
	count       int
	RandomValue int64
}

// NewMemStatsCollector создает новый сборщик метрик
func NewMemStatsCollector() *MemStatsCollector {
	return &MemStatsCollector{}
}

// Collect собирает все метрики памяти
func (c *MemStatsCollector) Collect() []GaugeMetric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	c.metricsLock.Lock()

	c.metrics = []GaugeMetric{
		{Name: "Alloc", Value: float64(m.Alloc)},
		{Name: "BuckHashSys", Value: float64(m.BuckHashSys)},
		{Name: "Frees", Value: float64(m.Frees)},
		{Name: "GCCPUFraction", Value: m.GCCPUFraction},
		{Name: "GCSys", Value: float64(m.GCSys)},
		{Name: "HeapAlloc", Value: float64(m.HeapAlloc)},
		{Name: "HeapIdle", Value: float64(m.HeapIdle)},
		{Name: "HeapInuse", Value: float64(m.HeapInuse)},
		{Name: "HeapObjects", Value: float64(m.HeapObjects)},
		{Name: "HeapReleased", Value: float64(m.HeapReleased)},
		{Name: "HeapSys", Value: float64(m.HeapSys)},
		{Name: "LastGC", Value: float64(m.LastGC)},
		{Name: "Lookups", Value: float64(m.Lookups)},
		{Name: "MCacheInuse", Value: float64(m.MCacheInuse)},
		{Name: "MCacheSys", Value: float64(m.MCacheSys)},
		{Name: "MSpanInuse", Value: float64(m.MSpanInuse)},
		{Name: "MSpanSys", Value: float64(m.MSpanSys)},
		{Name: "Mallocs", Value: float64(m.Mallocs)},
		{Name: "NextGC", Value: float64(m.NextGC)},
		{Name: "NumForcedGC", Value: float64(m.NumForcedGC)},
		{Name: "NumGC", Value: float64(m.NumGC)},
		{Name: "OtherSys", Value: float64(m.OtherSys)},
		{Name: "PauseTotalNs", Value: float64(m.PauseTotalNs)},
		{Name: "StackInuse", Value: float64(m.StackInuse)},
		{Name: "StackSys", Value: float64(m.StackSys)},
		{Name: "Sys", Value: float64(m.Sys)},
		{Name: "TotalAlloc", Value: float64(m.TotalAlloc)},
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

		url := fmt.Sprintf("%s/%s/%s/%s/%f", endpoint, "update", "gauge", metric.Name, metric.Value)
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
func (c *MemStatsCollector) GetMetric(name string) (GaugeMetric, bool) {
	for _, metric := range c.metrics {
		if metric.Name == name {
			return metric, true
		}
	}
	return GaugeMetric{}, false
}

// PrintMetrics печатает все метрики в читаемом формате
func (c *MemStatsCollector) PrintMetrics() {
	fmt.Printf("=== Memory Metrics at %s ===\n", time.Now().Format(time.RFC3339))
	for _, metric := range c.metrics {
		fmt.Printf("%-20s: %.2f\n", metric.Name, metric.Value)
	}
	fmt.Println("=====================================")
}

// ToJSON возвращает метрики в формате JSON
func (c *MemStatsCollector) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c.metrics, "", "  ")
}
