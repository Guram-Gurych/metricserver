package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/handler"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"net/http"
)

func main() {
	storage := repository.NewMemStorage()
	metricHandler := handler.NewMetricHandler(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", metricHandler.UpdateMetric)

	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		panic(err)
	}
}
