package fileservingService

import (
	"context"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"
	fileservingRepository "presentator/internal/features/fileserving/repository"

	"go.uber.org/zap"
)

type ServingRepo interface {
	GetAllWorks(ctx context.Context, brandName string) ([]entity.Works, error)
}

type ServingService struct {
	log *zap.Logger
	db  ServingRepo
	jwt *JWT.ServingJWT
}

func NewServingService(log *zap.Logger, jwt *JWT.ServingJWT, repo *fileservingRepository.ServingRepo) *ServingService {
	return &ServingService{
		log: log,
		db:  repo,
		jwt: jwt,
	}
}
