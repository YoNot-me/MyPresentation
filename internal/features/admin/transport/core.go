package adminTransport

import (
	"context"
	"presentator/internal/core/entity"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminTrnsprt interface {
	AuthAdmin(c *gin.Context) // POST				admin/auth
	LogOut(c *gin.Context)    // POST				admin/logout

	ListBrands(c *gin.Context)          // GET 				admin/brands
	AddNewBrand(c *gin.Context)         // POST JSON 		admin/brands/add
	DeleteBrand(c *gin.Context)         // DELETE 			admin/:brandName
	ChangeBrandPassword(c *gin.Context) // PUT JSON 		admin/:brandName/password

	ListAllBrandWorks(c *gin.Context) // LIST 			admin/:brandName/works
	AddNewWork(c *gin.Context)        // POST JSON 		admin/:brandName/:workName/add
	DeleteWork(c *gin.Context)        // DELETE 			admin/:brandName/:workName
	ChangeWorkFields(c *gin.Context)  // PUT JSON 		admin/:brandName/:workName/change

	ServingWork(c *gin.Context) // GET 	JSON 		admin/:brandName/:workName
}

type AdminService interface {
	AuthAdmin(ctx context.Context, ip string, req *entity.Admin) (string, error)
	LogOut(ctx context.Context, ip string) error

	ListAllBrands(ctx context.Context) ([]entity.Brand, error)
	AddNewBrand(ctx context.Context, brand *entity.Brand) error
	DeleteBrand(ctx context.Context, brandName string) error
	ChangeBrandPassword(ctx context.Context, brand *entity.Brand) error

	AddNewWork(ctx context.Context, brandName, workName, url string, c *gin.Context) (int, error)
	ListAllWorks(ctx context.Context, brandName string) ([]entity.Works, error)
	DeleteWork(ctx context.Context, brandName, workName string) error
	ChangeWorkFields(ctx context.Context, brandName, workName string, work *entity.Works) error
	RenameFolders(brandName, workName string, newInfo *entity.Works) error
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
