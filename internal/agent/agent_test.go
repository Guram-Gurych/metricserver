package agent

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestAgent_pollMetrics(t *testing.T) {
	a := NewAgent("http://localhost:8080", 1*time.Second, 2*time.Second)

	a.pollMetrics()

	assert.Equal(t, int64(1), a.storage.Counters["PollCount"], "PollCount должно быть 1 после одного polls")
	assert.Contains(t, a.storage.Gauges, "Alloc", "Gauges должен содержать метрику Alloc")
	assert.Contains(t, a.storage.Gauges, "RandomValue", "Gauges должен содержать метрику RandomValue")

	a.pollMetrics()
	assert.Equal(t, int64(2), a.storage.Counters["PollCount"], "PollCount должно быть 2 после второго polls")
}

func TestAgent_reportMetrics(t *testing.T) {
	var receivedRequests []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		receivedRequests = append(receivedRequests, r.URL.String())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testCases := []struct {
		name              string
		initialStorage    *AgentMetric
		expectedRequests  []string
		expectedPollCount int64
	}{
		{
			name: "Смешанные gauge and counter метрики",
			initialStorage: &AgentMetric{
				Gauges:   map[string]float64{"TestGauge": 123.45},
				Counters: map[string]int64{"PollCount": 5, "TestCounter": 10},
			},
			expectedRequests: []string{
				"/update/gauge/TestGauge/123.45",
				"/update/counter/PollCount/5",
				"/update/counter/TestCounter/10",
			},
			expectedPollCount: 0,
		},
		{
			name: "Только gauges",
			initialStorage: &AgentMetric{
				Gauges:   map[string]float64{"Alloc": 500.5, "Sys": 1024},
				Counters: make(map[string]int64),
			},
			expectedRequests: []string{
				"/update/gauge/Alloc/500.5",
				"/update/gauge/Sys/1024",
			},
			expectedPollCount: 0,
		},
		{
			name: "Только PollCount",
			initialStorage: &AgentMetric{
				Gauges:   make(map[string]float64),
				Counters: map[string]int64{"PollCount": 10},
			},
			expectedRequests: []string{
				"/update/counter/PollCount/10",
			},
			expectedPollCount: 0,
		},
		{
			name: "Пустое хранилище",
			initialStorage: &AgentMetric{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			},
			expectedRequests:  []string{},
			expectedPollCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			receivedRequests = nil
			mu.Unlock()

			agent := NewAgent(server.URL, 1*time.Second, 2*time.Second)
			agent.storage = tc.initialStorage

			agent.reportMetrics()

			require.Len(t, receivedRequests, len(tc.expectedRequests), "Количество запросов не совпадает")
			assert.ElementsMatch(t, tc.expectedRequests, receivedRequests, "URL запросов не совпадают с ожидаемыми")

			if _, ok := tc.initialStorage.Counters["PollCount"]; ok {
				assert.Equal(t, tc.expectedPollCount, agent.storage.Counters["PollCount"])
			}
		})
	}
}
