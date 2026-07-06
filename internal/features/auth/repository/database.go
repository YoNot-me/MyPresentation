package authRepository

import "context"

func (r *AuthRepo) GetPass(ctx context.Context, name string) (string, error) {

	const query = `SELECT password FROM brands WHERE name = $1`

	var password string

	err := r.db.QueryRow(ctx, query, name).Scan(&password)
	if err != nil {
		return "", err
	}

	return password, nil

}
