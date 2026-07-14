package adminService

import (
	"context"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"

	"go.uber.org/zap"
)

type AdminRepository interface {
	ListAllBrands(ctx context.Context) ([]entity.BrandsResponse, error)
	AddNewBrand(ctx context.Context, brandName, hashPass string) error
	RenameBrand(ctx context.Context, brandName, newName string) error
	ChangeBrandPassword(ctx context.Context, brandName, newPass string) error
	DeleteBrand(ctx context.Context, brandName string) error

	ListAllWorks(ctx context.Context, brandName string) ([]entity.WorksResponse, error)
	AddNewWork(ctx context.Context, req *entity.Works) error
	DeleteWork(ctx context.Context, brandName, workName string) error
	ChangeWorkFields(ctx context.Context, brandName, workName string, work *entity.Works) error
	GetWork(ctx context.Context, brandName, workName string) (entity.Works, error)
	IsWorkExist(ctx context.Context, brandName, workName string) (bool, error)

	BruteCount(ctx context.Context, ip string) (int, error)
	IncCount(ctx context.Context, ip string) error
}

type AdminService struct {
	log *zap.Logger
	rep AdminRepository
	jwt *JWT.ServingJWT
}

func NewAdminService(log *zap.Logger, rep AdminRepository, jwt *JWT.ServingJWT) *AdminService {

	return &AdminService{
		log: log,
		rep: rep,
		jwt: jwt,
	}
}
