package persistence

import (
	"context"
	"github.com/Guram-Gurych/metricserver.git/internal/repository"
	"go.uber.org/zap"
)

type PersistentStorage struct {
	repo      repository.MetricRepository
	persister *Persister
	isSync    bool
}

func NewPersistentStorage(repo repository.MetricRepository, persister *Persister, storeInterval bool) *PersistentStorage {
	return &PersistentStorage{repo: repo, persister: persister, isSync: storeInterval}
}

func (ps *PersistentStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	err := ps.repo.UpdateGauge(ctx, name, value)
	if err != nil {
		return err
	}

	if ps.isSync {
		if saveErr := ps.persister.Save(); saveErr != nil {
			ps.persister.logger.Error("Sync save failed", zap.Error(saveErr))
		}
	}

	return err
}

func (ps *PersistentStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	err := ps.repo.UpdateCounter(ctx, name, value)
	if err != nil {
		return err
	}

	if ps.isSync {
		if saveErr := ps.persister.Save(); saveErr != nil {
			ps.persister.logger.Error("Sync save failed", zap.Error(saveErr))
		}
	}

	return err
}

func (ps *PersistentStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	return ps.repo.GetGauge(ctx, name)
}

func (ps *PersistentStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	return ps.repo.GetCounter(ctx, name)
}

func (ps *PersistentStorage) GetAllGauges(ctx context.Context) map[string]float64 {
	return ps.repo.GetAllGauges(ctx)
}

func (ps *PersistentStorage) GetAllCounters(ctx context.Context) map[string]int64 {
	return ps.repo.GetAllCounters(ctx)
}
