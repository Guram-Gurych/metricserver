package repository

//go:generate mockgen -source=interface.go -destination=mocks/mock_repository.go -package=mocks
type MetricRepository interface {
	UpdateGauge(name string, value float64) error
	UpdateCounter(name string, value int64) error
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetAllGauges() map[string]float64
	GetAllCounters() map[string]int64
}
