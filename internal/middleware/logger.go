package middleware

import (
	"github.com/Guram-Gurych/metricserver.git/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWrite struct {
	http.ResponseWriter
	data *responseData
}

func (r *loggingResponseWrite) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.data.size += size
	return size, err
}

func (r *loggingResponseWrite) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.data.status = statusCode
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rd := &responseData{}
		lw := loggingResponseWrite{
			ResponseWriter: w,
			data:           rd,
		}

		next.ServeHTTP(&lw, r)

		logger.Log.Info("Request processed",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status_code", rd.status),
			zap.Int("response_size", rd.size),
		)
	})
}
