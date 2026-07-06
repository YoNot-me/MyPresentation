package adminRepository

import (
	"context"
	"errors"
	"presentator/internal/core/entity"

	"github.com/jackc/pgx/v5/pgconn"
)

func (ar *AdminRepo) ListAllBrands(ctx context.Context) ([]entity.Brand, error) {

	const query = "SELECT name FROM brands"

	res, err := ar.db.Query(ctx, query)
	if err != nil {
		return []entity.Brand{}, err
	}
	defer res.Close()

	var respBrands []entity.Brand

	for res.Next() {
		brand := entity.Brand{}
		if err = res.Scan(&brand.Name); err != nil {
			return []entity.Brand{}, err
		}

		respBrands = append(respBrands, brand)
	}

	if res.Err() != nil {
		return []entity.Brand{}, res.Err()
	}

	return respBrands, nil
}

func (ar *AdminRepo) AddNewBrand(ctx context.Context, brandName, hashPass string) error {

	const query = "INSERT INTO brands (name, password) VALUES ($1, $2)"
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

	const query = "UPDATE brands SET password = $1 WHERE name = $2"
	res, err := ar.db.Exec(ctx, query, newPass, brandName)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.NotFound
	}

	return nil
}

func (ar *AdminRepo) DeleteBrand(ctx context.Context, brandName string) error {

	const query = "DELETE FROM brands WHERE name = $1"
	res, err := ar.db.Exec(ctx, query, brandName)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.NotFound
	}

	return nil
}

func (ar *AdminRepo) ListAllWorks(ctx context.Context, brandName string) ([]entity.Works, error) {

	const query = "SELECT brand, workName, url FROM works WHERE brand = $1"

	res, err := ar.db.Query(ctx, query, brandName)
	if err != nil {
		return []entity.Works{}, err
	}
	defer res.Close()

	var respWorks []entity.Works

	for res.Next() {
		work := entity.Works{}
		if err = res.Scan(&work.Brand, &work.WorkName, &work.Url); err != nil {
			return []entity.Works{}, err
		}

		respWorks = append(respWorks, work)
	}

	if res.Err() != nil {
		return []entity.Works{}, res.Err()
	}

	return respWorks, nil
}

func (ar *AdminRepo) AddNewWork(ctx context.Context, brandName, workName, url string) error {

	switch url {
	case "":
		const query = "INSERT INTO works (brand, workName) VALUES ($1, $2)"

		res, err := ar.db.Exec(ctx, query, brandName, workName)
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

	default:
		const query = "INSERT INTO works (brand, workName, url) VALUES ($1, $2, $3)"
		res, err := ar.db.Exec(ctx, query, brandName, workName, url)
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
}

func (ar *AdminRepo) DeleteWork(ctx context.Context, brandName, workName string) error {

	const query = "DELETE FROM works WHERE brand = $1 AND workName = $2"
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
		 UPDATE works
        SET
            workName = CASE WHEN $3 <> '' THEN $3 ELSE workName END,
            url      = CASE WHEN $4 <> '' THEN $4 ELSE url END
        WHERE brand = $1 
          AND workName = $2
          AND (
              (workName IS DISTINCT FROM CASE WHEN $3 <> '' THEN $3 ELSE workName END)
              OR
              (url IS DISTINCT FROM CASE WHEN $4 <> '' THEN $4 ELSE url END)
          )
	`

	res, err := ar.db.Exec(ctx, query, brandName, workName, work.WorkName, work.Url)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return entity.AlreadyExist
		}
		return err
	}

	if res.RowsAffected() == 0 {
		return entity.BadRequest
	}

	return nil
}

func (ar *AdminRepo) GetWork(ctx context.Context, workName string) (entity.Works, error) {

	const query = "SELECT id, brand, workName, url FROM works WHERE workName = $1"

	var work entity.Works

	if err := ar.db.QueryRow(ctx, query, workName).Scan(
		&work.Id,
		&work.Brand,
		&work.WorkName,
		&work.Url,
	); err != nil {
		return entity.Works{},
			err
	}

	return work, nil
}
