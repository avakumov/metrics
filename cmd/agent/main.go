package main

import (
	"time"

	"github.com/avakumov/metrics/internal/agent"
	"github.com/avakumov/metrics/internal/agent/config"
	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/utils"
	"go.uber.org/zap"
)

func main() {

	options := config.GetOptions()

	logger.Init(options.Level, "client")
	defer logger.Log.Sync() //nolint:errcheck
	logger.Log.Sugar().Infof("START OPTIONS: %+v", options)

	collector := agent.NewMetricsCollector("http://"+options.Address, options.Key)

	collectTicker := time.NewTicker(time.Duration(options.PollInterval) * time.Second)
	defer collectTicker.Stop()

	sendTicker := time.NewTicker(time.Duration(options.ReportInterval) * time.Second)
	defer sendTicker.Stop()

	numWorker := options.RateLimit
	pool := utils.NewWorkerPool(numWorker, 20, func() {
		err := collector.PostMetrics()
		if err != nil {
			logger.Log.Error("post metrics error", zap.Error(err))
		}
	})
	pool.Start()
	defer pool.Stop()
	counter := 1

	for {
		select {
		case <-collectTicker.C:
			go collector.Collect()
			go collector.CollectSystemMetrics()
		case <-sendTicker.C:
			counter++
			go pool.Submit(counter)
		}
	}
}
