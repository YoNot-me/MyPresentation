package adminService

import (
	"context"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"
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
