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
	var reportIntervalSec, pollIntervalSec int64

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Адрес для запуска HTTP-сервера")
	flag.Int64Var(&reportIntervalSec, "r", 10, "Частота отправки метрик на сервер (в секундах)")
	flag.Int64Var(&pollIntervalSec, "p", 2, "Частота опроса метрик (в секундах)")

	flag.Parse()

	config.ReportInterval = time.Duration(reportIntervalSec) * time.Second
	config.PollInterval = time.Duration(pollIntervalSec) * time.Second

	return &config
}

func InitConfigAgent() *Config {
	var config Config
	var reportIntervalSec, pollIntervalSec int64

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Адрес эндпоинта HTTP-сервера")
	flag.Int64Var(&reportIntervalSec, "r", 10, "Частота отправки метрик на сервер (в секундах)")
	flag.Int64Var(&pollIntervalSec, "p", 2, "Частота опроса метрик (в секундах)")
	flag.Parse()

	config.ReportInterval = time.Duration(reportIntervalSec) * time.Second
	config.PollInterval = time.Duration(pollIntervalSec) * time.Second

	if !strings.HasPrefix(config.ServerAddress, "http://") {
		config.ServerAddress = "http://" + config.ServerAddress
	}

	return &config
}
