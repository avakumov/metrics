package router

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		contentType      string
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
		},
		{
			name:           "PUT method not allowed",
			method:         http.MethodPut,
			path:           "/update/counter/testCounter/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE method not allowed",
			method:         http.MethodDelete,
			path:           "/update/counter/testCounter/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusMethodNotAllowed,
		},

		// Тесты на неправильный путь
		{
			name:           "Too short path",
			method:         http.MethodPost,
			path:           "/update",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty path parts",
			method:         http.MethodPost,
			path:           "///",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},

		// Тесты на неправильный формат пути
		{
			name:           "Wrong first path part",
			method:         http.MethodPost,
			path:           "/wrong/counter/test/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Wrong metric type",
			method:         http.MethodPost,
			path:           "/update/wrongtype/test/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},

		// Тесты на неполный путь
		{
			name:           "Missing metric value",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Too long path",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/1/extra",
			contentType:    "text/plain",
			expectedStatus: http.StatusBadRequest,
		},

		// Тесты на пустое имя метрики
		{
			name:           "Empty metric name",
			method:         http.MethodPost,
			path:           "/update/counter//1",
			contentType:    "text/plain",
			expectedStatus: http.StatusNotFound,
		},

		// Тесты на валидные пути
		{
			name:   "Valid counter metric",
			method: http.MethodPost,
			path:   "/update/counter/testCounter/1",
			//contentType:    "text/plain",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid gauge metric",
			method:         http.MethodPost,
			path:           "/update/gauge/testGauge/1.5",
			contentType:    "text/plain",
			expectedStatus: http.StatusOK,
		},

		//считывание метрик
		{
			name:             "Get metric",
			method:           http.MethodGet,
			path:             "/value/counter/testCounter",
			expectedStatus:   http.StatusOK,
			expectedResponse: "1",
		},

		{
			name:           "Valid counter metric +3",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/3",
			expectedStatus: http.StatusOK,
		},
		{
			name:             "Get metric",
			method:           http.MethodGet,
			path:             "/value/counter/testCounter",
			expectedStatus:   http.StatusOK,
			expectedResponse: "4",
		},
		{
			name:             "Get metric with value",
			method:           http.MethodGet,
			path:             "/value/gauge/testGauge",
			expectedStatus:   http.StatusOK,
			expectedResponse: "1.5",
		},

		{
			name:           "Post with empty counter name",
			method:         http.MethodPost,
			path:           "/update/counter/",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Post with empty gauge name",
			method:         http.MethodPost,
			path:           "/update/gauge/",
			expectedStatus: http.StatusNotFound,
		},
	}

	r := MetricsRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, v.method, v.path)
			// Используем Cleanup для гарантированного закрытия
			t.Cleanup(func() {
				resp.Body.Close()
			})
			assert.Equal(t, v.expectedStatus, resp.StatusCode)
			if v.expectedResponse != "" {
				assert.Equal(t, v.expectedResponse, body)
			}

		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
