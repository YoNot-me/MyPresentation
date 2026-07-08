package authTransport

import (
	"context"
	"errors"
	"net/http"
	"presentator/internal/core/entity"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (t *AuthTransport) Auth(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "./public/auth/auth.html")
}

func (t *AuthTransport) AuthBrand(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := entity.Brand{}
	err := c.BindJSON(&data)
	if err != nil {
		t.log.Error("Failed to parse request", zap.Error(err))
		response(c, entity.Response{
			Status: http.StatusBadRequest,
		})
		return
	}

	token, err := t.s.AuthUser(ctx, data, c.ClientIP())
	if err != nil {

		if errors.Is(err, entity.AlreadySigned) {
			c.Redirect(http.StatusSeeOther, "/works")
			return
		}

		t.log.Error(err.Error())
		response(c, entity.Response{
			Status: http.StatusUnauthorized,
		})
		return
	}

	t.log.Info("auth to work", zap.String("name", data.Name), zap.String("ip", c.ClientIP()))

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"Pres-Access",
		token,
		12*60*60,
		"/",
		"",
		false,
		true,
	)

	c.Redirect(http.StatusSeeOther, "/works/serve")
}

func (t *AuthTransport) Logout(c *gin.Context) {

	err := t.jwt.LogOut(context.Background(), c.ClientIP())
	if err != nil {
		t.log.Error(err.Error())

		response(c, entity.Response{
			Status: http.StatusInternalServerError,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/auth")
}
