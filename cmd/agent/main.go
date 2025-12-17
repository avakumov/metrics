package main

import (
	"time"

	"github.com/avakumov/metrics/internal/agent"
	"github.com/avakumov/metrics/internal/agent/config"
	"github.com/avakumov/metrics/internal/logger"
	"go.uber.org/zap"
)

func main() {

	options := config.GetOptions()

	logger.Init(options.Level)
	defer logger.Log.Sync()

	logger.Log.Info("metrics client app starting...", zap.String("address", options.Address))
	collector := agent.NewMemStatsCollector("http://" + options.Address)

	collectTicker := time.NewTicker(time.Duration(options.PollInterval) * time.Second)
	defer collectTicker.Stop()

	sendTicker := time.NewTicker(time.Duration(options.ReportInterval) * time.Second)
	defer sendTicker.Stop()

	for {
		select {
		case <-collectTicker.C:
			collector.Collect()
		case <-sendTicker.C:
			collector.PostMetricsByJSON()
		}
	}
}
