package adminService

import (
	"context"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"
)

type AdminRepository interface {
	ListAllBrands(ctx context.Context) ([]entity.Brand, error)
	AddNewBrand(ctx context.Context, brandName, hashPass string) error
	ChangeBrandPassword(ctx context.Context, brandName, newPass string) error
	DeleteBrand(ctx context.Context, brandName string) error

	ListAllWorks(ctx context.Context, brandName string) ([]entity.Works, error)
	AddNewWork(ctx context.Context, brandName, workName, url string) error
	DeleteWork(ctx context.Context, brandName, workName string) error
	ChangeWorkFields(ctx context.Context, brandName, workName string, work *entity.Works) error
	GetWork(ctx context.Context, workName string) (entity.Works, error)
}

type AdminService struct {
	rep AdminRepository
	jwt *JWT.ServingJWT
}

func NewAdminService(rep AdminRepository, jwt *JWT.ServingJWT) *AdminService {

	return &AdminService{
		rep: rep,
		jwt: jwt,
	}
}
