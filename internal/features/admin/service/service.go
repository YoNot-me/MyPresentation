package adminService

import (
	"context"
	"encoding/json"
	"errors"
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

	if validateErr := s.validateNotNil([]string{brand.Name, brand.Password}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return validateErr
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

	if validateErr := s.validateNotNil([]string{brandName, newInfo.Name}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return validateErr
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

	if validateErr := s.validateNotNil([]string{brandName}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return validateErr
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

	if validateErr := s.validateNotNil([]string{brand.Name, brand.Password}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return validateErr
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(brand.Password), bcrypt.DefaultCost)
	if err != nil {
		return entity.InvalidPass
	}

	err = s.rep.ChangeBrandPassword(ctx, brand.Name, string(hashPass))
	if err != nil {
		s.log.Error("failed to change brand password", zap.Error(err))
		return err
	}

	return nil
}

func (s *AdminService) ListAllWorks(ctx context.Context, brandName string) ([]entity.WorksResponse, error) {

	if validateErr := s.validateNotNil([]string{brandName}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return nil, validateErr
	}

	work, err := s.rep.ListAllWorks(ctx, brandName)
	if err != nil {
		s.log.Error("failed to list all works", zap.Error(err))
		return nil, err
	}

	return work, nil
}

func (s *AdminService) AddNewWork(ctx context.Context, req *entity.Works, c *gin.Context) (int, error) {

	if validateErr := s.validateNotNil([]string{req.Brand, req.WorkName}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return 0, validateErr
	}

	if req.Status != "" {
		if statusErr := s.validateStatus(req.Status); statusErr != nil {
			s.log.Error("not valid status", zap.Error(statusErr))
			return 0, statusErr
		}
	}

	ok, err := s.rep.IsWorkExist(ctx, req.Brand, req.WorkName)
	if err != nil {
		s.log.Error("failed to check if work exists", zap.Error(err))
		return 0, entity.InternalError
	}
	if ok {
		s.log.Error("work already exists", zap.Error(errors.New(
			req.Brand+" "+req.WorkName+" already exists")))
		return 0, entity.BadRequest
	}

	dir := filepath.Join(dst, req.Brand, req.WorkName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.log.Error("failed to create directory", zap.Error(err))
		return 0, err
	}
	previewDir := filepath.Join(dst, req.Brand, req.WorkName, "preview")
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

	if err = s.rep.AddNewWork(ctx, req); err != nil {
		if errors.Is(err, entity.AlreadyExist) {
			return 0, entity.AlreadyExist
		}

		if deleteErr := s.DeleteWork(ctx, req.Brand, req.WorkName); deleteErr != nil {
			s.log.Error("failed to delete work", zap.Error(deleteErr))
			return 0, entity.InternalError
		}

		s.log.Error("failed to add new work", zap.Error(err))
		return 0, err
	}

	return count, nil
}

func (s *AdminService) GetWorkImages(brandName, workName string) ([]string, error) {

	if validateErr := s.validateNotNil([]string{brandName, workName}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return nil, validateErr
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

	if validateErr := s.validateNotNil([]string{brandName, workName}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return validateErr
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

func (s *AdminService) ChangeWorkFields(ctx context.Context, brandName, workName string,
	c *gin.Context) error {

	if validateErr := s.validateNotNil([]string{brandName, workName}); validateErr != nil {
		s.log.Error("not valid request", zap.Error(validateErr))
		return validateErr
	}

	form, err := c.MultipartForm()
	if err != nil {
		s.log.Error("failed to get multipart form", zap.Error(err))
		return err
	}

	rawJSON := c.PostForm("data")
	newInfo := entity.Works{}
	if rawJSON != "" {
		if err = json.Unmarshal([]byte(rawJSON), &newInfo); err != nil {
			s.log.Error("failed to unmarshal work data", zap.Error(err))
			return err
		}
	}

	if newInfo.Status != "" {
		if statusErr := s.validateStatus(newInfo.Status); statusErr != nil {
			s.log.Error("not valid status", zap.Error(err))
			return statusErr
		}
	}

	renamed := newInfo.WorkName != "" && newInfo.WorkName != workName
	if renamed {

		s.log.Info("renaming work folder :349")

		if err = s.renameFolders(brandName, workName, newInfo.WorkName); err != nil {
			s.log.Error("failed to rename work folder", zap.Error(err))
			return err
		}
	}

	currentWorkName := workName
	if renamed {
		currentWorkName = newInfo.WorkName
	}

	if rawJSON != "" {

		s.log.Info("changing work fields :362")

		if err = s.rep.ChangeWorkFields(ctx, brandName, workName, &newInfo); err != nil {
			s.log.Error("failed to change work fields", zap.Error(err))

			if renamed {
				if rbErr := s.renameFolders(brandName, newInfo.WorkName, workName); rbErr != nil {
					s.log.Error("failed to rollback folder rename after db error",
						zap.Error(rbErr))
					return fmt.Errorf("change failed: %w; rollback also failed: %v", err, rbErr)
				}
			}

			s.log.Error("failed to change work fields :377")
			return err
		}
	}

	if previewFiles := form.File["preview"]; len(previewFiles) > 0 {
		previewFile := previewFiles[0]

		s.log.Info("changing preview file :385")

		ext := strings.ToLower(filepath.Ext(previewFile.Filename))
		if isAllowedImageExt(ext) {
			s.log.Info("changing preview file :389")
			previewDir := filepath.Join(dst, brandName, currentWorkName, "preview")

			if err = os.MkdirAll(previewDir, 0755); err != nil {
				s.log.Error("failed to create preview directory", zap.Error(err))
				return err
			}

			entr, readErr := os.ReadDir(previewDir)
			if readErr == nil {
				s.log.Info("removing preview files :399")

				for _, v := range entr {
					if err = os.Remove(filepath.Join(previewDir, v.Name())); err != nil {
						s.log.Error("failed to remove preview file", zap.Error(err))
						return err
					}
				}
			}

			previewPath := filepath.Join(previewDir, "preview"+ext)
			if saveErr := c.SaveUploadedFile(previewFile, previewPath); saveErr != nil {
				s.log.Error("failed to save preview file", zap.Error(saveErr))
				return saveErr
			}

			return nil
		}
	}

	return nil
}
