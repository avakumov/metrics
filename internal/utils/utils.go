package utils

import (
	"fmt"
	"strconv"
)

func ParseMetric(metricType, metricName, metricValue string) error {
	switch metricType {
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid counter value: %s", metricValue)
		}
		fmt.Printf("Counter: %s = %d \n", metricValue, value)

	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("invalid gauge value: %s", metricValue)
		}
		fmt.Printf("Gauge: %s = %f\n", metricName, value)
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
	return nil
}

// Простой constraint для числовых типов
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Float64Ptr[T Number](value T) *float64 {
	f := float64(value)
	return &f
}
