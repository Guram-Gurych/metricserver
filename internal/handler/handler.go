package handler

import (
	"encoding/json"
	"fmt"
	"github.com/Guram-Gurych/metricserver.git/internal/logger"
	"github.com/Guram-Gurych/metricserver.git/internal/model"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sort"
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

func (h *MetricHandler) Post(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			h.handlePostJSON(w, r)
			return
		}
	}

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")

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

func (h *MetricHandler) handlePostJSON(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var err error
	switch metrics.MType {
	case models.Gauge:
		if metrics.Value == nil {
			http.Error(w, "Bad Request: Invalid gauge value", http.StatusBadRequest)
			return
		}
		err = h.repo.UpdateGauge(metrics.ID, *metrics.Value)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		newValue, ok := h.repo.GetGauge(metrics.ID)
		if !ok {
			http.Error(w, "Internal Server Error after update", http.StatusInternalServerError)
			return
		}
		metrics.Value = &newValue

	case models.Counter:
		if metrics.Delta == nil {
			http.Error(w, "Bad Request: Invalid counter value", http.StatusBadRequest)
			return
		}
		err = h.repo.UpdateCounter(metrics.ID, *metrics.Delta)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		newDelta, ok := h.repo.GetCounter(metrics.ID)
		if !ok {
			http.Error(w, "Internal Server Error after update", http.StatusInternalServerError)
			return
		}
		metrics.Delta = &newDelta
	default:
		http.Error(w, "Bad Request: Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

func (h *MetricHandler) PostValue(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	var metrics models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	switch metrics.MType {
	case models.Gauge:
		value, ok := h.repo.GetGauge(metrics.ID)
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		metrics.Value = &value
	case models.Counter:
		delta, ok := h.repo.GetCounter(metrics.ID)
		if !ok {
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		metrics.Delta = &delta
	default:
		http.Error(w, "Bad Request: Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
	}
}

func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	var valueStr string
	var ok bool

	switch metricType {
	case models.Gauge:
		var value float64

		value, ok = h.repo.GetGauge(metricName)
		if ok {
			valueStr = strconv.FormatFloat(value, 'f', -1, 64)
		}
	case models.Counter:
		var value int64

		value, ok = h.repo.GetCounter(metricName)
		if ok {
			valueStr = strconv.FormatInt(value, 10)
		}
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if !ok {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(valueStr))
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
}

func (h *MetricHandler) GetAllMetricsHTML(w http.ResponseWriter, r *http.Request) {
	gauges := h.repo.GetAllGauges()
	counters := h.repo.GetAllCounters()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	gaugeNames := make([]string, 0, len(gauges))
	for k := range gauges {
		gaugeNames = append(gaugeNames, k)
	}
	sort.Strings(gaugeNames)

	counterNames := make([]string, 0, len(counters))
	for k := range counters {
		counterNames = append(counterNames, k)
	}
	sort.Strings(counterNames)

	io.WriteString(w, "<html><head><title>Metrics</title></head><body>")
	io.WriteString(w, "<h1>Metrics</h1>")
	io.WriteString(w, "<h2>Gauges</h2><ul>")
	for _, name := range gaugeNames {
		io.WriteString(w, fmt.Sprintf("<li>%s: %f</li>", name, gauges[name]))
	}
	io.WriteString(w, "</ul>")

	io.WriteString(w, "<h2>Counters</h2><ul>")
	for _, name := range counterNames {
		io.WriteString(w, fmt.Sprintf("<li>%s: %d</li>", name, counters[name]))
	}
	io.WriteString(w, "</ul>")

	io.WriteString(w, "</body></html>")
}
