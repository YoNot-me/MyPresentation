package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenDatabase(ctx context.Context, url string) (*pgxpool.Pool, error) {

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 20
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	config.ConnConfig.ConnectTimeout = 5 * time.Second

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
