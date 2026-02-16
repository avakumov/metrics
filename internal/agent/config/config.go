package config

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	PollInterval   int    `env:"POLL_INTERVAL"`   //-p интервал получения метрик в секундах
	ReportInterval int    `env:"REPORT_INTERVAL"` //-r интервал отправки метрик в секундах
	Address        string `env:"ADDRESS"`         //-a address for report metrics
	Level          string `env:"LOG_LEVEL"`       //-log level of logger
	Key            string `env:"KEY"`             //-k key for hash data
	RateLimit      int    `env:"RATE_LIMIT"`      //-l key for rate limit
}

func (a *Options) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return fmt.Errorf("need address in a format host:port. recieve: %s", s)
	}
	_, err := strconv.Atoi(hp[1])
	if err != nil {
		return fmt.Errorf("port [%s] is not integer: %w", hp[1], err)
	}
	a.Address = s
	return nil
}

func GetOptions() Options {
	// defalult options
	options := Options{
		Address:        "localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		Level:          "info",
		RateLimit:      5,
		//Key:            "superpassword",
	}

	// options from env
	err := env.Parse(&options)
	if err != nil {
		log.Printf("error parse options from env: %s", err)
	}

	// options from flags
	flag.StringVar(
		&options.Address,
		"a",
		options.Address,
		"Server address in format host:port",
	)
	flag.StringVar(&options.Level, "log", options.Level, "Level of logging")
	flag.IntVar(
		&options.ReportInterval,
		"r",
		options.ReportInterval,
		"Report interval in seconds",
	)
	flag.IntVar(
		&options.PollInterval,
		"p",
		options.PollInterval,
		"Poll interval in seconds",
	)
	flag.StringVar(&options.Key, "k", options.Key, "Key for for hash")
	flag.IntVar(&options.RateLimit, "l", options.RateLimit, "Rate limit of gourutines")

	flag.Parse()

	return options
}
