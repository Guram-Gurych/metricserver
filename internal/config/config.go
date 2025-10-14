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

func InitConfigServer() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Адрес для запуска HTTP-сервера")
	flag.Parse()

	return &config
}

func InitConfigAgent() *Config {
	var config Config

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Адрес эндпоинта HTTP-сервера")
	flag.DurationVar(&config.ReportInterval, "r", 10*time.Second, "Частота отправки метрик на сервер")
	flag.DurationVar(&config.PollInterval, "p", 2*time.Second, "Частота опроса метрик")
	flag.Parse()

	if !strings.HasPrefix(config.ServerAddress, "http://") {
		config.ServerAddress = "http://" + config.ServerAddress
	}

	return &config
}
