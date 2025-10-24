package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	models "github.com/Guram-Gurych/metricserver.git/internal/model"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

type Agent struct {
	storage        *AgentMetric
	client         *resty.Client
	serverAddress  string
	pollInterval   time.Duration
	reportInterval time.Duration
}

func NewAgent(serverAddress string, pollInterval, reportInterval time.Duration) *Agent {
	return &Agent{
		storage: &AgentMetric{
			Gauges:   make(map[string]float64),
			Counters: make(map[string]int64),
		},
		client:         resty.New(),
		serverAddress:  serverAddress,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}
}

func (a *Agent) Run() {
	pollTicker := time.NewTicker(a.pollInterval)
	reportTicker := time.NewTicker(a.reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			a.pollMetrics()
			log.Println("Метрики собранны")
		case <-reportTicker.C:
			a.reportMetrics()
			log.Println("Метрики отправлены")
		}
	}
}

func (a *Agent) pollMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	v := reflect.ValueOf(m)

	for _, metricName := range GaugeMetrics {
		value := v.FieldByName(metricName)

		var floatValue float64
		if value.CanFloat() {
			floatValue = value.Float()
		} else if value.CanUint() {
			floatValue = float64(value.Uint())
		}

		a.storage.Gauges[metricName] = floatValue
	}

	a.storage.Gauges["RandomValue"] = rand.Float64()
	a.storage.Counters["PollCount"] += 1
}

func (a *Agent) reportMetrics() {
	for name, value := range a.storage.Gauges {
		valueStr := strconv.FormatFloat(value, 'f', -1, 64)
		a.sendMetric("gauge", name, valueStr)
	}

	for name, value := range a.storage.Counters {
		valueStr := strconv.FormatInt(value, 10)
		a.sendMetric("counter", name, valueStr)
	}
}

func (a *Agent) sendMetric(metricType, metricName, metricValue string) {
	m := models.Metrics{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Printf("Ошибка конвертации gauge %s: %v", metricName, err)
			return
		}
		m.Value = &value
	case models.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			log.Printf("Ошибка конвертации counter %s: %v", metricName, err)
			return
		}
		m.Delta = &value

		if metricName == "PollCount" {
			a.storage.Counters["PollCount"] = 0
		}
	default:
		log.Printf("Неизвестный тип метрики: %s", metricType)
		return
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	encoder := json.NewEncoder(gzWriter)

	if err := encoder.Encode(m); err != nil {
		log.Printf("Ошибка сжатия метрики %s: %v", metricName, err)
		return
	}

	if err := gzWriter.Close(); err != nil {
		log.Printf("Ошибка закрытия gzip writer для метрики %s: %v", metricName, err)
		return
	}

	url := a.serverAddress + "/update/"

	var responseMetrics models.Metrics

	resp, err := a.client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(&buf).
		SetResult(&responseMetrics).
		Post(url)

	if err != nil {
		log.Printf("Ошибка отправки метрики %s (%s): %v", metricName, metricType, err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("Сервер ответил со статусом %s для метрики %s. Тело: %s", resp.Status(), metricName, resp.String())
		return
	}
}
