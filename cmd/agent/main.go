package main

import (
	"github.com/avakumov/metrics/internal/agent"
	"time"
)

var pollInterval = 2
var reportInterval = 10

func main() {
	collector := agent.NewMemStatsCollector()

	collectTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer collectTicker.Stop()

	sendTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	defer sendTicker.Stop()

	for {
		select {
		case <-collectTicker.C:
			collector.Collect()
		case <-sendTicker.C:
			collector.SendMetrics()
		}
	}
}
