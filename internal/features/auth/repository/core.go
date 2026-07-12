package authRepository

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type AuthRepo struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewAuthRepo(rdb *redis.Client, db *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{
		rdb: rdb,
		db:  db,
	}
}
