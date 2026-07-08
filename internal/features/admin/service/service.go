package adminService

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const dst = "./works"

func (s *AdminService) ListAllBrands(ctx context.Context) ([]entity.BrandsResponse, error) {

	brands, err := s.rep.ListAllBrands(ctx)
	if err != nil {
		return nil, err
	}

	return brands, nil
}

func (s *AdminService) AddNewBrand(ctx context.Context, brand *entity.Brand) error {

	if brand.Name == "" {
		return entity.BadRequest
	}
	if brand.Password == "" {
		return entity.BadRequest
	}

	brand.Name = filepath.Clean(brand.Name)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(brand.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.rep.AddNewBrand(ctx, brand.Name, string(hashPass))
	if err != nil {
		return err
	}

	dir := fmt.Sprintf("%s/%s", dst, brand.Name)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) RenameBrand(ctx context.Context, brandName string, newInfo *entity.Brand) error {

	if brandName == "" {
		return entity.BadRequest
	}
	if newInfo.Name == "" {
		return entity.BadRequest
	}

	err := os.Rename(
		fmt.Sprintf("%s/%s", dst, brandName),
		fmt.Sprintf("%s/%s", dst, newInfo.Name),
	)
	if err != nil {
		return err
	}

	err = s.rep.RenameBrand(ctx, brandName, newInfo.Name)
	if err != nil {

		_ = os.Rename(
			fmt.Sprintf("%s/%s", dst, newInfo.Name),
			fmt.Sprintf("%s/%s", dst, brandName),
		)

		return err
	}

	return nil
}

func (s *AdminService) DeleteBrand(ctx context.Context, brandName string) error {

	if brandName == "" {
		return entity.BadRequest
	}

	err := s.rep.DeleteBrand(ctx, brandName)
	if err != nil {
		return err
	}

	dir := fmt.Sprintf("%s/%s", dst, brandName)
	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) ChangeBrandPassword(ctx context.Context, brand *entity.Brand) error {

	if brand.Name == "" {
		return entity.BadRequest
	}
	if brand.Password == "" {
		return entity.BadRequest
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(brand.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.rep.ChangeBrandPassword(ctx, brand.Name, string(hashPass))
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) ListAllWorks(ctx context.Context, brandName string) ([]entity.WorksResponse, error) {

	if brandName == "" {
		return nil, entity.BadRequest
	}
	work, err := s.rep.ListAllWorks(ctx, brandName)
	if err != nil {
		return nil, err
	}

	if work == nil {
		return nil, entity.NotFound
	}

	return work, nil
}

func (s *AdminService) AddNewWork(ctx context.Context, brandName string, req *entity.Works, c *gin.Context) (int, error) {

	if brandName == "" || req.WorkName == "" {
		return 0, entity.BadRequest
	}

	dir := filepath.Join(dst, brandName, req.WorkName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, err
	}

	previewDir := filepath.Join(dst, brandName, req.WorkName, "preview")
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return 0, err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return 0, err
	}

	if previewFiles := form.File["preview"]; len(previewFiles) > 0 && len(previewFiles) < 50 {
		previewFile := previewFiles[0]

		ext := strings.ToLower(filepath.Ext(previewFile.Filename))
		if isAllowedImageExt(ext) {

			previewName := "preview" + ext
			previewPath := filepath.Join(previewDir, previewName)

			if err = c.SaveUploadedFile(previewFile, previewPath); err != nil {
				return 0, err
			}
		}
	}

	files := form.File["files"]
	var count int

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Filename))
		if !isAllowedImageExt(ext) {
			continue
		}

		safeFilename := filepath.Base(f.Filename)
		filePath := filepath.Join(dir, safeFilename)

		if err = c.SaveUploadedFile(f, filePath); err != nil {
			return 0, err
		}

		count++
	}

	if count == 0 {
		return 0, entity.BadRequest
	}
	req.Brand = brandName

	if err = s.rep.AddNewWork(ctx, req); err != nil {
		_ = s.DeleteWork(ctx, req.Brand, req.WorkName)
		return 0, err
	}

	return count, nil
}

func isAllowedImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

func (s *AdminService) DeleteWork(ctx context.Context, brandName, workName string) error {

	if brandName == "" {
		return entity.BadRequest
	}
	if workName == "" {
		return entity.BadRequest
	}

	err := s.rep.DeleteWork(ctx, brandName, workName)
	if err != nil {
		return err
	}

	dir := fmt.Sprintf("%s/%s/%s", dst, brandName, workName)
	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminService) ChangeWorkFields(
	ctx context.Context,
	brandName, workName string,
	newInfo *entity.Works,
	c *gin.Context,
) error {

	if workName == "" {
		return entity.BadRequest
	}
	if brandName == "" {
		return entity.BadRequest
	}
	if newInfo == nil {
		return entity.BadRequest
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	if previewFiles := form.File["preview"]; len(previewFiles) > 0 {
		previewFile := previewFiles[0]

		previewDir := filepath.Join(dst, brandName, workName, "preview")
		if err = os.MkdirAll(previewDir, 0755); err != nil {
			return err
		}

		ext := strings.ToLower(filepath.Ext(previewFile.Filename))
		if isAllowedImageExt(ext) {

			previewName := "preview" + ext
			previewPath := filepath.Join(previewDir, previewName)

			_ = c.SaveUploadedFile(previewFile, previewPath)
		}
	}

	err = s.rep.ChangeWorkFields(ctx, brandName, workName, newInfo)
	if err != nil {
		return err
	}

	currentWorkName := workName
	if newInfo.WorkName != "" {
		currentWorkName = newInfo.WorkName
	}

	workInfo, err := s.rep.GetWork(ctx, brandName, currentWorkName)
	if err != nil {
		return err
	}

	if renameErr := s.RenameFolders(brandName, workName, newInfo.WorkName); renameErr != nil {

		rollBack := entity.Works{
			Brand:    brandName,
			WorkName: workName,
		}

		rollbackErr := s.rep.ChangeWorkFields(ctx, workInfo.Brand, workInfo.WorkName, &rollBack)
		if rollbackErr != nil {
			return rollbackErr
		}

		return renameErr
	}

	return nil
}

func (s *AdminService) RenameFolders(brandName, workName string, newWorkName string) error {

	if newWorkName != "" {
		err := os.Rename(
			fmt.Sprintf("%s/%s/%s", dst, brandName, workName),
			fmt.Sprintf("%s/%s/%s", dst, brandName, newWorkName),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
