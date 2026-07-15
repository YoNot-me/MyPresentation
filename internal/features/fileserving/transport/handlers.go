package fileservingTransport

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (t *FileServingTransport) ServeHTML(c *gin.Context) {
	c.File("./public/presentation/index.html")
}

func (t *FileServingTransport) ListWorkFiles(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
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
	relPath := filepath.Clean(c.Param("filepath"))

	brandName, err := t.srv.GetName(c)
	if err != nil {
		t.log.Error("err get brand name", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	root, _ := filepath.Abs("./works")
	full := filepath.Join(root, brandName, workName, relPath)

	full, _ = filepath.Abs(full)
	if !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		t.log.Error("Err path traversal", zap.Error(entity.ErrPathTraversal))
		response(c, entity.Response{
			Err: entity.ErrPathTraversal,
		})
		return
	}

	fullPath := filepath.Join("./works", brandName, workName, relPath)

	if info, statErr := os.Stat(fullPath); statErr == nil && info.IsDir() {
		entries, readErr := os.ReadDir(fullPath)
		if readErr != nil || len(entries) != 1 || entries[0].IsDir() {
			c.Status(http.StatusNotFound)
			return
		}
		fullPath = filepath.Join(fullPath, entries[0].Name())
	}

	c.File(fullPath)
}

func (t *FileServingTransport) ListWorkImages(c *gin.Context) {

	workName := c.Param("name")

	brandName, err := t.srv.GetName(c)
	if err != nil {
		t.log.Error("err get brand name", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	files, err := t.srv.GetWorkImages(brandName, workName)
	if err != nil {
		t.log.Error("err list work images", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	response(c, entity.Response{
		Status: http.StatusOK,
		Data:   files,
	})
}
