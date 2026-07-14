package adminService

import (
	"context"
	"errors"
	"os"
	"presentator/internal/core/entity"
	JWT "presentator/internal/features/auth/token"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *AdminService) AuthAdmin(ctx context.Context, ip string, req *entity.Admin) (string, error) {

	count, err := s.rep.BruteCount(ctx, ip)
	if err != nil {
		return "", err
	}
	if count >= 5 {
		return "", entity.TooManyAttempts
	}

	login := os.Getenv("PRES_ADMIN_LOGIN")
	password := os.Getenv("PRES_ADMIN_PASSWORD")

	if login == "" || password == "" {
		return "", entity.EnvNotLoaded
	}

	if !strings.EqualFold(req.Login, login) {
		incErr := s.rep.IncCount(ctx, ip)
		if incErr != nil {
			s.log.Error("failed to equal login: ",
				zap.Error(errors.New("login not equal: "+req.Login+"/"+incErr.Error())))

			return "", entity.InvalidLogin
		}

		return "", entity.InvalidLogin
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); compareErr != nil {
		incErr := s.rep.IncCount(ctx, ip)
		if incErr != nil {
			s.log.Error("failed to compare pass and inc count: ",
				zap.Error(errors.New(compareErr.Error()+"/"+incErr.Error())))

			return "", entity.InvalidPass
		}

		return "", entity.InvalidPass
	}

	token, err := s.jwt.CreateToken(ctx, "", ip, "admin")
	if err != nil {
		return "", err
	}

	return token, err
}

func (s *AdminService) LogOut(ctx context.Context, cookie string) error {

	token, err := s.jwt.ParseToken(cookie)
	if err != nil {
		s.log.Error("failed to parse token", zap.Error(err))
		return err
	}

	claims := token.Claims.(*JWT.JWT)

	return s.jwt.LogOut(ctx, claims.ID)
}
