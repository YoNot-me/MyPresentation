package fileservingRepository

import (
	"context"
	"presentator/internal/core/entity"
)

func (r *ServingRepo) GetAllWorks(ctx context.Context, brandName string) ([]entity.Works, error) {

	const query = "SELECT workName, url, description FROM works WHERE brand = $1"

	res, err := r.db.Query(ctx, query, brandName)
	if err != nil {
		return []entity.Works{}, err
	}
	defer res.Close()

	var respWorks []entity.Works

	for res.Next() {
		work := entity.Works{}
		if err = res.Scan(&work.WorkName, &work.Url, &work.Description); err != nil {
			return []entity.Works{}, err
		}

		respWorks = append(respWorks, work)
	}

	if res.Err() != nil {
		return []entity.Works{}, res.Err()
	}

	return respWorks, nil
}
