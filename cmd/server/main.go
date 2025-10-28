package main

import (
	"github.com/Guram-Gurych/metricserver.git/internal/config"
	"github.com/Guram-Gurych/metricserver.git/internal/handler"
	"github.com/Guram-Gurych/metricserver.git/internal/logger"
	"github.com/Guram-Gurych/metricserver.git/internal/middleware"
	"github.com/Guram-Gurych/metricserver.git/internal/persistence"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	if err := logger.Initalize("info"); err != nil {
		panic(err)
	}
	defer logger.Log.Sync()

	cnfg := config.InitConfigServer()
	storage := repository.NewMemStorage()
	persister := persistence.NewPersister(storage, cnfg.FileStoragePath, logger.Log)

	if cnfg.Restore {
		if err := persister.Load(); err != nil {
			logger.Log.Error("Failed to load metrics from file", zap.Error(err))
		} else {
			logger.Log.Info("Metrics loaded from file", zap.String("file", cnfg.FileStoragePath))
		}
	}

	defer func() {
		logger.Log.Info("Shutting down, saving metrics...")
		if err := persister.Save(); err != nil {
			logger.Log.Error("Failed to save metrics on shutdown", zap.Error(err))
		} else {
			logger.Log.Info("Metrics saved on shutdown")
		}
	}()

	if cnfg.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(cnfg.StoreInterval)
			for range ticker.C {
				logger.Log.Debug("Saving metrics periodically")
				if err := persister.Save(); err != nil {
					logger.Log.Error("Failed to save metrics periodically", zap.Error(err))
				}
			}
		}()
	}

	var metricRepo repository.MetricRepository = storage
	if cnfg.StoreInterval == 0 {
		logger.Log.Info("Sync storage mode enabled")
		metricRepo = persistence.NewPersistentStorage(storage, persister, true)
	}

	metricHandler := handler.NewMetricHandler(metricRepo)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	r.Get("/", metricHandler.GetAllMetricsHTML)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Post)
	r.Post("/value/", metricHandler.PostValue)
	r.Get("/value/{metricType}/{metricName}", metricHandler.Get)
	r.Post("/update/", metricHandler.Post)

	logger.Log.Info("Starting server", zap.String("address", cnfg.ServerAddress))

	if err := http.ListenAndServe(cnfg.ServerAddress, r); err != nil {
		logger.Log.Fatal("The server crashed", zap.Error(err))
	}
}
