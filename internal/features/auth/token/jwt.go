package JWT

import (
	"context"
	"errors"
	"fmt"
	"presentator/internal/core/entity"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func (j *ServingJWT) ValidateToken(token string) (*jwt.Token, bool) {

	parsed, err := j.ParseToken(token)
	if err != nil {
		j.log.Error("jwt.go :20 >> failed to parse token", zap.Error(err))
		return nil, false
	}
	if !parsed.Valid {
		return nil, false
	}

	claims := parsed.Claims.(*JWT)
	if ok := j.IsExist(context.Background(), claims.RegisteredClaims.ID); !ok {
		return nil, false
	}

	return parsed, true
}

func (j *ServingJWT) ParseToken(token string) (*jwt.Token, error) {

	claims := JWT{}

	parsed, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.env.JWTKey), nil
	})
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

func (j *ServingJWT) CreateToken(ctx context.Context, brandName, ip, role string) (string, error) {

	jti := uuid.NewString()

	customClaims := &JWT{
		Ip:        ip,
		BrandName: brandName,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    j.env.Issuer,
			Subject:   ip,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	newToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims).SignedString([]byte(j.env.JWTKey))
	if err != nil {
		j.log.Error("jwt.go :72 >> failed to create token: "+ip+" "+brandName, zap.Error(err))
		return "", err
	}

	if err = j.rdb.Set(ctx, jti, newToken, 12*time.Hour).Err(); err != nil {
		j.log.Error("jwt.go :77 >> failed to store token: "+ip+" "+brandName, zap.Error(err))
		return "", err
	}

	return newToken, nil
}

func (j *ServingJWT) CheckAdminAccess(token *jwt.Token, role string) bool {

	claims := token.Claims.(*JWT)

	if claims.Role != role {
		j.log.Error("jwt.go :89 >> invalid token: " + claims.ID)
		return false
	}

	if ok := j.IsExist(context.Background(), claims.RegisteredClaims.ID); !ok {
		j.log.Error("jwt.go :95 >> token not found: " + claims.RegisteredClaims.ID)
		return false
	}

	return true
}

func (j *ServingJWT) LogOut(ctx context.Context, jti string) error {

	if err := j.rdb.Del(ctx, jti).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			j.log.Error("jwt.go :105 >> token not found: " + jti)
			return entity.NotFound
		}

		j.log.Error("jwt.go :109 >> retry to delete token: " + jti)
		retryErr := j.RetryDeleteToken(ctx, jti)
		if retryErr != nil {
			j.log.Error("jwt.go :112 >> failed to enqueue retry: "+jti, zap.Error(retryErr))
		}
		if retryErr == nil {
			return nil
		}

		return err
	}

	return nil
}

func (j *ServingJWT) RetryDeleteToken(ctx context.Context, jti string) error {

	var lastErr error

	for i := 0; i < 3; i++ {

		err := j.rdb.Del(ctx, jti).Err()
		if err == nil {
			j.log.Info("jwt.go :132 >> token deleted: " + jti)
			return nil
		}
		if errors.Is(err, redis.Nil) {
			return nil
		}

		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("jwt.go :143 >> logout failed after retries: %v", lastErr)
}

func (j *ServingJWT) IsExist(ctx context.Context, jti string) bool {

	ok := j.rdb.Exists(ctx, jti).Val() == 1

	return ok

}
