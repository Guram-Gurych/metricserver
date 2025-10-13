package handler

import (
	"github.com/Guram-Gurych/metricserver.git/internal/model"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
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

func (h *MetricHandler) Get(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	var valueStr string
	var ok bool

	switch metricType {
	case models.Gauge:
		value, ok := h.repo.GetGauge(metricName)
		if ok {
			valueStr = strconv.FormatFloat(value, 'f', -1, 64)
		}
	case models.Counter:
		value, ok := h.repo.GetCounter(metricName)
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
