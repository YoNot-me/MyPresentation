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
		return "", err
	}

	hashPass, err := s.rep.GetPass(ctx, data.Name)
	if err != nil {
		err = s.rep.IncCount(ctx, ip)
		if err != nil {
			return "", err
		}

		s.log.Error("failed to get password", zap.Error(err))
		return "", err
	}
	if hashPass == "" {
		err = s.rep.IncCount(ctx, ip)
		if err != nil {
			return "", err
		}

		s.log.Error("password not found")
		return "", entity.InternalError
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(data.Password))
	if err != nil {
		err = s.rep.IncCount(ctx, ip)
		if err != nil {
			return "", err
		}

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

func (s *AuthService) CheckData(ctx context.Context, data entity.Brand, ip string) error {
	if data.Name == "" {
		s.log.Error("CheckData: name is empty")
		return entity.BadRequest
	}
	if data.Password == "" {
		s.log.Error("CheckData: password is empty")
		return entity.BadRequest
	}
	if ip == "" {
		s.log.Error("CheckData: ip is empty")
		return entity.BadRequest
	}

	count, err := s.rep.BruteCount(ctx, ip)
	if err != nil {
		s.log.Error("CheckData: failed to get count", zap.Error(err))
		return err
	}
	if count > 5 {
		s.log.Error("CheckData: too many attempts: " + ip)
		return entity.TooManyAttempts
	}

	return nil
}
