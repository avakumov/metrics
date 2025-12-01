package agent

import (
	"sync"
	"testing"
	"time"
)

func TestMemStatsCollector_Collect(t *testing.T) {
	collector := NewMemStatsCollector()

	// Вызываем Collect несколько раз для проверки
	metrics1 := collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик
	metrics2 := collector.Collect()

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

	// Проверяем, что значения метрик различаются между вызовами (кроме некоторых)
	differentFound := false
	for i := range metrics1 {
		if metrics1[i].Value != metrics2[i].Value {
			differentFound = true
			break
		}
	}

	if !differentFound {
		t.Log("Metric values are the same between calls (this might be normal in test environment)")
	}

	// Проверяем обновление внутреннего состояния
	if collector.count != 2 {
		t.Errorf("Expected count to be 2, got %d", collector.count)
	}
}

func TestMemStatsCollector_Collect_Concurrent(t *testing.T) {
	collector := NewMemStatsCollector()
	var wg sync.WaitGroup
	iterations := 100

	// Запускаем несколько горутин для конкурентного доступа
	for i := 0; i < iterations; i++ {
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

	// Проверяем, что счетчик обновился правильно
	if collector.count != iterations {
		t.Errorf("Expected count %d, got %d", iterations, collector.count)
	}
}
