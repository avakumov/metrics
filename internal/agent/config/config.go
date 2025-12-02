package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	PollInterval   int
	ReportInterval int
	Port           int
	Host           string
}

func (a Options) String() string {
	return fmt.Sprintf("host:%s, port:%d, pollInterval:%d, reportInterval:%d", a.Host, a.Port, a.PollInterval, a.ReportInterval)
}

func (a *Options) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("Need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func GetOptions() Options {
	options := Options{
		Host:           "localhost",
		Port:           8080,
		PollInterval:   2,
		ReportInterval: 10,
	}

	flag.Var(&options, "a", "Server address in format host:port")
	flag.IntVar(&options.ReportInterval, "r", options.ReportInterval, "Report interval in seconds")
	flag.IntVar(&options.PollInterval, "p", options.PollInterval, "Poll interval in seconds")

	flag.Parse()

	fmt.Printf("options: %s\n", options)
	return options
}
