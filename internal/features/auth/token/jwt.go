package JWT

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func (j *ServingJWT) ValidateToken(token string) (*jwt.Token, bool) {

	claims := JWT{}

	parsed, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.env.JWTKey), nil
	})
	if err != nil {
		return nil, false
	}

	if !parsed.Valid {
		return nil, false
	}

	return parsed, true
}

func (j *ServingJWT) CreateToken(ctx context.Context, brandName, ip, role string) (string, error) {

	customClaims := &JWT{
		Ip:        ip,
		BrandName: brandName,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "YoNot",
			Subject:   ip,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	newToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims).SignedString([]byte(j.env.JWTKey))
	if err != nil {
		j.log.Error("failed to create token: "+ip+" "+brandName, zap.Error(err))
		return "", err
	}

	if err = j.rdb.Set(ctx, ip, newToken, 12*time.Hour).Err(); err != nil {
		j.log.Error("failed to store token: "+ip+" "+brandName, zap.Error(err))
		return "", err
	}

	return newToken, nil
}

func (j *ServingJWT) CheckAccess(token *jwt.Token, ip string) bool {

	claims := token.Claims.(*JWT)

	if claims.Ip != ip {
		j.log.Error("invalid token: " + ip)
		return false
	}

	return true
}

func (j *ServingJWT) CheckAdminAccess(token *jwt.Token, ip, role string) bool {

	claims := token.Claims.(*JWT)

	if claims.Ip != ip {
		j.log.Error("invalid token: " + ip)
		return false
	}
	if claims.Role != role {
		j.log.Error("invalid token: " + ip)
		return false
	}

	return true
}

func (j *ServingJWT) LogOut(ctx context.Context, ip string) error {

	if err := j.rdb.Del(ctx, ip).Err(); err != nil {

		go func() {
			j.log.Error("retry to delete token: " + ip)
			retryErr := j.RetryDeleteToken(ip)
			if retryErr != nil {
				j.log.Error("failed to enqueue retry: "+ip, zap.Error(retryErr))
			}
		}()
	}

	return nil
}

func (j *ServingJWT) RetryDeleteToken(ip string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var lastErr error

	for i := 0; i < 3; i++ {

		err := j.rdb.Del(ctx, ip).Err()
		if err == nil {
			j.log.Info("token deleted: " + ip)
			return nil
		}

		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("logout failed after retries: %v", lastErr)
}

func (j *ServingJWT) IsExist(ctx context.Context, ip string) bool {

	ok := j.rdb.Exists(ctx, ip).Val() == 1

	return ok

}
