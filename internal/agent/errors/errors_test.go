package agenterrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	//	"github.com/stretchr/testify/require"
)

// TestIsRetryableError_StatusCode проверяет логику по статус-кодам
func TestIsRetryableError_StatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		expected   bool
	}{
		// 2xx - не retryable
		{
			name:       "status 200 OK - not retryable",
			statusCode: 200,
			err:        nil,
			expected:   false,
		},
		{
			name:       "status 201 Created - not retryable",
			statusCode: 201,
			err:        nil,
			expected:   false,
		},
		{
			name:       "status 204 No Content - not retryable",
			statusCode: 204,
			err:        nil,
			expected:   false,
		},

		// 3xx - не retryable (по текущей логике)
		{
			name:       "status 301 Moved - not retryable",
			statusCode: 301,
			err:        nil,
			expected:   false,
		},

		// 4xx - не retryable (кроме 429)
		{
			name:       "status 400 Bad Request - not retryable",
			statusCode: 400,
			err:        nil,
			expected:   false,
		},
		{
			name:       "status 401 Unauthorized - not retryable",
			statusCode: 401,
			err:        nil,
			expected:   false,
		},
		{
			name:       "status 403 Forbidden - not retryable",
			statusCode: 403,
			err:        nil,
			expected:   false,
		},
		{
			name:       "status 404 Not Found - not retryable",
			statusCode: 404,
			err:        nil,
			expected:   false,
		},
		{
			name:       "status 408 Request Timeout - not retryable",
			statusCode: 408,
			err:        nil,
			expected:   false, // По текущей логике! Но обычно 408 - retryable
		},
		{
			name:       "status 429 Too Many Requests - retryable",
			statusCode: 429,
			err:        nil,
			expected:   true, // Особый случай в коде
		},

		// 5xx - retryable
		{
			name:       "status 500 Internal Error - retryable",
			statusCode: 500,
			err:        nil,
			expected:   true,
		},
		{
			name:       "status 502 Bad Gateway - retryable",
			statusCode: 502,
			err:        nil,
			expected:   true,
		},
		{
			name:       "status 503 Service Unavailable - retryable",
			statusCode: 503,
			err:        nil,
			expected:   true,
		},
		{
			name:       "status 504 Gateway Timeout - retryable",
			statusCode: 504,
			err:        nil,
			expected:   true,
		},
		{
			name:       "Connection refused - retryable",
			statusCode: 0,
			err:        fmt.Errorf("connection refused"),
			expected:   true,
		},
		{
			name:       "Unknown host - not retryable",
			statusCode: 0,
			err:        fmt.Errorf("unknown host"),
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.statusCode, tt.err)
			assert.Equal(t, tt.expected, result,
				"For status %d expected %v, got %v", tt.statusCode, tt.expected, result)
		})
	}
}
