package authService

import (
	"context"
	authRepository "presentator/internal/features/auth/repository"
	"presentator/internal/features/auth/token"
)

type AuthRepo interface {
	GetPass(ctx context.Context, name string) (string, error)
}

type AuthService struct {
	rep AuthRepo
	jwt *JWT.ServingJWT
}

func NewAuthService(jwt *JWT.ServingJWT, rep *authRepository.AuthRepo) *AuthService {
	return &AuthService{
		rep: rep,
		jwt: jwt,
	}
}
