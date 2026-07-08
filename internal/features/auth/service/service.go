package authService

import (
	"context"
	"presentator/internal/core/entity"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) AuthUser(ctx context.Context, data entity.Brand, ip string) (string, error) {

	err := s.CheckData(ctx, data, ip)
	if err != nil {
		s.log.Error("failed to check data", zap.Error(err))
		return "", err
	}

	hashPass, err := s.rep.GetPass(ctx, data.Name)
	if err != nil {
		s.log.Error("failed to get password", zap.Error(err))
		return "", err
	}
	if hashPass == "" {
		s.log.Error("password not found")
		return "", entity.InternalError
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(data.Password))
	if err != nil {
		s.log.Error("failed to compare password", zap.Error(err))
		return "", entity.InvalidPass
	}

	token, err := s.jwt.CreateToken(ctx, data.Name, ip, "guest")
	if err != nil {
		s.log.Error("failed to create token", zap.Error(err))
		return "", err
	}

	return token, nil

}

func (a *AuthService) CheckData(ctx context.Context, data entity.Brand, ip string) error {
	if data.Name == "" {
		a.log.Error("name is empty")
		return entity.BadRequest
	}
	if data.Password == "" {
		a.log.Error("password is empty")
		return entity.BadRequest
	}
	if ip == "" {
		a.log.Error("ip is empty")
		return entity.BadRequest
	}

	if ok := a.jwt.IsExist(ctx, ip); !ok {
		a.log.Error("token already exists")
		return entity.AlreadySigned
	}

	return nil
}
