package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/avakumov/metrics/internal/models"
	"github.com/avakumov/metrics/internal/utils"

	//"github.com/avakumov/metrics/internal/utils"
	//"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStatsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector("http://localhost:8080", "")

	// Вызываем Collect несколько раз для проверки
	collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик

	// Проверяем, что возвращается непустой слайс
	if len(collector.metrics) == 0 {
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
	for _, metric := range collector.metrics {
		metricMap[metric.ID] = true
	}

	for _, expected := range expectedMetrics {
		if !metricMap[expected] {
			t.Errorf("Expected metric %s not found in collected metrics", expected)
		}
	}

}

func TestMemStatsCollector_Collect_Concurrent(t *testing.T) {
	collector := NewMetricsCollector("http://localhost:8080", "")
	var wg sync.WaitGroup
	iterations := 100

	// Запускаем несколько горутин для конкурентного доступа
	for range iterations {
		wg.Add(1)
		go func() {
			defer wg.Done()
			collector.Collect()
			metrics := collector.getMetrics()
			if len(metrics) == 0 {
				t.Error("Expected non-empty metrics in concurrent access")
			}
		}()
	}

	wg.Wait()

}

func TestPostMetrics_Integration(t *testing.T) {
	// Запускаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		// Декомпрессим тело
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.NoError(t, err)

		// Парсим JSON
		var metrics []models.Metric
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)

		// Проверяем данные
		assert.Len(t, metrics, 29)

		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	collector := NewMetricsCollector(server.URL, "")
	collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик
	err := collector.PostMetrics()

	assert.NoError(t, err)
}

func TestPostMetrics_Integration_With_Retry(t *testing.T) {
	attempts := 0
	// Запускаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	collector := NewMetricsCollector(server.URL, "")
	collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик
	err := collector.PostMetrics()

	assert.NoError(t, err)
	assert.Equal(t, 2, attempts, "Should retry after 500 error")
}

func TestPostMetrics_Integration_Without_retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	collector := NewMetricsCollector(server.URL, "")
	collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик
	err := collector.PostMetrics()

	assert.NoError(t, err)
	assert.Equal(t, 1, attempts, "Should no retry after 403 error")
}

func TestPostMetricsHashSHA256_Integration(t *testing.T) {
	key := "superpassword"
	// Запускаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Проверяем заголовки
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		//проверяем hash
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		// ВОССТАНАВЛИВАЕМ тело для дальнейшего использования
		r.Body = io.NopCloser(bytes.NewReader(data))

		hashSHA256, err := utils.HashHMACSHA256(data, key)
		require.NoError(t, err)
		assert.Equal(t, hashSHA256, r.Header.Get("HashSHA256"))

		// Декомпрессим тело
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gz.Close()

		body, err := io.ReadAll(gz)
		require.NoError(t, err)

		// Парсим JSON
		var metrics []models.Metric
		err = json.Unmarshal(body, &metrics)
		require.NoError(t, err)

		// Проверяем данные
		assert.Len(t, metrics, 29)

		w.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	collector := NewMetricsCollector(server.URL, key)
	collector.Collect()
	time.Sleep(10 * time.Millisecond) // Даем время для изменения метрик
	err := collector.PostMetrics()

	assert.NoError(t, err)
}
