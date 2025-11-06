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
