package fileservingService

import (
	"context"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"
	fileservingRepository "presentator/internal/features/fileserving/repository"
)

type ServingRepo interface {
	GetAllWorks(ctx context.Context, brandName string) ([]entity.Works, error)
}

type ServingService struct {
	db  ServingRepo
	jwt *JWT.ServingJWT
}

func NewServingService(jwt *JWT.ServingJWT, repo *fileservingRepository.ServingRepo) *ServingService {
	return &ServingService{
		db:  repo,
		jwt: jwt,
	}
}
