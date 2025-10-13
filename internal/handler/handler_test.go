package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockMemStorage struct {
	UpdateGaugeFunc   func(name string, value float64) error
	UpdateCounterFunc func(name string, value int64) error
}

func (m *MockMemStorage) UpdateGauge(name string, value float64) error {
	if m.UpdateGaugeFunc != nil {
		return m.UpdateGaugeFunc(name, value)
	}
	return nil
}

func (m *MockMemStorage) UpdateCounter(name string, value int64) error {
	if m.UpdateCounterFunc != nil {
		return m.UpdateCounterFunc(name, value)
	}
	return nil
}

func TestMetricHandler_UpdateMetric(t *testing.T) {
	mockRepo := &MockMemStorage{}
	handler := NewMetricHandler(mockRepo)

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "Success - Gauge Update",
			method:         http.MethodPost,
			url:            "/update/gauge/TestGauge/123.45",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Success - Counter Update",
			method:         http.MethodPost,
			url:            "/update/counter/TestCounter/123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error - Method GET Not Allowed",
			method:         http.MethodGet,
			url:            "/update/gauge/TestGauge/123.45",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Error - Malformed URL",
			method:         http.MethodPost,
			url:            "/update/gauge/TestGauge",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error - Empty Metric Name",
			method:         http.MethodPost,
			url:            "/update/gauge//123.45",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Error - Invalid Metric Type",
			method:         http.MethodPost,
			url:            "/update/unknown/TestMetric/100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Invalid Gauge Value",
			method:         http.MethodPost,
			url:            "/update/gauge/TestGauge/abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Invalid Counter Value",
			method:         http.MethodPost,
			url:            "/update/counter/TestCounter/123.45",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.url, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/update/{metricType}/{metricName}/{metricValue}", handler.UpdateMetric)
			router.ServeHTTP(rec, req)
			
			assert.Equal(t, test.expectedStatus, rec.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}
