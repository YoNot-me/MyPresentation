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

func (s *AdminService) ListAllBrands(ctx context.Context) ([]entity.Brand, error) {

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

func (s *AdminService) ListAllWorks(ctx context.Context, brandName string) ([]entity.Works, error) {

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

func (s *AdminService) AddNewWork(ctx context.Context, brandName, workName, url string, c *gin.Context) (int, error) {

	if brandName == "" {
		return 0, entity.BadRequest
	}
	if workName == "" {
		return 0, entity.BadRequest
	}

	err := s.rep.AddNewWork(ctx, brandName, workName, url)
	if err != nil {
		return 0, err
	}

	dir := fmt.Sprintf("%s/%s/%s", dst, brandName, workName)
	err = os.MkdirAll(dir, 0755)
	if err != nil {

		err = s.DeleteWork(ctx, brandName, workName)
		return 0, err
	}

	form, err := c.MultipartForm()
	if err != nil {

		err = s.DeleteWork(ctx, brandName, workName)
		return 0, err
	}

	files := form.File["files"]
	if len(files) == 0 {

		err = s.DeleteWork(ctx, brandName, workName)
		return 0, entity.BadRequest
	}

	var count int

	for _, f := range files {
		safeFilename := filepath.Base(f.Filename)

		ext := strings.ToLower(filepath.Ext(safeFilename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		if !allowed[ext] {
			continue
		}

		count++

		filePath := filepath.Join(dir, safeFilename)

		err = c.SaveUploadedFile(f, filePath)
		if err != nil {

			err = s.DeleteWork(ctx, brandName, workName)
			return 0, err
		}
	}

	return count, nil
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

func (s *AdminService) ChangeWorkFields(ctx context.Context, brandName, workName string, newInfo *entity.Works) error {

	if workName == "" {
		return entity.BadRequest
	}
	if brandName == "" {
		return entity.BadRequest
	}
	if newInfo == nil {
		return entity.BadRequest
	}

	err := s.rep.ChangeWorkFields(ctx, brandName, workName, newInfo)
	if err != nil {
		return err
	}

	workInfo, err := s.rep.GetWork(ctx, newInfo.WorkName)
	if err != nil {
		return err
	}

	if renameErr := s.RenameFolders(brandName, workName, newInfo); renameErr != nil {

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

func (s *AdminService) RenameFolders(brandName, workName string, newInfo *entity.Works) error {

	if newInfo.WorkName != "" {
		err := os.Rename(
			fmt.Sprintf("%s/%s/%s", dst, brandName, workName),
			fmt.Sprintf("%s/%s/%s", dst, brandName, newInfo.WorkName),
		)
		if err != nil {
			return err
		}
	}

	if newInfo.Brand != "" {
		err := os.Rename(
			fmt.Sprintf("%s/%s", dst, brandName),
			fmt.Sprintf("./works/%s", newInfo.Brand),
		)
		if err != nil {
			return err
		}
	}

	return nil
}
