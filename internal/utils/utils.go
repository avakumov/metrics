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
