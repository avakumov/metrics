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
	Address string `env:"ADDRESS"`
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
	}

	//options from env
	err := env.Parse(&options)
	if err != nil {
		log.Printf("error parse options from env: %s", err)
	}

	//options from flags
	flag.Var(&options, "a", "Server address in format host:port")
	flag.Parse()

	log.Printf("Options: %s\n", options)
	return options
}
