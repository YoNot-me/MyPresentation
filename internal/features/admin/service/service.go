package adminService

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"presentator/internal/core/entity"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const dst = "./works"

func (s *AdminService) ListAllBrands(ctx context.Context) ([]entity.BrandsResponse, error) {

	brands, err := s.rep.ListAllBrands(ctx)
	if err != nil {
		s.log.Error("failed to list all brands from database", zap.Error(err))
		return nil, err
	}

	return brands, nil
}

func (s *AdminService) AddNewBrand(ctx context.Context, brand *entity.Brand) error {

	if brand.Name == "" {
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}
	if brand.Password == "" {
		s.log.Error("brand password is empty")
		return entity.BadRequest
	}

	brand.Name = filepath.Clean(brand.Name)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(brand.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash brand password", zap.Error(err))
		return err
	}

	err = s.rep.AddNewBrand(ctx, brand.Name, string(hashPass))
	if err != nil {
		s.log.Error("failed to add new brand to database", zap.Error(err))
		return err
	}

	dir := fmt.Sprintf("%s/%s", dst, brand.Name)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		s.log.Error("failed to create brand directory", zap.Error(err))
		return err
	}

	return nil
}

func (s *AdminService) RenameBrand(ctx context.Context, brandName string, newInfo *entity.Brand) error {

	if brandName == "" {
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}
	if newInfo.Name == "" {
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}

	err := os.Rename(
		fmt.Sprintf("%s/%s", dst, brandName),
		fmt.Sprintf("%s/%s", dst, newInfo.Name),
	)
	if err != nil {
		s.log.Error("failed to rename brand directory", zap.Error(err))
		return err
	}

	if renameErr := s.rep.RenameBrand(ctx, brandName, newInfo.Name); renameErr != nil {
		_ = os.Rename(
			fmt.Sprintf("%s/%s", dst, newInfo.Name),
			fmt.Sprintf("%s/%s", dst, brandName),
		)

		s.log.Error("failed to rename brand in database", zap.Error(renameErr))
		return renameErr
	}

	return nil
}

func (s *AdminService) DeleteBrand(ctx context.Context, brandName string) error {

	if brandName == "" {
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}

	err := s.rep.DeleteBrand(ctx, brandName)
	if err != nil {
		s.log.Error("failed to delete brand", zap.Error(err))
		return err
	}

	dir := fmt.Sprintf("%s/%s", dst, brandName)
	err = os.RemoveAll(dir)
	if err != nil {
		s.log.Error("failed to remove brand directory", zap.Error(err))
		return err
	}

	return nil
}

func (s *AdminService) ChangeBrandPassword(ctx context.Context, brand *entity.Brand) error {

	if brand.Name == "" {
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}
	if brand.Password == "" {
		s.log.Error("brand password is empty")
		return entity.BadRequest
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(brand.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.rep.ChangeBrandPassword(ctx, brand.Name, string(hashPass))
	if err != nil {
		s.log.Error("failed to change brand password", zap.Error(err))
		return err
	}

	return nil
}

func (s *AdminService) ListAllWorks(ctx context.Context, brandName string) ([]entity.WorksResponse, error) {

	if brandName == "" {
		s.log.Error("brand name is empty")
		return nil, entity.BadRequest
	}
	work, err := s.rep.ListAllWorks(ctx, brandName)
	if err != nil {
		s.log.Error("failed to list all works", zap.Error(err))
		return nil, err
	}

	if work == nil {
		s.log.Error("works not found")
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
		s.log.Error("failed to create directory", zap.Error(err))
		return 0, err
	}

	previewDir := filepath.Join(dst, brandName, req.WorkName, "preview")
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		s.log.Error("failed to create preview directory", zap.Error(err))
		return 0, err
	}

	form, err := c.MultipartForm()
	if err != nil {
		s.log.Error("failed to get multipart form", zap.Error(err))
		return 0, err
	}

	if previewFiles := form.File["preview"]; len(previewFiles) > 0 && len(previewFiles) < 50 {
		previewFile := previewFiles[0]

		ext := strings.ToLower(filepath.Ext(previewFile.Filename))
		if isAllowedImageExt(ext) {

			previewName := "preview" + ext
			previewPath := filepath.Join(previewDir, previewName)

			if err = c.SaveUploadedFile(previewFile, previewPath); err != nil {
				s.log.Error("failed to save uploaded preview file", zap.Error(err))
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
			s.log.Error("failed to save uploaded file", zap.Error(err))
			return 0, err
		}

		count++
	}

	if count == 0 {
		s.log.Error("no valid files uploaded")
		return 0, entity.BadRequest
	}

	req.Brand = brandName

	if err = s.rep.AddNewWork(ctx, req); err != nil {
		if deleteErr := s.DeleteWork(ctx, req.Brand, req.WorkName); deleteErr != nil {
			s.log.Error("failed to delete work", zap.Error(deleteErr))
		}
		s.log.Error("failed to add new work", zap.Error(err))
		return 0, err
	}

	return count, nil
}

func (s *AdminService) GetWorkImages(brandName, workName string) ([]string, error) {

	if brandName == "" || workName == "" {
		return nil, entity.BadRequest
	}

	dir := filepath.Join(dst, filepath.Clean(brandName), filepath.Clean(workName))

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
		if isAllowedImageExt(strings.ToLower(filepath.Ext(e.Name()))) {
			files = append(files, e.Name())
		}
	}

	sort.Strings(files)
	return files, nil
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
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}
	if workName == "" {
		s.log.Error("work name is empty")
		return entity.BadRequest
	}

	err := s.rep.DeleteWork(ctx, brandName, workName)
	if err != nil {
		s.log.Error("failed to delete work", zap.Error(err))
		return err
	}

	dir := fmt.Sprintf("%s/%s/%s", dst, brandName, workName)
	err = os.RemoveAll(dir)
	if err != nil {
		s.log.Error("failed to remove directory", zap.Error(err))
		return err
	}

	return nil
}

func (s *AdminService) ChangeWorkFields(
	ctx context.Context,
	brandName, workName string,
	c *gin.Context,
) error {

	if workName == "" {
		s.log.Error("work name is empty")
		return entity.BadRequest
	}
	if brandName == "" {
		s.log.Error("brand name is empty")
		return entity.BadRequest
	}

	form, err := c.MultipartForm()
	if err != nil {
		s.log.Error("failed to get multipart form", zap.Error(err))
		return err
	}

	if previewFiles := form.File["preview"]; len(previewFiles) > 0 {
		previewFile := previewFiles[0]

		previewDir := filepath.Join(dst, brandName, workName, "preview")
		if err = os.MkdirAll(previewDir, 0755); err != nil {
			s.log.Error("failed to create preview directory", zap.Error(err))
			return err
		}

		ext := strings.ToLower(filepath.Ext(previewFile.Filename))
		if isAllowedImageExt(ext) {

			previewName := "preview" + ext
			previewPath := filepath.Join(previewDir, previewName)

			_ = c.SaveUploadedFile(previewFile, previewPath)
		}
	}

	rawJSON := c.PostForm("data")
	newInfo := entity.Works{}

	if rawJSON != "" {
		s.log.Info("rawJSON",
			zap.String("value", "rawJSON"),
			zap.Int("len", len(rawJSON)),
		)

		if marshalErr := json.Unmarshal([]byte(rawJSON), &newInfo); marshalErr != nil {
			s.log.Error("failed to unmarshal work data", zap.Error(marshalErr))
			return marshalErr
		}
		if err = s.rep.ChangeWorkFields(ctx, brandName, workName, &newInfo); err != nil {
			s.log.Error("failed to change work fields", zap.Error(err))
			return err
		}
	}

	currentWorkName := workName
	if newInfo.WorkName != "" {
		currentWorkName = newInfo.WorkName
	}

	workInfo, err := s.rep.GetWork(ctx, brandName, currentWorkName)
	if err != nil {
		s.log.Error("failed to get work", zap.Error(err))
		return err
	}

	if workName != newInfo.WorkName {
		if renameErr := s.RenameFolders(brandName, workName, newInfo.WorkName); renameErr != nil {
			rollBack := entity.Works{
				Brand:    brandName,
				WorkName: workName,
			}

			rollbackErr := s.rep.ChangeWorkFields(ctx, workInfo.Brand, workInfo.WorkName, &rollBack)
			if rollbackErr != nil {
				s.log.Error("failed to rollback work fields", zap.Error(rollbackErr))
				return rollbackErr
			}

			return renameErr
		}
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
			s.log.Error("failed to rename work folder", zap.Error(err))
			return err
		}
	}

	return nil
}
