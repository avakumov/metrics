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
	Address         string `env:"ADDRESS"`           //flag -a
	FileStoragePath string `env:"FILE_STORAGE_PATH"` //flag -f
	StoreInterval   int    `env:"STORE_INTERVAL"`    //flag -i
	Restore         bool   `env:"RESTORE"`           //flag -r
	LogLevel        string `env:"LOG_LEVEL"`         //flag -log
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
		Address:         "localhost:8080",
		FileStoragePath: "data.json",
		StoreInterval:   300,
		Restore:         true,
		LogLevel:        "info",
	}

	//options from env
	err := env.Parse(&options)
	if err != nil {
		log.Printf("error parse options from env %v\n", err)
	}

	//options from flags
	flag.StringVar(&options.Address, "a", options.Address, "Server address in format host:port")
	flag.StringVar(&options.FileStoragePath, "f", options.FileStoragePath, "Path to storage file")
	flag.IntVar(&options.StoreInterval, "i", options.StoreInterval, "Store data interval")
	flag.BoolVar(&options.Restore, "r", options.Restore, "Restore data on start")
	flag.StringVar(&options.LogLevel, "log", options.LogLevel, "Level of logging")
	flag.Parse()

	return options
}
