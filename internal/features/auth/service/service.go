package authService

import (
	"context"
	"presentator/internal/core/entity"

	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) AuthUser(ctx context.Context, data entity.Brand, ip string) (string, error) {

	err := s.CheckData(ctx, data, ip)
	if err != nil {
		return "", err
	}

	hashPass, err := s.rep.GetPass(ctx, data.Name)
	if err != nil {
		return "", err
	}
	if hashPass == "" {
		return "", entity.InternalError
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(data.Password))
	if err != nil {
		return "", entity.InvalidPass
	}

	token, err := s.jwt.CreateToken(ctx, data.Name, ip, "guest")
	if err != nil {
		return "", err
	}

	return token, nil

}

func (a *AuthService) CheckData(ctx context.Context, data entity.Brand, ip string) error {
	if data.Name == "" {
		return entity.BadRequest
	}
	if data.Password == "" {
		return entity.BadRequest
	}
	if ip == "" {
		return entity.BadRequest
	}

	if ok := a.jwt.IsExist(ctx, ip); !ok {
		return entity.AlreadySigned
	}

	return nil
}
