package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/config"
	"github.com/Guram-Gurych/metricserver.git/internal/handler"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	cnfg := config.InitConfigServer()
	storage := repository.NewMemStorage()
	metricHandler := handler.NewMetricHandler(storage)

	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Post)
	r.Get("/value/{metricType}/{metricName}", metricHandler.Get)
	r.Post("/update/", metricHandler.Post)

	if err := http.ListenAndServe(cnfg.ServerAddress, r); err != nil {
		panic(err)
	}
}
