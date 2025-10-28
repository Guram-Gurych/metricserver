package persistence

import (
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

func (ps *PersistentStorage) UpdateGauge(name string, value float64) error {
	err := ps.repo.UpdateGauge(name, value)

	if err == nil && ps.isSync {
		if saveErr := ps.persister.Save(); saveErr != nil {
			ps.persister.logger.Error("Sync save failed", zap.Error(saveErr))
		} else {
			return saveErr
		}
	}

	return err
}

func (ps *PersistentStorage) UpdateCounter(name string, value int64) error {
	err := ps.repo.UpdateCounter(name, value)

	if err == nil && ps.isSync {
		if saveErr := ps.persister.Save(); saveErr != nil {
			ps.persister.logger.Error("Sync save failed", zap.Error(saveErr))
		} else {
			return saveErr
		}
	}

	return err
}

func (ps *PersistentStorage) GetGauge(name string) (float64, bool) {
	return ps.repo.GetGauge(name)
}

func (ps *PersistentStorage) GetCounter(name string) (int64, bool) {
	return ps.repo.GetCounter(name)
}

func (ps *PersistentStorage) GetAllGauges() map[string]float64 {
	return ps.repo.GetAllGauges()
}

func (ps *PersistentStorage) GetAllCounters() map[string]int64 {
	return ps.repo.GetAllCounters()
}
