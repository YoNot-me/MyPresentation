package adminTransport

import (
	"context"
	"presentator/internal/core/entity"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminTrnsprt interface {
	AuthAdmin(c *gin.Context) // POST			admin/auth
	LogOut(c *gin.Context)    // POST			admin/logout

	ListBrands(c *gin.Context)          // GET 			admin/brands
	AddNewBrand(c *gin.Context)         // POST JSON 	admin/brands/add
	RenameBrand(c *gin.Context)         // PUT 	JSON	admin/:brandName/rename
	DeleteBrand(c *gin.Context)         // DELETE 		admin/:brandName
	ChangeBrandPassword(c *gin.Context) // PUT JSON 	admin/:brandName/password

	ListAllBrandWorks(c *gin.Context) // LIST 		admin/:brandName/works
	AddNewWork(c *gin.Context)        // POST JSON 	admin/:brandName/works/add
	DeleteWork(c *gin.Context)        // DELETE 		admin/:brandName/:workName
	ChangeWorkFields(c *gin.Context)  // PUT JSON 	admin/:brandName/:workName/change

	ServingWork(c *gin.Context) // GET 	JSON 	admin/:brandName/:workName
}

type AdminService interface {
	AuthAdmin(ctx context.Context, ip string, req *entity.Admin) (string, error)
	LogOut(ctx context.Context, ip string) error

	ListAllBrands(ctx context.Context) ([]entity.BrandsResponse, error)
	AddNewBrand(ctx context.Context, brand *entity.Brand) error
	RenameBrand(ctx context.Context, brandName string, newInfo *entity.Brand) error
	DeleteBrand(ctx context.Context, brandName string) error
	ChangeBrandPassword(ctx context.Context, brand *entity.Brand) error

	AddNewWork(ctx context.Context, brandName string, req *entity.Works, c *gin.Context) (int, error)
	ListAllWorks(ctx context.Context, brandName string) ([]entity.WorksResponse, error)
	DeleteWork(ctx context.Context, brandName, workName string) error
	ChangeWorkFields(ctx context.Context, brandName, workName string, work *entity.Works, c *gin.Context) error
	RenameFolders(brandName, workName string, newWorkName string) error
}

type AdminTransport struct {
	log *zap.Logger
	srv AdminService
}

func NewAdminTransport(log *zap.Logger, srv AdminService) *AdminTransport {
	return &AdminTransport{
		log: log,
		srv: srv,
	}
}
