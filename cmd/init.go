package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"presentator/internal/core/logger"
	"presentator/internal/core/repository"

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

	err := godotenv.Load(".env")
	if err != nil {
		log.Print("Error loading .env file")
		return nil, nil, nil, nil, err
	}

	env := entity.Config{
		Addr:      os.Getenv("PRES_ADDR"),
		DBURL:     os.Getenv("PRES_DATABASE_URL"),
		JWTKey:    os.Getenv("PRES_JWT_SECRET"),
		RedisPass: os.Getenv("REDIS_PASSWORD"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
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

	return &env, db, logFile, rdb, nil
}
