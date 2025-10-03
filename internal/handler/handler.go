package handler

import (
	"github.com/Guram-Gurych/metricserver.git/internal/model"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

type MetricHandler struct {
	repo repository.MetricRepository
}

func NewMetricHandler(repo repository.MetricRepository) *MetricHandler {
	return &MetricHandler{
		repo: repo,
	}
}

func (h *MetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
	}

	parts := strings.Split(r.URL.Path, "/")
	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	if metricName == "" {
		http.Error(w, "Not Found: Metric name is required", http.StatusNotFound)
		return
	}

	var err error
	switch metricType {
	case models.Gauge:
		value, parseErr := strconv.ParseFloat(metricValue, 64)
		if parseErr != nil {
			http.Error(w, "Bad Request: Invalid gauge value", http.StatusBadRequest)
			return
		}
		err = h.repo.UpdateGauge(metricName, value)
	case models.Counter:
		value, parseErr := strconv.ParseInt(metricValue, 10, 64)
		if parseErr != nil {
			http.Error(w, "Bad Request: Invalid counter value", http.StatusBadRequest)
			return
		}
		err = h.repo.UpdateCounter(metricName, value)
	default:
		http.Error(w, "Bad Request: Invalid metric type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
