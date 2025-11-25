package repository

import (
	"context"
	"sync"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateGauge(_ context.Context, name string, value float64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.gauges[name] = value
	return nil
}

func (ms *MemStorage) UpdateCounter(_ context.Context, name string, value int64) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.counters[name] += value
	return nil
}

func (ms *MemStorage) GetGauge(_ context.Context, name string) (float64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, ok := ms.gauges[name]
	return val, ok
}

func (ms *MemStorage) GetCounter(_ context.Context, name string) (int64, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	val, ok := ms.counters[name]
	return val, ok
}

func (ms *MemStorage) GetAllGauges(_ context.Context) map[string]float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make(map[string]float64, len(ms.gauges))
	for k, v := range ms.gauges {
		result[k] = v
	}

	return result
}

func (ms *MemStorage) GetAllCounters(_ context.Context) map[string]int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make(map[string]int64, len(ms.counters))
	for k, v := range ms.counters {
		result[k] = v
	}

	return result
}
