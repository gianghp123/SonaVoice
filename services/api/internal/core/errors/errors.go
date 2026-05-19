package errors

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, defaultMsg string, msg ...string) *AppError {
	message := defaultMsg
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	return &AppError{Code: code, Message: message}
}

func BadRequest(msg ...string) *AppError {
	return New(http.StatusBadRequest, "Bad request", msg...)
}

func Unauthorized(msg ...string) *AppError {
	return New(http.StatusUnauthorized, "Unauthorized", msg...)
}

func Forbidden(msg ...string) *AppError {
	return New(http.StatusForbidden, "Forbidden", msg...)
}

func NotFound(msg ...string) *AppError {
	return New(http.StatusNotFound, "Resource not found", msg...)
}

func Conflict(msg ...string) *AppError {
	return New(http.StatusConflict, "Conflict occurred", msg...)
}

func Internal(msg ...string) *AppError {
	return New(http.StatusInternalServerError, "Internal server error", msg...)
}

func AlreadyExists(msg ...string) *AppError {
	return New(http.StatusConflict, "Already exists", msg...)
}

// MapRepoError maps ORM/database errors to sentinel errors.
// If the error is already an AppError, it passes through directly.
func MapRepoError(err error) *AppError {
	var pgErr *pgconn.PgError
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound()
	}

	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Conflict()
	}

	return Internal()
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
