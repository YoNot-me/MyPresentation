package authRepository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (r *AuthRepo) GetPass(ctx context.Context, name string) (string, error) {

	const query = `SELECT password FROM presentation.brands WHERE name = $1`

	var password string

	err := r.db.QueryRow(ctx, query, name).Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil

}

func (r *AuthRepo) BruteCount(ctx context.Context, ip string) (int, error) {

	query := fmt.Sprintf("brute:auth:%s", ip)

	cmd, err := r.rdb.Get(ctx, query).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return -1, err
	}

	if cmd == "" {
		return 0, nil
	}

	res, err := strconv.Atoi(cmd)
	if err != nil {
		return -1, err
	}

	return res, nil
}

func (r *AuthRepo) IncCount(ctx context.Context, ip string) error {

	query := fmt.Sprintf("brute:auth:%s", ip)

	val, err := r.rdb.Incr(ctx, query).Result()
	if err != nil {
		return err
	}

	if val == 1 {
		err = r.rdb.Expire(ctx, query, 3*time.Minute).Err()
		if err != nil {
			delErr := r.rdb.Del(ctx, query).Err()
			if delErr != nil {
				return delErr
			}

			return err
		}
	}

	return nil
}
