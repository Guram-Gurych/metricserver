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

	flag.StringVar(&config.ServerAddress, "a", ":8080", "Адрес для запуска HTTP-сервера")
	flag.Parse()

	return &config
}

func InitConfigAgent() *Config {
	var config Config
	var reportInterval int
	var pollInterval int

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Адрес эндпоинта HTTP-сервера")
	flag.IntVar(&reportInterval, "r", 10, "Частота отправки метрик на сервер")
	flag.IntVar(&pollInterval, "p", 2, "Частота опроса метрик")
	flag.Parse()

	config.ReportInterval = time.Duration(reportInterval) * time.Second
	config.PollInterval = time.Duration(pollInterval) * time.Second

	if !strings.HasPrefix(config.ServerAddress, "http://") {
		config.ServerAddress = "http://" + config.ServerAddress
	}

	return &config
}
