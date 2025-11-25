package repository

import (
	"context"
	"database/sql"
)

type dbRepository struct {
	db *sql.DB
}

func NewDBRepository(db *sql.DB) *dbRepository {
	return &dbRepository{db: db}
}

func (db *dbRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	query := `
		INSERT INTO metrics (id, mtype, value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (id)
		DO UPDATE SET
		   value = EXCLUDED.value,
		   mtype = 'gauge'
	`
	_, err := db.db.ExecContext(ctx, query, name, value)
	return err
}

func (db *dbRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO metrics (id, mtype, delta)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (id)
		DO UPDATE SET
		   delta = metrics.delta + EXCLUDED.delta,
		   mtype = 'counter'
	`
	_, err := db.db.ExecContext(ctx, query, name, value)
	return err
}
func (db *dbRepository) GetGauge(ctx context.Context, name string) (float64, bool) {
	var value float64
	query := "SELECT value FROM metrics WHERE id = $1 and mtype = 'gauge'"

	err := db.db.QueryRowContext(ctx, query, name).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, false
	} else if err != nil {
		return 0, false
	}

	return value, true
}
func (db *dbRepository) GetCounter(ctx context.Context, name string) (int64, bool) {
	var delta int64
	query := "SELECT delta FROM metrics WHERE id = $1 and mtype = 'counter'"

	err := db.db.QueryRowContext(ctx, query, name).Scan(&delta)
	if err == sql.ErrNoRows {
		return 0, false
	} else if err != nil {
		return 0, false
	}

	return delta, true
}
func (db *dbRepository) GetAllGauges(ctx context.Context) map[string]float64 {
	gauges := make(map[string]float64)
	query := "SELECT id, value FROM metrics WHERE mtype = 'gauge'"
	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return gauges
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var value float64
		if err := rows.Scan(&id, &value); err != nil {
			continue
		}
		gauges[id] = value
	}

	if err := rows.Err(); err != nil {
		return gauges
	}

	return gauges
}
func (db *dbRepository) GetAllCounters(ctx context.Context) map[string]int64 {
	counters := make(map[string]int64)
	query := "SELECT id, delta FROM metrics WHERE mtype = 'counter'"
	rows, err := db.db.QueryContext(ctx, query)
	if err != nil {
		return counters
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var delta int64
		if err := rows.Scan(&id, &delta); err != nil {
			continue
		}
		counters[id] = delta
	}

	if err := rows.Err(); err != nil {
		return counters
	}

	return counters
}
