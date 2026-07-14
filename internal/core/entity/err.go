package entity

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

var (
	NotFound         = errors.New("work folder not found")
	DBError          = errors.New("db error")
	EnvNotLoaded     = errors.New("env not loaded")
	InvalidToken     = errors.New("invalid token")
	InvalidLogin     = errors.New("invalid login")
	InvalidPass      = errors.New("invalid password")
	InternalError    = errors.New("internal error")
	BadRequest       = errors.New("bad request")
	AlreadyExist     = errors.New("already exist")
	AlreadySigned    = errors.New("already signed")
	TooManyAttempts  = errors.New("too many attempts")
	FilesNotChanged  = errors.New("files not changed")
	ErrPathTraversal = errors.New("path traversal")
)

func FindStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrPathTraversal):
		return http.StatusForbidden
	case errors.Is(err, FilesNotChanged):
		return http.StatusOK
	case errors.Is(err, TooManyAttempts):
		return http.StatusTooManyRequests
	case errors.Is(err, AlreadySigned):
		return http.StatusConflict
	case errors.Is(err, AlreadyExist):
		return http.StatusBadRequest
	case errors.Is(err, BadRequest):
		return http.StatusBadRequest
	case errors.Is(err, InvalidToken):
		return http.StatusUnauthorized
	case errors.Is(err, InvalidLogin):
		return http.StatusUnauthorized
	case errors.Is(err, InvalidPass):
		return http.StatusUnauthorized
	case errors.Is(err, EnvNotLoaded):
		return http.StatusInternalServerError
	case errors.Is(err, pgx.ErrNoRows):
		return http.StatusNotFound
	case errors.Is(err, DBError):
		return http.StatusInternalServerError
	case errors.Is(err, NotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
