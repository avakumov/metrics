package router

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compress/gzip"
	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/avakumov/metrics/internal/server/handlers"
	"github.com/avakumov/metrics/internal/server/repository"
	"github.com/avakumov/metrics/internal/server/service"
)

func TestUpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		contentType      string
		body             string
		expectedStatus   int
		expectedResponse string
	}{
		// Тесты на неподдерживаемые HTTP методы
		{
			name:           "GET method not allowed",
			method:         http.MethodGet,
			path:           "/update/counter/testCounter/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
			body:           "",
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			path:           "/update/counter/testCounter/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
			body:           "",
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			path:           "/update/counter/testCounter/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
			body:           "",
		},

		// Тесты на неправильный путь
		{
			name:           "Too short path",
			method:         http.MethodPost,
			path:           "/update",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			body:           "",
		},
		{
			name:           "Empty path parts",
			method:         http.MethodPost,
			path:           "///",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			body:           "",
		},

		// Тесты на неправильный формат пути
		{
			name:           "Wrong first path part",
			method:         http.MethodPost,
			path:           "/wrong/counter/test/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			body:           "",
		},
		{
			name:           "Wrong metric type",
			method:         http.MethodPost,
			path:           "/update/wrongtype/test/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			body:           "",
		},

		// Тесты на неполный путь
		{
			name:           "Missing metric value",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			body:           "",
		},
		{
			name:           "Too long path",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/1/extra",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
			body:           "",
		},

		// Тесты на пустое имя метрики
		{
			name:           "Empty metric name",
			method:         http.MethodPost,
			path:           "/update/counter//1",
			contentType:    "text/plain",
			expectedStatus: http.StatusNotFound,
			body:           "",
		},

		// Тесты на валидные пути
		{
			name:   "Valid counter metric",
			method: http.MethodPost,
			path:   "/update/counter/testCounter/1",
			//contentType:    "text/plain",
			expectedStatus: http.StatusOK,
			body:           "",
		},
		{
			name:           "Valid gauge metric",
			method:         http.MethodPost,
			path:           "/update/gauge/testGauge/1.5",
			contentType:    "text/plain",
			expectedStatus: http.StatusOK,
			body:           "",
		},

		//считывание метрик
		{
			name:             "Get metric",
			method:           http.MethodGet,
			path:             "/value/counter/testCounter",
			expectedStatus:   http.StatusOK,
			expectedResponse: "1",
			body:             "",
		},

		{
			name:           "Valid counter metric +3",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/3",
			expectedStatus: http.StatusOK,
			body:           "",
		},
		{
			name:             "Get metric",
			method:           http.MethodGet,
			path:             "/value/counter/testCounter",
			expectedStatus:   http.StatusOK,
			expectedResponse: "4",
			body:             "",
		},
		{
			name:             "Get metric with value",
			method:           http.MethodGet,
			path:             "/value/gauge/testGauge",
			expectedStatus:   http.StatusOK,
			expectedResponse: "1.5",
			body:             "",
		},

		{
			name:           "Post with empty counter name",
			method:         http.MethodPost,
			path:           "/update/counter/",
			expectedStatus: http.StatusNotFound,
			body:           "",
		},
		{
			name:           "Post with empty gauge name",
			method:         http.MethodPost,
			path:           "/update/gauge/",
			expectedStatus: http.StatusNotFound,
			body:           "",
		},
		{
			name:           "Post json metric by /update/",
			method:         http.MethodPost,
			contentType:    "application/json",
			path:           "/update/",
			expectedStatus: http.StatusOK,
			body:           `{"id":"testGauge", "type":"gauge", "value":155.5}`,
		},
		{
			name:             "Post json with on gauge metric metrics with id, type",
			method:           http.MethodPost,
			contentType:      "application/json",
			path:             "/value/",
			expectedStatus:   http.StatusOK,
			body:             `{"id":"testGauge", "type":"gauge"}`,
			expectedResponse: `{"id":"testGauge", "type":"gauge", "value":155.5}`,
		},
		{
			name:             "Post json with  counter metrics with id, type",
			method:           http.MethodPost,
			contentType:      "application/json",
			path:             "/value/",
			expectedStatus:   http.StatusOK,
			body:             ` {"id":"testCounter", "type":"counter"}`,
			expectedResponse: `{"id":"testCounter", "type":"counter", "delta":4}`,
		},
		{
			name:           "Get metric which not exist",
			method:         http.MethodPost,
			contentType:    "application/json",
			path:           "/value/",
			expectedStatus: http.StatusNotFound,
			body:           ` {"id":"testCounter", "type":"gauge"}`,
		},
	}

	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo, "data.json", 800)
	metricHandler := handlers.NewMetricsHandler(metricService)
	r := MetricsRouter(metricHandler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path, strings.NewReader(v.body), v.contentType)
			// Используем Cleanup для гарантированного закрытия
			t.Cleanup(func() {
				resp.Body.Close()
			})
			assert.Equal(t, v.expectedStatus, resp.StatusCode)
			if v.expectedResponse != "" {
				t.Logf("Body: %s", body)
				assert.JSONEq(t, v.expectedResponse, body)
			}

		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader, contentType string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)
	// Устанавливаем Content-Type, если передан
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

// TestGzipDecoding проверяет декодирование входящих gzip данных
func TestGzipDecoding(t *testing.T) {
	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo, "data.json", 800)
	metricHandler := handlers.NewMetricsHandler(metricService)
	r := MetricsRouter(metricHandler)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Создаем JSON данные
	jsonData := map[string]interface{}{
		"id":    "testMetric",
		"type":  "gauge",
		"value": 123.45,
	}
	data, _ := json.Marshal(jsonData)

	// Сжимаем данные gzip
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write(data)
	gz.Close()

	// Создаем запрос с gzip сжатием
	req := httptest.NewRequest("POST", "/update/", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip") // Для ответа

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Проверяем, что запрос обработан (не 400)
	if rr.Code != http.StatusOK {
		t.Errorf("Handler вернул BadRequest для gzip запроса. Возможно декодирование не работает")
	}

	// Проверяем, что данные корректно распаковались и обработались
	// Для этого можно проверить, что метрика сохранилась
	// или что ответ не содержит ошибок формата
}
