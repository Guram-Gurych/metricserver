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
	ServerAddress   string
	ReportInterval  time.Duration
	PollInterval    time.Duration
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func InitConfigServer() *Config {
	var config Config
	var storeInterval int64

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "The address for launching the HTTP server")
	flag.Int64Var(&storeInterval, "i", 300, "the time interval after which the server readings are saved to disk (in seconds)")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/metrics-db.json", "The name of the file where the current values are saved")
	flag.BoolVar(&config.Restore, "r", true, "The value that determines whether or not to load previously saved values from the specified file at server startup")
	flag.Parse()

	config.StoreInterval = time.Duration(storeInterval) * time.Second

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		config.ServerAddress = envAddr
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if val, err := strconv.ParseInt(envStoreInterval, 10, 64); err != nil {
			// loger
		} else {
			config.StoreInterval = time.Duration(val) * time.Second
		}
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		config.FileStoragePath = envFileStoragePath
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if val, err := strconv.ParseBool(envRestore); err != nil {
			// loger
		} else {
			config.Restore = val
		}
	}

	return &config
}

func InitConfigAgent() *Config {
	var config Config
	var reportIntervalSec, pollIntervalSec int64

	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "HTTP Server endpoint address")
	flag.Int64Var(&reportIntervalSec, "r", 10, "The frequency of sending metrics to the server (in seconds)")
	flag.Int64Var(&pollIntervalSec, "p", 2, "The frequency of polling metrics (in seconds)")
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
