package JWT

import (
	"context"
	"presentator/internal/core/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type token interface {
	ValidateToken(token string) (*jwt.Token, bool)
	ParseToken(token string) (*jwt.Token, error)
	CreateToken(ctx context.Context, brandName, ip, role string) (string, error)
	CheckAdminAccess(token *jwt.Token, role string) bool
	LogOut(ctx context.Context, id string) error
	RetryDeleteToken(id string) error
	IsExist(ctx context.Context, jti string) bool
}

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
