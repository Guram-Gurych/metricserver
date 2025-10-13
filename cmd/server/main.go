package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/handler"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	storage := repository.NewMemStorage()
	metricHandler := handler.NewMetricHandler(storage)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.UpdateMetric)
	r.Post("/update/", metricHandler.UpdateMetric)

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}
}
