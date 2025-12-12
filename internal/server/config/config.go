package config

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type Options struct {
	Address string `env:"ADDRESS"`
	Level   string
}

func (a Options) String() string {
	return fmt.Sprintf("address:%s", a.Address)
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

	//default options
	options := Options{
		Address: "localhost:8080",
		Level:   "info",
	}

	//options from env
	err := env.Parse(&options)
	if err != nil {
		log.Printf("error parse options from env %v\n", err)
	}

	//options from flags
	flag.StringVar(&options.Address, "a", options.Address, "Server address in format host:port")
	flag.StringVar(&options.Level, "l", options.Level, "Level of logging")
	flag.Parse()

	log.Printf("Start server options %v\n", zap.String("Address", options.Address))
	return options
}
