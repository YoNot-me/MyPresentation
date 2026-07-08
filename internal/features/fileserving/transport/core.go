package fileservingTransport

import (
	"context"
	"presentator/internal/core/entity"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ServingTrasnport interface {
	ServeHTML(c *gin.Context)      // GET 	/works/serve
	ListWorkFiles(c *gin.Context)  // GET 	/works
	ListWorkImages(c *gin.Context) // GET 	/works/files/:name
	GetWork(c *gin.Context)        // GET 	/presentation/:name/*filepath
}

type ServingSrv interface {
	GetName(c *gin.Context) (string, error)
	GetBrandWorks(ctx context.Context, brandName string) ([]entity.Works, error)
	GetWorkImages(brandName, workName string) ([]string, error)
}

type FileServingTransport struct {
	srv ServingSrv
	log *zap.Logger
}

func NewServingTransport(srv ServingSrv, log *zap.Logger) *FileServingTransport {
	return &FileServingTransport{
		srv: srv,
		log: log,
	}
}
