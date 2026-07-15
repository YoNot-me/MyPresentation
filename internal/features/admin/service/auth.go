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
		s.log.Error("failed to get brute count: ", zap.Error(err))
		return "", err
	}
	if count >= 5 {
		s.log.Error("too many attempts")
		return "", entity.TooManyAttempts
	}

	login := os.Getenv("PRES_ADMIN_LOGIN")
	password := os.Getenv("PRES_ADMIN_PASSWORD")

	if login == "" || password == "" {
		s.log.Error("failed to load env")
		return "", entity.EnvNotLoaded
	}

	if !strings.EqualFold(req.Login, login) {
		s.log.Error("failed to equal login: ", zap.String("login", req.Login))

		incErr := s.rep.IncCount(ctx, ip)
		if incErr != nil {
			s.log.Error("failed to equal login: ",
				zap.Error(errors.New("login not equal: "+req.Login+"/"+incErr.Error())))

			return "", entity.InvalidLogin
		}

		return "", entity.InvalidLogin
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); compareErr != nil {
		s.log.Error("failed to compare pass: ", zap.Error(compareErr))

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
		s.log.Error("failed to create token: ", zap.Error(err))
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
