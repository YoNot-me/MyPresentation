package adminService

import (
	"context"
	"os"
	"presentator/internal/core/entity"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (s *AdminService) AuthAdmin(ctx context.Context, ip string, req *entity.Admin) (string, error) {

	count, err := s.rep.BruteCount(ctx, ip)
	if err != nil {
		return "", err
	}
	if count > 5 {
		return "", entity.TooManyAttempts
	}

	login := os.Getenv("PRES_ADMIN_LOGIN")
	password := os.Getenv("PRES_ADMIN_PASSWORD")

	if login == "" || password == "" {
		return "", entity.EnvNotLoaded
	}

	if !strings.EqualFold(req.Login, login) {
		return "", entity.InvalidLogin
	}
	if compareErr := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); compareErr != nil {
		incErr := s.rep.IncCount(ctx, ip)
		if incErr != nil {
			return "", incErr
		}

		return "", compareErr
	}

	token, err := s.jwt.CreateToken(ctx, "", ip, "admin")
	if err != nil {
		return "", err
	}

	return token, err
}

func (s *AdminService) LogOut(ctx context.Context, ip string) error {

	if ip == "" {
		return entity.BadRequest
	}

	return s.jwt.LogOut(ctx, ip)
}
