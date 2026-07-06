package fileservingTransport

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"presentator/internal/core/entity"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (t *FileServingTransport) ServeHTML(c *gin.Context) {

	http.ServeFile(c.Writer, c.Request, "./public/presentation/index.html")
}

func (t *FileServingTransport) ListWorkFiles(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	brandName, err := t.srv.GetName(c)
	if err != nil {
		t.log.Error("err get brand name", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	presentations, err := t.srv.GetBrandWorks(ctx, brandName)
	if err != nil {
		t.log.Error("err get works", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	response(c, entity.Response{
		Status: http.StatusOK,
		Data:   presentations,
	})
}

func (t *FileServingTransport) GetWork(c *gin.Context) {

	workName := filepath.Clean(c.Param("name"))

	brandName, err := t.srv.GetName(c)
	if err != nil {
		t.log.Error("err get brand name", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	dir := fmt.Sprintf("./works/%s/%s", brandName, workName)

	fs := http.StripPrefix("/presentation/"+workName, http.FileServer(http.Dir(dir)))
	fs.ServeHTTP(c.Writer, c.Request)
}
