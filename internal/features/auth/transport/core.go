package authTransport

import (
	"context"
	"presentator/internal/core/entity"
	authService "presentator/internal/features/auth/service"
	"presentator/internal/features/auth/token"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthTrnsprt interface {
	AuthBrand(c *gin.Context) // 		POST 	/auth/check
	Auth(c *gin.Context)      // 		GET 	/auth (front)
	Logout(c *gin.Context)    // 		POST 	/logout
}

type AuthSrv interface {
	AuthUser(ctx context.Context, data entity.Brand, ip string) (string, error)
	CheckData(ctx context.Context, data entity.Brand, ip string) error
}

type AuthTransport struct {
	jwt *JWT.ServingJWT
	log *zap.Logger
	s   AuthSrv
}

func NewAuthTransport(jwt *JWT.ServingJWT, log *zap.Logger, s *authService.AuthService) *AuthTransport {
	return &AuthTransport{
		jwt: jwt,
		log: log,
		s:   s,
	}
}
