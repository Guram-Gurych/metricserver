package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/config"
	"github.com/Guram-Gurych/metricserver.git/internal/handler"
	"github.com/Guram-Gurych/metricserver.git/internal/logger"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	if err := logger.Initalize("info"); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	cnfg := config.InitConfigServer()
	storage := repository.NewMemStorage()
	metricHandler := handler.NewMetricHandler(storage)

	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Post)
	r.Post("/value/", metricHandler.PostValue)
	r.Get("/value/{metricType}/{metricName}", metricHandler.Get)
	r.Post("/update/", metricHandler.Post)

	logger.Log.Info("Starting server", zap.String("address", cnfg.ServerAddress))

	if err := http.ListenAndServe(cnfg.ServerAddress, r); err != nil {
		logger.Log.Fatal("The server crashed", zap.Error(err))
	}
}
