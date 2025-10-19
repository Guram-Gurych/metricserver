package handler

import (
	models "github.com/Guram-Gurych/metricserver.git/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		body           string
		contentType    string
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - Gauge Update",
			method:      http.MethodPost,
			url:         "/update/gauge/TestGauge/123.45",
			body:        "",
			contentType: "text/plain",
			setupMock: func() {
				mockRepo.UpdateGaugeFunc = func(name string, value float64) error {
					assert.Equal(t, "TestGauge", name)
					assert.Equal(t, 123.45, value)
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:        "Success - Counter Update",
			method:      http.MethodPost,
			url:         "/update/counter/TestCounter/123",
			body:        "",
			contentType: "text/plain",
			setupMock: func() {
				mockRepo.UpdateCounterFunc = func(name string, value int64) error {
					assert.Equal(t, "TestCounter", name)
					assert.Equal(t, int64(123), value)
					return nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Error - Method GET Not Allowed",
			method:         http.MethodGet,
			url:            "/update/gauge/TestGauge/123.45",
			setupMock:      func() {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Error - Empty Metric Name",
			method:         http.MethodPost,
			url:            "/update/gauge//123.45",
			setupMock:      func() {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "Success - Gauge Update",
			method:      http.MethodPost,
			url:         "/update/",
			body:        `{"id":"TestGaugeJSON","type":"gauge","value":123.45}`,
			contentType: "application/json",
			setupMock: func() {
				mockRepo.UpdateGaugeFunc = func(name string, value float64) error {
					assert.Equal(t, "TestGaugeJSON", name)
					assert.Equal(t, 123.45, value)
					return nil
				}
				mockRepo.GetGaugeFunc = func(name string) (float64, bool) {
					assert.Equal(t, "TestGaugeJSON", name)
					return 123.45, true
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"TestGaugeJSON","type":"gauge","value":123.45}`,
		},
		{
			name:        "Success - Counter Update",
			method:      http.MethodPost,
			url:         "/update/",
			body:        `{"id":"TestCounterJSON","type":"counter","delta":123}`,
			contentType: "application/json",
			setupMock: func() {
				mockRepo.UpdateCounterFunc = func(name string, value int64) error {
					assert.Equal(t, "TestCounterJSON", name)
					assert.Equal(t, int64(123), value)
					return nil
				}
				mockRepo.GetCounterFunc = func(name string) (int64, bool) {
					assert.Equal(t, "TestCounterJSON", name)
					return 123, true
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"TestCounterJSON","type":"counter","delta":123}`,
		},
		{
			name:           "Error - Bad Request",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestGauge",`,
			contentType:    "application/json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Invalid Metric Type",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestInvalid","type":"unknown","value":123}`,
			contentType:    "application/json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Missing Value",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestGauge","type":"gauge"}`,
			contentType:    "application/json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Missing Delta",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestCounter","type":"counter"}`,
			contentType:    "application/json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo.UpdateGaugeFunc = nil
			mockRepo.UpdateCounterFunc = nil
			mockRepo.GetGaugeFunc = nil
			mockRepo.GetCounterFunc = nil
			test.setupMock()

			var reqBody io.Reader
			if test.body != "" {
				reqBody = strings.NewReader(test.body)
			}

			req := httptest.NewRequest(test.method, test.url, reqBody)
			if test.contentType != "" {
				req.Header.Set("Content-Type", test.contentType)
			}

			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/update/{metricType}/{metricName}/{metricValue}", handler.Post)
			router.Post("/update/", handler.Post)

			router.ServeHTTP(rec, req)

			assert.Equal(t, test.expectedStatus, rec.Code, "Код ответа не совпадает с ожидаемым")

			if test.expectedBody != "" {
				assert.JSONEq(t, test.expectedBody, rec.Body.String(), "Тело ответа не совпадает")
			}
		})
	}
}

func TestMetricHandler_PostValue(t *testing.T) {
	tests := []struct {
		name             string
		body             string
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
			body:           `{"id":"TestGauge","type":"gauge"}`,
			mockMetricName: "TestGauge",
			mockMetricType: models.Gauge,
			mockGaugeValue: 123.456,
			mockFound:      true,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"TestGauge","type":"gauge","value":123.456}`,
		},
		{
			name:             "Успешное получение counter",
			body:             `{"id":"TestCounter","type":"counter"}`,
			mockMetricName:   "TestCounter",
			mockMetricType:   models.Counter,
			mockCounterValue: 789,
			mockFound:        true,
			expectedStatus:   http.StatusOK,
			expectedBody:     `{"id":"TestCounter","type":"counter","delta":789}`,
		},
		{
			name:           "Метрика не найдена",
			body:           `{"id":"NotFoundMetric","type":"gauge"}`,
			mockMetricName: "NotFoundMetric",
			mockMetricType: models.Gauge,
			mockFound:      false,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "",
		},
		{
			name:           "Неверный тип метрики",
			body:           `{"id":"SomeMetric","type":"invalidType"}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "Битый JSON",
			body:           `{"id":"SomeMetric",`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := &MockMemStorage{}

			if test.mockMetricType == models.Gauge {
				mockRepo.GetGaugeFunc = func(name string) (float64, bool) {
					if name == test.mockMetricName {
						return test.mockGaugeValue, test.mockFound
					}
					return 0, false
				}
			}
			if test.mockMetricType == models.Counter {
				mockRepo.GetCounterFunc = func(name string) (int64, bool) {
					if name == test.mockMetricName {
						return test.mockCounterValue, test.mockFound
					}
					return 0, false
				}
			}

			handler := NewMetricHandler(mockRepo)

			reqBody := strings.NewReader(test.body)
			req := httptest.NewRequest(http.MethodPost, "/value/", reqBody)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Post("/value/", handler.PostValue)
			router.ServeHTTP(rec, req)

			assert.Equal(t, test.expectedStatus, rec.Code, "Код ответа не совпадает")

			if test.expectedBody != "" {
				assert.JSONEq(t, test.expectedBody, rec.Body.String(), "Тело ответа не совпадает")
			}
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
