package main

import (
	"context"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"presentator/internal/core/logger"
	"presentator/internal/core/repository"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func initFeatures(ctx context.Context) (
	*entity.Config,
	*pgxpool.Pool,
	*zap.Logger,
	*redis.Client,
	error,
) {

	_ = godotenv.Load(".env")

	env := entity.Config{
		Addr:      os.Getenv("PRES_ADDR"),
		DBURL:     os.Getenv("PRES_DATABASE_URL"),
		JWTKey:    os.Getenv("PRES_JWT_SECRET"),
		RedisPass: os.Getenv("REDIS_PASSWORD"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
		Issuer:    os.Getenv("PRES_JWT_ISSUER"),
	}

	db, err := repository.OpenDatabase(ctx, env.DBURL)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = os.MkdirAll("./out", 0755)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	file := filepath.Join(".", "out", "logger.log")
	openFile, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	logFile := logger.InitLogger(openFile)

	rdb := redis.NewClient(&redis.Options{
		Addr:     env.RedisAddr,
		Password: env.RedisPass,
		DB:       0,
		Protocol: 2,
	})

	err = pingRdb(rdb)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return &env, db, logFile, rdb, nil
}

func pingRdb(rdb *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}

	return nil
}
