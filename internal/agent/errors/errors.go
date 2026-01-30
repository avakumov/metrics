package agenterrors

import (
	"strings"

	"github.com/go-resty/resty/v2"
)

func IsRetryableError(resp *resty.Response, err error) bool {
	statusCode := resp.StatusCode()
	if statusCode == 0 && err == nil {
		return false
	}
	//5xx
	if statusCode >= 500 && statusCode <= 599 {
		return true
	}

	if statusCode == 0 && err != nil {
		errString := err.Error()
		retreablePatterns := []string{
			"timeout",
			"deadline exceeded",
			"connection refused",
			"connection reset",
			"temporary failure",
			"network is unreachable",
			"EOF",
			"broken pipe",
			"429", //Too many requests
		}
		for _, pattern := range retreablePatterns {
			if strings.Contains(errString, pattern) {
				return true
			}
		}

		nonRetryablePatterns := []string{
			"no such host",
			"unknown host",
			"x509",
			"certificate",
			"context cancelled",
			"malformed",
			"invalid",
			"401",
			"403",
		}

		for _, pattern := range nonRetryablePatterns {
			if strings.Contains(errString, pattern) {
				return false
			}
		}

	}
	//default status code 0 is retryable
	return statusCode == 0
}
