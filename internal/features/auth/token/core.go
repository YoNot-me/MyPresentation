package JWT

import (
	"presentator/internal/core/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ServingJWT struct {
	rdb *redis.Client
	env *entity.Config
	log *zap.Logger
}

type JWT struct {
	Ip        string `json:"ip"`
	BrandName string `json:"brand_name"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

func NewServingJWT(rdb *redis.Client, env *entity.Config, log *zap.Logger) *ServingJWT {
	return &ServingJWT{
		rdb: rdb,
		env: env,
		log: log,
	}
}
