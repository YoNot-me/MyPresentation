package adminTransport

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

func (a *AdminTransport) ListBrands(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

func (a *AdminTransport) DeleteBrand(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

// нужно чтобы начал отличать расширения, можно тольео png, jpg, jpeg, gif
func (a *AdminTransport) AddNewWork(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	brandName := filepath.Clean(c.Param("brandName"))
	workName := filepath.Clean(c.Param("workName"))
	url := c.PostForm("url")

	count, err := a.srv.AddNewWork(ctx, brandName, workName, url, c)
	if err != nil {
		a.log.Error("Err add new work", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	answ := fmt.Sprintf("Added to %s %v files", workName, count)

	response(c, entity.Response{
		Status: http.StatusOK,
		Data:   answ,
	})
}

func (a *AdminTransport) DeleteWork(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := entity.Works{}
	brandName := filepath.Clean(c.Param("brandName"))
	workName := filepath.Clean(c.Param("workName"))

	err := c.ShouldBindJSON(&req)
	if err != nil {
		a.log.Error("Bad request", zap.Error(err))
		response(c, entity.Response{
			Status: http.StatusBadRequest,
			Err:    entity.BadRequest,
		})
		return
	}

	err = a.srv.ChangeWorkFields(ctx, brandName, workName, &req)
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

	const dst = "./works"
	dir := fmt.Sprintf("%s/%s/%s", dst, brandName, workName)
	prefix := fmt.Sprintf("/admin/%s/serve/%s", brandName, workName)

	fs := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	fs.ServeHTTP(c.Writer, c.Request)
}
