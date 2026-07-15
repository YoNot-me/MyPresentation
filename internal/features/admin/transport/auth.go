package adminTransport

import (
	"context"
	"errors"
	"net/http"
	"presentator/internal/core/entity"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (a *AdminTransport) MainAdmin(c *gin.Context) {
	c.Redirect(http.StatusFound, "/auth/admin.html")
}

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

		if errors.Is(err, entity.TooManyAttempts) {
			a.log.Error("auth admin", zap.Error(err))
			response(c, entity.Response{
				Err: err,
			})
			return
		}

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
		true,
		true,
	)

	a.log.Info("Admin logged in", zap.String("IP: ", c.ClientIP()))
	c.Redirect(http.StatusFound, "/admin/brands")
}

func (a *AdminTransport) LogOut(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	cookie, err := c.Cookie("Pres-Access")
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/auth/admin.html")
		c.Abort()
		return
	}

	logOutErr := a.srv.LogOut(ctx, cookie)
	if logOutErr != nil {
		a.log.Error("Cant log out: "+c.ClientIP(), zap.Error(err))

		return
	}

	a.log.Info("Admin logged out", zap.String("IP: ", c.ClientIP()))
	c.Redirect(http.StatusSeeOther, "/auth/admin.html")
}
