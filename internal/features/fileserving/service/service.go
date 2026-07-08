package fileservingService

import (
	"context"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"presentator/internal/features/auth/token"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const worksDir = "./works"

func (s *ServingService) GetName(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("Pres-Access")
	if err != nil {
		s.log.Error("failed to get cookie", zap.Error(err))
		return "", err
	}

	token, ok := s.jwt.ValidateToken(cookie)
	if !ok {
		s.log.Error("failed to validate token")
		return "", entity.InvalidToken
	}

	claims, ok := token.Claims.(*JWT.JWT)
	if !ok {
		s.log.Error("failed to parse token claims")
		return "", entity.InvalidToken
	}

	if claims.BrandName == "" {
		s.log.Error("brand name is empty")
		return "", entity.InvalidToken
	}

	return claims.BrandName, nil
}

func (s *ServingService) GetBrandWorks(ctx context.Context, brandName string) ([]entity.Works, error) {

	response, err := s.db.GetAllWorks(ctx, brandName)
	if err != nil {
		s.log.Error("failed to get all works", zap.Error(err))
		return nil, err
	}

	return response, nil
}

// GetWorkImages returns the slide image file names inside a work folder,
// sorted by name. Sub-folders (such as "preview") and non-images are skipped.
func (s *ServingService) GetWorkImages(brandName, workName string) ([]string, error) {

	if brandName == "" || workName == "" {
		return nil, entity.BadRequest
	}

	dir := filepath.Join(worksDir, filepath.Clean(brandName), filepath.Clean(workName))

	entries, err := os.ReadDir(dir)
	if err != nil {
		s.log.Error("failed to read work directory", zap.Error(err))
		return nil, entity.NotFound
	}

	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".jpg", ".jpeg", ".png", ".gif":
			files = append(files, e.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}
