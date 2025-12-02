package main

import (
	"fmt"
	"github.com/avakumov/metrics/internal/agent"
	"github.com/avakumov/metrics/internal/agent/config"
	"time"
)

func main() {
	options := config.GetOptions()
	url := fmt.Sprintf("http://%s:%d", options.Host, options.Port)

	collector := agent.NewMemStatsCollector(url)

	collectTicker := time.NewTicker(time.Duration(options.PollInterval) * time.Second)
	defer collectTicker.Stop()

	sendTicker := time.NewTicker(time.Duration(options.ReportInterval) * time.Second)
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
