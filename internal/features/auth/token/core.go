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
	CreateToken(ctx context.Context, brandName, ip, role string) (string, error)
	CheckAccess(token *jwt.Token, ip string) bool
	CheckAdminAccess(token *jwt.Token, ip, role string) bool
	LogOut(ctx context.Context, ip string) error
	RetryDeleteToken(ip string) error
	IsExist(ctx context.Context, ip string) bool
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
