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

func (a *AdminTransport) AuthAdmin(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	req := entity.Admin{}
	if err := c.BindJSON(&req); err != nil {
		response(c, entity.Response{
			Status: http.StatusBadRequest,
		})
		return
	}

	token, err := a.srv.AuthAdmin(ctx, c.ClientIP(), &req)
	if err != nil {
		a.log.Error("auth admin", zap.Error(err))
		response(c, entity.Response{
			Err: err,
		})
		return
	}

	c.SetCookie(
		"Pres-Access",
		token,
		12*60*60,
		"/",
		"",
		false,
		true,
	)

	path := fmt.Sprintf("/%s", filepath.Join("admin", "brands"))
	c.Redirect(http.StatusFound, path)
}

func (a *AdminTransport) LogOut(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	err := a.srv.LogOut(ctx, c.ClientIP())
	if err != nil {
		a.log.Error("Cant log out: "+c.ClientIP(), zap.Error(err))

		return
	}

	c.Redirect(http.StatusSeeOther, "/admin")
}
