package db

import (
	"context"
	"database/sql"
	"time"
)

func Initialize(DatabaseDSN string) (*sql.DB, error) {
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func InitializeSchema(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS metrics (
		id TEXT PRIMARY KEY,
		mtype TEXT NOT NULL CHECK (mtype IN ('gauge', 'counter')),
		value DOUBLE PRECISION,
		delta BIGINT
		);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
