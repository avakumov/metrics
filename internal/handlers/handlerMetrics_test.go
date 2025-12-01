package handlers

import (
	"github.com/avakumov/metrics/internal/repository"
	"github.com/avakumov/metrics/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateMetricHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		contentType    string
		expectedStatus int
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

		// Тесты на неправильный Content-Type
		{
			name:           "Wrong content type",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/1",
			contentType:    "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Empty content type",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/1",
			contentType:    "",
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
			name:           "Valid counter metric",
			method:         http.MethodPost,
			path:           "/update/counter/testCounter/1",
			contentType:    "text/plain",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid gauge metric",
			method:         http.MethodPost,
			path:           "/update/gauge/testGauge/1.5",
			contentType:    "text/plain",
			expectedStatus: http.StatusOK,
		},
	}

	metricsRepo := repository.NewMemoryRepository()
	metricService := service.NewMetricService(metricsRepo)
	metricHandler := NewMetricHandler(metricService)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("Could not create request: %v", err)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(metricHandler.UpdateMetricHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

		})
	}
}

