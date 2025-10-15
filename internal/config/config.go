package config

import (
	"flag"
	"log"
	"os"
	"strconv"
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

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		config.ServerAddress = envAddr
	}

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

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		config.ServerAddress = envAddr
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		if val, err := strconv.ParseInt(envReportInterval, 10, 64); err != nil {
			log.Printf("WARN: неверное значение переменной REPORT_INTERVAL: '%s'. Используется значение по умолчанию.", envReportInterval)
		} else {
			config.ReportInterval = time.Duration(val) * time.Second
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		if val, err := strconv.ParseInt(envPollInterval, 10, 64); err != nil {
			log.Printf("WARN: неверное значение переменной POLL_INTERVAL: '%s'. Используется значение по умолчанию.", envPollInterval)
		} else {
			config.PollInterval = time.Duration(val) * time.Second
		}
	}

	if !strings.HasPrefix(config.ServerAddress, "http://") {
		config.ServerAddress = "http://" + config.ServerAddress
	}

	return &config
}
