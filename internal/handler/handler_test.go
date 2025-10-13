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
	GetGaugeFunc      func(name string) (float64, bool)
	GetCounterFunc    func(name string) (int64, bool)
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

func (m *MockMemStorage) GetGauge(name string) (float64, bool) {
	if m.GetGaugeFunc != nil {
		return m.GetGaugeFunc(name)
	}
	return 0, false
}

func (m *MockMemStorage) GetCounter(name string) (int64, bool) {
	if m.GetCounterFunc != nil {
		return m.GetCounterFunc(name)
	}
	return 0, false
}

func TestMetricHandler_Post(t *testing.T) {
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
			router.Post("/update/{metricType}/{metricName}/{metricValue}", handler.Post)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.expectedStatus, rec.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestMetricHandler_Get(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		mockMetricName   string
		mockMetricType   string
		mockGaugeValue   float64
		mockCounterValue int64
		mockFound        bool
		expectedStatus   int
		expectedBody     string
	}{
		{
			name:           "Успешное получение gauge",
			url:            "/value/gauge/TestGauge",
			mockMetricName: "TestGauge",
			mockMetricType: "gauge",
			mockGaugeValue: 123.456,
			mockFound:      true,
			expectedStatus: http.StatusOK,
			expectedBody:   "123.456",
		},
		{
			name:             "Успешное получение counter",
			url:              "/value/counter/TestCounter",
			mockMetricName:   "TestCounter",
			mockMetricType:   "counter",
			mockCounterValue: 789,
			mockFound:        true,
			expectedStatus:   http.StatusOK,
			expectedBody:     "789",
		},
		{
			name:           "Метрика не найдена",
			url:            "/value/gauge/NotFoundMetric",
			mockMetricName: "NotFoundMetric",
			mockMetricType: "gauge",
			mockFound:      false,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metric not found\n",
		},
		{
			name:           "Неверный тип метрики",
			url:            "/value/invalidType/SomeMetric",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := &MockMemStorage{}
			if test.mockMetricType == "gauge" {
				mockRepo.GetGaugeFunc = func(name string) (float64, bool) {
					if name == test.mockMetricName {
						return test.mockGaugeValue, test.mockFound
					}
					return 0, false
				}
			}
			if test.mockMetricType == "counter" {
				mockRepo.GetCounterFunc = func(name string) (int64, bool) {
					if name == test.mockMetricName {
						return test.mockCounterValue, test.mockFound
					}
					return 0, false
				}
			}

			handler := NewMetricHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get("/value/{metricType}/{metricName}", handler.Get)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.expectedStatus, rec.Code, "Код ответа не совпадает")
			assert.Equal(t, test.expectedBody, rec.Body.String(), "Тело ответа не совпадает")
		})
	}
}
