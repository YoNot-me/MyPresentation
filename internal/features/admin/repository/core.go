package adminRepository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AdminRepo struct {
	rdb *redis.Client
	db  *pgxpool.Pool
}

func NewAdminRepo(rdb *redis.Client, db *pgxpool.Pool) *AdminRepo {
	return &AdminRepo{
		rdb: rdb,
		db:  db,
	}
}
