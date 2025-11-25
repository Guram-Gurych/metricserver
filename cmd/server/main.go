package main

import (
	"database/sql"
	"github.com/Guram-Gurych/metricserver.git/internal/config"
	"github.com/Guram-Gurych/metricserver.git/internal/config/db"
	"github.com/Guram-Gurych/metricserver.git/internal/handler"
	"github.com/Guram-Gurych/metricserver.git/internal/logger"
	"github.com/Guram-Gurych/metricserver.git/internal/middleware"
	"github.com/Guram-Gurych/metricserver.git/internal/persistence"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v4/stdlib"
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

	var dbConn *sql.DB
	var metricRepo repository.MetricRepository
	var err error

	if cnfg.DatabaseDSN != "" {
		dbConn, err = db.Initialize(cnfg.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Initialization error DB", zap.Error(err))
		}
		defer dbConn.Close()

		if err := db.InitializeSchema(dbConn); err != nil {
			logger.Log.Fatal("Schema initialization error", zap.Error(err))
		}

		logger.Log.Info("DB connection established and schema initialized")
		metricRepo = repository.NewDBRepository(dbConn)
	}

	if metricRepo == nil {
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

		metricRepo = storage
		if cnfg.FileStoragePath != "" && cnfg.StoreInterval == 0 {
			logger.Log.Info("Sync storage mode enabled (File persistence)")
			metricRepo = persistence.NewPersistentStorage(storage, persister, true)
		}

		if metricRepo == storage {
			logger.Log.Info("In-memory storage enabled.")
		}
	} else {
		logger.Log.Info("PostgreSQL storage enabled. File persistence logic skipped.")
	}

	metricHandler := handler.NewMetricHandler(metricRepo, dbConn)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.GzipMiddleware)
	r.Get("/", metricHandler.GetAllMetricsHTML)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", metricHandler.Post)
	r.Post("/value/", metricHandler.PostValue)
	r.Get("/value/{metricType}/{metricName}", metricHandler.Get)
	r.Post("/update/", metricHandler.Post)
	r.Get("/ping", metricHandler.GetPing)

	logger.Log.Info("Starting server", zap.String("address", cnfg.ServerAddress))

	if err := http.ListenAndServe(cnfg.ServerAddress, r); err != nil {
		logger.Log.Fatal("The server crashed", zap.Error(err))
	}
}
