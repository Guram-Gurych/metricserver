package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"go.uber.org/zap"
	"os"
	"time"
)

type storageFile struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

type Persister struct {
	repo     repository.MetricRepository
	filePath string
	logger   *zap.Logger
}

func NewPersister(repo repository.MetricRepository, filePath string, logger *zap.Logger) *Persister {
	return &Persister{
		repo:     repo,
		filePath: filePath,
		logger:   logger,
	}
}

func (p *Persister) Save() error {
	if p.filePath == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gauges := p.repo.GetAllGauges(ctx)
	counters := p.repo.GetAllCounters(ctx)

	storage := storageFile{Gauges: gauges, Counters: counters}
	storageJSON, err := json.Marshal(storage)
	if err != nil {
		return err
	}

	if err = os.WriteFile(p.filePath, storageJSON, 0644); err != nil {
		return err
	}

	return nil
}

func (p *Persister) Load() error {
	if p.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(p.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var storage storageFile
	if err = json.Unmarshal(data, &storage); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for key, value := range storage.Gauges {
		err = p.repo.UpdateGauge(ctx, key, value)
		if err != nil {
			return err
		}
	}

	for key, value := range storage.Counters {
		err = p.repo.UpdateCounter(ctx, key, value)
		if err != nil {
			return err
		}
	}

	return nil
}
