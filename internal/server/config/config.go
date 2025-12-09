package config

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Options struct {
	Port int
	Host string
}

func (a Options) String() string {
	return fmt.Sprintf("host:%s, port:%d", a.Host, a.Port)
}

func (a *Options) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return fmt.Errorf("need address in a format host:port. recieve: %s", s)
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return fmt.Errorf("port [%s] is not integer: %w", hp[1], err)
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}

func GetOptions() Options {

	options := Options{
		Host: "localhost",
		Port: 8080,
	}

	flag.Var(&options, "a", "Server address in format host:port")
	flag.Parse()

	log.Printf("Options: %s\n", options)
	return options
}
