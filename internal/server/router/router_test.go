package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
			name:             "Post json arrays metrics with id, type",
			method:           http.MethodPost,
			contentType:      "application/json",
			path:             "/value",
			expectedStatus:   http.StatusOK,
			body:             `[{"id":"testgauge", "type":"gauge"}]`,
			expectedResponse: `[{"id":"testgauge", "type":"gauge", "value":"1.5"}]`,
		},
	}

	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo)
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
