package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"strconv"
	"strings"
)

type Options struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Address        string `env:"ADDRESS"`
	Level          string
}

func (a Options) String() string {
	return fmt.Sprintf("address:%s, pollInterval:%d, reportInterval:%d", a.Address, a.PollInterval, a.ReportInterval)
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

	//defalult options
	options := Options{
		Address:        "localhost:8080",
		PollInterval:   2,
		ReportInterval: 10,
		Level:          "info",
	}

	//options from env
	err := env.Parse(&options)
	if err != nil {
		log.Printf("error parse options from env: %s", err)
	}

	//options from flags
	flag.StringVar(&options.Address, "a", options.Address, "Server address in format host:port")
	flag.StringVar(&options.Level, "l", options.Level, "Level of logging")
	flag.IntVar(&options.ReportInterval, "r", options.ReportInterval, "Report interval in seconds")
	flag.IntVar(&options.PollInterval, "p", options.PollInterval, "Poll interval in seconds")

	flag.Parse()

	log.Printf("options: %s\n", options)
	return options
}
