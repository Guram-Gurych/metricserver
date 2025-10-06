package agent

import (
	"fmt"
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
	client         *http.Client
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
		client:         &http.Client{},
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

	if _, ok := a.storage.Counters["PollCount"]; ok {
		a.storage.Counters["PollCount"] = 0
	}
}

func (a *Agent) sendMetric(metricType, metricName, metricValue string) {
	url := fmt.Sprintf("%s/update/%s/%s/%s", a.serverAddress, metricType, metricName, metricValue)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Printf("Ошибка при создании запроса gauge %s %s: %v", metricType, metricName, err)
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Printf("Ошибка отправки %s %s: %v", metricType, metricName, err)
		return
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Printf("Не удалось закрыть тело ответа: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Сервер ответил не 201 статус for gauge %s: %s", metricName, resp.Status)
	}
}
