package agent

import (
	"sync"
	"testing"
	"time"
)

func TestMemStatsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector("http://localhost:8080")

	// Вызываем Collect несколько раз для проверки
	metrics1 := collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик

	// Проверяем, что возвращается непустой слайс
	if len(metrics1) == 0 {
		t.Error("Expected non-empty metrics slice, got empty")
	}

	// Проверяем структуру метрик
	expectedMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
		"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
		"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
		"Sys", "TotalAlloc",
	}

	// Проверяем наличие всех ожидаемых метрик
	metricMap := make(map[string]bool)
	for _, metric := range metrics1 {
		metricMap[metric.ID] = true
	}

	for _, expected := range expectedMetrics {
		if !metricMap[expected] {
			t.Errorf("Expected metric %s not found in collected metrics", expected)
		}
	}

}

func TestMemStatsCollector_Collect_Concurrent(t *testing.T) {
	collector := NewMetricsCollector("http://localhost:8080")
	var wg sync.WaitGroup
	iterations := 100

	// Запускаем несколько горутин для конкурентного доступа
	for range iterations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			metrics := collector.Collect()
			if len(metrics) == 0 {
				t.Error("Expected non-empty metrics in concurrent access")
			}
		}()
	}

	wg.Wait()

}
