package adminTransport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

func (a *AdminTransport) ListBrands(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
	defer cancel()

	brands, err := a.srv.ListAllBrands(ctx)
	if err != nil {
		a.log.Error("ListAllBrands error", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	response(c, entity.Response{
		Status: http.StatusOK,
		Data:   brands,
	})
}

func (a *AdminTransport) AddNewBrand(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	req := entity.Brand{}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.log.Error("Bad request", zap.Error(err))
		response(c, entity.Response{
			Status: http.StatusBadRequest,
			Err:    entity.BadRequest,
		})
		return
	}

	if err := a.srv.AddNewBrand(ctx, &req); err != nil {
		a.log.Error("Err add new brand", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	const query = "/admin/brands"
	c.Redirect(http.StatusSeeOther, query)
}

func (a *AdminTransport) RenameBrand(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	req := entity.Brand{}

	if err := c.BindJSON(&req); err != nil {
		a.log.Error("Bad request from: "+c.ClientIP(), zap.Error(err))
		response(c, entity.Response{
			Status: http.StatusBadRequest,
			Err:    entity.BadRequest,
		})
		return
	}

	brandName := filepath.Clean(c.Param("brandName"))

	err := a.srv.RenameBrand(ctx, brandName, &req)
	if err != nil {
		a.log.Error("Err rename brand "+req.Name, zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/brands")
}

func (a *AdminTransport) DeleteBrand(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	brandName := filepath.Clean(c.Param("brandName"))

	err := a.srv.DeleteBrand(ctx, brandName)
	if err != nil {
		a.log.Error("Err delete brand", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/brands")
}

func (a *AdminTransport) ChangeBrandPassword(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	brandName := filepath.Clean(c.Param("brandName"))

	req := entity.Brand{}
	req.Name = brandName

	if err := c.ShouldBindJSON(&req); err != nil {
		a.log.Error("Bad request", zap.Error(err))
		response(c, entity.Response{
			Status: http.StatusBadRequest,
			Err:    entity.BadRequest,
		})
		return
	}

	err := a.srv.ChangeBrandPassword(ctx, &req)
	if err != nil {
		a.log.Error("Err change brand password"+req.Name, zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	response(c, entity.Response{
		Status: http.StatusOK,
	})
}

func (a *AdminTransport) ListAllBrandWorks(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
	defer cancel()

	brandName := filepath.Clean(c.Param("brandName"))

	works, err := a.srv.ListAllWorks(ctx, brandName)
	if err != nil {
		a.log.Error("ListAllWorks error", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	response(c, entity.Response{
		Status: http.StatusOK,
		Data:   works,
	})
}

func (a *AdminTransport) AddNewWork(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	rawJSON := c.PostForm("data")
	if rawJSON == "" {
		response(c, entity.Response{Err: entity.BadRequest})
		return
	}

	var req entity.Works
	if err := json.Unmarshal([]byte(rawJSON), &req); err != nil {
		response(c, entity.Response{Err: entity.BadRequest})
		return
	}

	brandName := filepath.Clean(c.Param("brandName"))

	count, err := a.srv.AddNewWork(ctx, brandName, &req, c)
	if err != nil {
		a.log.Error("Err add new work", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	answ := fmt.Sprintf("Added to %s %v files", req.WorkName, count)

	response(c, entity.Response{
		Status: http.StatusOK,
		Data:   answ,
	})
}

func (a *AdminTransport) DeleteWork(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	brandName := filepath.Clean(c.Param("brandName"))
	workName := filepath.Clean(c.Param("workName"))

	err := a.srv.DeleteWork(ctx, brandName, workName)
	if err != nil {
		a.log.Error("Failed to delete work", zap.String("name", workName), zap.Any("error", err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	const path = "/admin/%s/works"
	query := fmt.Sprintf(path, brandName)
	c.Redirect(http.StatusSeeOther, query)
}

func (a *AdminTransport) ChangeWorkFields(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	brandName := filepath.Clean(c.Param("brandName"))
	workName := filepath.Clean(c.Param("workName"))

	rawJSON := c.PostForm("data")
	if rawJSON == "" {
		response(c, entity.Response{Err: entity.BadRequest})
		return
	}

	req := entity.Works{}
	err := json.Unmarshal([]byte(rawJSON), &req)
	if err != nil {
		a.log.Error("Bad request", zap.Error(err))
		response(c, entity.Response{
			Status: http.StatusBadRequest,
			Err:    entity.BadRequest,
		})
		return
	}

	err = a.srv.ChangeWorkFields(ctx, brandName, workName, &req, c)
	if err != nil {
		a.log.Error("Cant change work fields", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	const path = "/admin/%s/works"
	query := fmt.Sprintf(path, brandName)
	c.Redirect(http.StatusSeeOther, query)
}

func (a *AdminTransport) ServingWork(c *gin.Context) {

	brandName := filepath.Clean(c.Param("brandName"))
	workName := filepath.Clean(c.Param("workName"))
	relPath := filepath.Clean(c.Param("filepath"))

	const dst = "./works"
	fullPath := filepath.Join(dst, brandName, workName, relPath)

	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		entries, err := os.ReadDir(fullPath)
		if err != nil || len(entries) != 1 || entries[0].IsDir() {
			c.Status(http.StatusNotFound)
			return
		}
		fullPath = filepath.Join(fullPath, entries[0].Name())
	}

	c.File(fullPath)
}
