package adminRepository

import (
	"context"
	"errors"
	"fmt"
	"presentator/internal/core/entity"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

func (ar *AdminRepo) ListAllBrands(ctx context.Context) ([]entity.BrandsResponse, error) {

	const query = `
		SELECT name FROM presentation.brands
	`

	res, err := ar.db.Query(ctx, query)
	if err != nil {
		return []entity.BrandsResponse{}, err
	}
	defer res.Close()

	var respBrands []entity.BrandsResponse

	for res.Next() {
		brand := entity.BrandsResponse{}
		if err = res.Scan(&brand.Name); err != nil {
			return []entity.BrandsResponse{}, err
		}

		respBrands = append(respBrands, brand)
	}

	if res.Err() != nil {
		return []entity.BrandsResponse{}, res.Err()
	}

	return respBrands, nil
}

func (ar *AdminRepo) AddNewBrand(ctx context.Context, brandName, hashPass string) error {

	const query = `
		INSERT INTO presentation.brands (name, password) VALUES ($1, $2)
	`
	res, err := ar.db.Exec(ctx, query, brandName, hashPass)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return entity.AlreadyExist
		}

		return err
	}

	if res.RowsAffected() == 0 {
		return entity.DBError
	}

	return nil
}

func (ar *AdminRepo) ChangeBrandPassword(ctx context.Context, brandName, newPass string) error {

	const query = `
		UPDATE presentation.brands SET password = $1 WHERE name = $2
	`
	res, err := ar.db.Exec(ctx, query, newPass, brandName)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.NotFound
	}

	return nil
}

func (ar *AdminRepo) RenameBrand(ctx context.Context, brandName, newName string) error {

	const query = `
		UPDATE presentation.brands SET name = $1 WHERE name = $2
	`
	res, err := ar.db.Exec(ctx, query, newName, brandName)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.NotFound
	}

	return nil
}

func (ar *AdminRepo) DeleteBrand(ctx context.Context, brandName string) error {

	const query = `
		DELETE FROM presentation.brands WHERE name = $1
	`
	res, err := ar.db.Exec(ctx, query, brandName)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.NotFound
	}

	return nil
}

func (ar *AdminRepo) ListAllWorks(ctx context.Context, brandName string) ([]entity.WorksResponse, error) {

	const query = `
		SELECT workName, url, description FROM presentation.works WHERE brand = $1
	`

	res, err := ar.db.Query(ctx, query, brandName)
	if err != nil {
		return []entity.WorksResponse{}, err
	}
	defer res.Close()

	var respWorks []entity.WorksResponse

	for res.Next() {
		work := entity.WorksResponse{}
		if err = res.Scan(&work.WorkName, &work.Url, &work.Description); err != nil {
			return []entity.WorksResponse{}, err
		}

		respWorks = append(respWorks, work)
	}

	if res.Err() != nil {
		return []entity.WorksResponse{}, res.Err()
	}

	return respWorks, nil
}

func (ar *AdminRepo) AddNewWork(ctx context.Context, req *entity.Works) error {

	const query = `
		INSERT INTO presentation.works (brand, workName, url, description)
		VALUES ($1, $2, $3, $4)
	`

	res, err := ar.db.Exec(ctx, query,
		req.Brand,
		req.WorkName,
		req.Url,
		req.Description,
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return entity.AlreadyExist
		}
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.DBError
	}

	return nil
}

func (ar *AdminRepo) DeleteWork(ctx context.Context, brandName, workName string) error {

	const query = `
		DELETE FROM presentation.works WHERE brand = $1 AND workName = $2
	`

	res, err := ar.db.Exec(ctx, query, brandName, workName)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.NotFound
	}

	return nil
}

func (ar *AdminRepo) ChangeWorkFields(
	ctx context.Context,
	brandName, workName string,
	work *entity.Works) error {

	const query = `
		UPDATE presentation.works
		SET
			workName    = CASE WHEN $3 <> '' THEN $3 ELSE workName END,
			url         = CASE WHEN $4 <> '' THEN $4 ELSE url END,
			description = CASE WHEN $5 <> '' THEN $5 ELSE description END
		WHERE brand = $1
		  AND workName = $2
		  AND (
		      (workName    		IS DISTINCT FROM CASE WHEN $3 <> '' THEN $3 ELSE workName END)
		      OR (url         	IS DISTINCT FROM CASE WHEN $4 <> '' THEN $4 ELSE url END)
		      OR (description 	IS DISTINCT FROM CASE WHEN $5 <> '' THEN $5 ELSE description END)
		  )
	`

	res, err := ar.db.Exec(ctx, query,
		brandName,
		workName,
		work.WorkName,
		work.Url,
		work.Description,
	)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return entity.AlreadyExist
		}
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.FilesNotChanged
	}

	return nil
}

func (ar *AdminRepo) GetWork(ctx context.Context, brandName, workName string) (entity.Works, error) {

	const query = `
		SELECT brand, workName, url, description FROM presentation.works WHERE brand = $1 AND workName = $2
	`

	var work entity.Works

	if err := ar.db.QueryRow(ctx, query, brandName, workName).Scan(
		&work.Brand,
		&work.WorkName,
		&work.Url,
		&work.Description,
	); err != nil {
		return entity.Works{},
			err
	}

	return work, nil
}

func (ar *AdminRepo) IsWorkExist(ctx context.Context, brandName, workName string) (bool, error) {

	const query = "SELECT EXISTS(SELECT 1 FROM presentation.works WHERE brand = $1 AND workName = $2)"
	var ok bool

	if err := ar.db.QueryRow(ctx, query, brandName, workName).Scan(&ok); err != nil {
		return false, err
	}

	return ok, nil
}

func (ar *AdminRepo) BruteCount(ctx context.Context, ip string) (int, error) {

	query := fmt.Sprintf("brute:admin:auth:%s", ip)

	cmd, err := ar.rdb.Get(ctx, query).Result()
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

func (ar *AdminRepo) IncCount(ctx context.Context, ip string) error {

	query := fmt.Sprintf("brute:admin:auth:%s", ip)

	val, err := ar.rdb.Incr(ctx, query).Result()
	if err != nil {
		return err
	}

	if val == 1 {
		err = ar.rdb.Expire(ctx, query, 3*time.Minute).Err()
		if err != nil {
			delErr := ar.rdb.Del(ctx, query).Err()
			if delErr != nil {
				return delErr
			}

			return err
		}
	}

	return nil
}
