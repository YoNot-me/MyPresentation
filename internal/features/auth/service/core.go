package authService

import (
	"context"
	authRepository "presentator/internal/features/auth/repository"
	"presentator/internal/features/auth/token"

	"go.uber.org/zap"
)

type AuthRepo interface {
	GetPass(ctx context.Context, name string) (string, error)
}

type AuthService struct {
	log *zap.Logger
	rep AuthRepo
	jwt *JWT.ServingJWT
}

func NewAuthService(log *zap.Logger, jwt *JWT.ServingJWT, rep *authRepository.AuthRepo) *AuthService {
	return &AuthService{
		log: log,
		rep: rep,
		jwt: jwt,
	}
}
