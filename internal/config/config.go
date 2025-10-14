package config

import (
	"flag"
	"strings"
	"time"
)

type Config struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func InitConfig() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", ":8080", "address and port to run server")
	flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "frequency of sending metrics to the server")
	flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "frequency of polling metrics from the runtime package")
	flag.Parse()

	if !strings.HasPrefix(config.ServerAddress, "http://") {
		config.ServerAddress = "http://" + config.ServerAddress
	}

	return &config
}
