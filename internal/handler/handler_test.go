package handler

import (
	models "github.com/Guram-Gurych/metricserver.git/internal/model"
	"github.com/Guram-Gurych/metricserver.git/internal/repository/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricHandler_Post(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		contentType    string
		setupMock      func(mockRepo *mocks.MockMetricRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - Gauge Update",
			method:      http.MethodPost,
			url:         "/update/gauge/TestGauge/123.45",
			body:        "",
			contentType: "text/plain",
			setupMock: func(mockRepo *mocks.MockMetricRepository) {
				mockRepo.EXPECT().UpdateGauge("TestGauge", 123.45).Return(nil)
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
			setupMock: func(mockRepo *mocks.MockMetricRepository) {
				mockRepo.EXPECT().UpdateCounter("TestCounter", int64(123)).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "Error - Method GET Not Allowed",
			method:         http.MethodGet,
			url:            "/update/gauge/TestGauge/123.45",
			setupMock:      func(mockRepo *mocks.MockMetricRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "Error - Empty Metric Name",
			method:         http.MethodPost,
			url:            "/update/gauge//123.45",
			setupMock:      func(mockRepo *mocks.MockMetricRepository) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "Success - Gauge Update",
			method:      http.MethodPost,
			url:         "/update/",
			body:        `{"id":"TestGaugeJSON","type":"gauge","value":123.45}`,
			contentType: "application/json",
			setupMock: func(mockRepo *mocks.MockMetricRepository) {
				gomock.InOrder(
					mockRepo.EXPECT().UpdateGauge("TestGaugeJSON", 123.45).Return(nil),
					mockRepo.EXPECT().GetGauge("TestGaugeJSON").Return(123.45, true),
				)
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
			setupMock: func(mockRepo *mocks.MockMetricRepository) {
				gomock.InOrder(
					mockRepo.EXPECT().UpdateCounter("TestCounterJSON", int64(123)).Return(nil),
					mockRepo.EXPECT().GetCounter("TestCounterJSON").Return(int64(123), true),
				)
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
			setupMock:      func(mockRepo *mocks.MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Invalid Metric Type",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestInvalid","type":"unknown","value":123}`,
			contentType:    "application/json",
			setupMock:      func(mockRepo *mocks.MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Missing Value",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestGauge","type":"gauge"}`,
			contentType:    "application/json",
			setupMock:      func(mockRepo *mocks.MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error - Missing Delta",
			method:         http.MethodPost,
			url:            "/update/",
			body:           `{"id":"TestCounter","type":"counter"}`,
			contentType:    "application/json",
			setupMock:      func(mockRepo *mocks.MockMetricRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockMetricRepository(ctrl)
			handler := NewMetricHandler(mockRepo, nil)
			test.setupMock(mockRepo)

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockMetricRepository(ctrl)
			handler := NewMetricHandler(mockRepo, nil)

			if test.mockMetricType == models.Gauge {
				mockRepo.EXPECT().GetGauge(test.mockMetricName).Return(test.mockGaugeValue, test.mockFound)
			}
			if test.mockMetricType == models.Counter {
				mockRepo.EXPECT().GetCounter(test.mockMetricName).Return(test.mockCounterValue, test.mockFound)
			}

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
			mockMetricType: "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid metric type\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockMetricRepository(ctrl)
			handler := NewMetricHandler(mockRepo, nil)

			if test.mockMetricType == "gauge" {
				mockRepo.EXPECT().GetGauge(test.mockMetricName).Return(test.mockGaugeValue, test.mockFound)
			}
			if test.mockMetricType == "counter" {
				mockRepo.EXPECT().GetCounter(test.mockMetricName).Return(test.mockCounterValue, test.mockFound)
			}

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
