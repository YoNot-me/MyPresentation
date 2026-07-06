package fileservingService

import (
	"context"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"

	"github.com/gin-gonic/gin"
)

func (s *ServingService) GetName(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("Pres-Access")
	if err != nil {
		return "", err
	}

	token, ok := s.jwt.ValidateToken(cookie)
	if !ok {
		return "", entity.InvalidToken
	}

	claims, ok := token.Claims.(*JWT.JWT)
	if !ok {
		return "", entity.InvalidToken
	}

	if claims.BrandName == "" {
		return "", entity.InvalidToken
	}

	return claims.BrandName, nil
}

func (s *ServingService) GetBrandWorks(ctx context.Context, brandName string) ([]entity.Works, error) {

	response, err := s.db.GetAllWorks(ctx, brandName)
	if err != nil {
		return nil, err
	}

	return response, nil
}
