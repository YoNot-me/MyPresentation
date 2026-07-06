package fileservingRepository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServingRepo struct {
	db *pgxpool.Pool
}

func NewServingRepo(db *pgxpool.Pool) *ServingRepo {
	return &ServingRepo{
		db: db,
	}
}
