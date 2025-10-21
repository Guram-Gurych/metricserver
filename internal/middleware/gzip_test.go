package middleware

import (
	"bytes"
	"compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	type testCase struct {
		name                  string
		headers               map[string]string
		requestBody           string
		compressRequestBody   bool
		handlerContentType    string
		expectedStatusCode    int
		expectResponseGzipped bool
		expectedResponseBody  string
	}

	requestBodyJSON := `{"id":"TestMetric","type":"gauge","value":123.45}`

	tests := []testCase{
		{
			name:                  "sends gzipped request, gets plain response",
			headers:               map[string]string{"Content-Encoding": "gzip", "Content-Type": "application/json"},
			requestBody:           requestBodyJSON,
			compressRequestBody:   true,
			handlerContentType:    "application/json",
			expectedStatusCode:    http.StatusOK,
			expectResponseGzipped: false,
			expectedResponseBody:  requestBodyJSON,
		},
		{
			name:                  "sends plain request, accepts gzipped json response",
			headers:               map[string]string{"Accept-Encoding": "gzip", "Content-Type": "application/json"},
			requestBody:           requestBodyJSON,
			compressRequestBody:   false,
			handlerContentType:    "application/json",
			expectedStatusCode:    http.StatusOK,
			expectResponseGzipped: true,
			expectedResponseBody:  requestBodyJSON,
		},
		{
			name:                  "sends plain request, accepts gzipped html response",
			headers:               map[string]string{"Accept-Encoding": "gzip"},
			requestBody:           "<h1>Hello</h1>",
			compressRequestBody:   false,
			handlerContentType:    "text/html",
			expectedStatusCode:    http.StatusOK,
			expectResponseGzipped: true,
			expectedResponseBody:  "<h1>Hello</h1>",
		},
		{
			name:                  "sends plain request, but handler content type is not compressible",
			headers:               map[string]string{"Accept-Encoding": "gzip"},
			requestBody:           "some plain text",
			compressRequestBody:   false,
			handlerContentType:    "text/plain",
			expectedStatusCode:    http.StatusOK,
			expectResponseGzipped: false,
			expectedResponseBody:  "some plain text",
		},
		{
			name:                  "no gzip involved at all",
			headers:               map[string]string{"Content-Type": "application/json"},
			requestBody:           requestBodyJSON,
			compressRequestBody:   false,
			handlerContentType:    "application/json",
			expectedStatusCode:    http.StatusOK,
			expectResponseGzipped: false,
			expectedResponseBody:  requestBodyJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, test.requestBody, string(body))

				w.Header().Set("Content-Type", test.handlerContentType)
				w.WriteHeader(test.expectedStatusCode)
				_, err = w.Write([]byte(test.expectedResponseBody))
				require.NoError(t, err)
			})

			handlerToTest := GzipMiddleware(dummyHandler)
			srv := httptest.NewServer(handlerToTest)
			defer srv.Close()

			var body io.Reader
			if test.compressRequestBody {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)
				_, err := gz.Write([]byte(test.requestBody))

				require.NoError(t, err)

				err = gz.Close()
				require.NoError(t, err)
				body = &buf
			} else {
				body = bytes.NewBufferString(test.requestBody)
			}

			req, err := http.NewRequest(http.MethodPost, srv.URL, body)
			require.NoError(t, err)
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)

			var respBody []byte
			if test.expectResponseGzipped {
				assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
				gz, err := gzip.NewReader(resp.Body)
				require.NoError(t, err)
				respBody, err = io.ReadAll(gz)
				require.NoError(t, err)
			} else {
				assert.Empty(t, resp.Header.Get("Content-Encoding"))
				respBody, err = io.ReadAll(resp.Body)
				require.NoError(t, err)
			}

			assert.Equal(t, test.expectedResponseBody, string(respBody))
		})
	}
}
