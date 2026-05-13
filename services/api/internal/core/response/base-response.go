package response

import "github.com/gianghp123/SonaVoice/api/internal/core/errors"

type BaseResponse[T any] struct {
	Success bool             `json:"success"`
	Data    T                `json:"data,omitempty"`
	Error   *errors.AppError `json:"error,omitempty"`
	Meta    *Meta            `json:"meta,omitempty"`
}

func Success[T any](data T) BaseResponse[T] {
	return BaseResponse[T]{Success: true, Data: data}
}

func SuccessWithMeta[T any](data T, meta *Meta) BaseResponse[T] {
	return BaseResponse[T]{Success: true, Data: data, Meta: meta}
}

func Fail(err *errors.AppError) BaseResponse[any] {
	return BaseResponse[any]{Success: false, Error: err}
}
